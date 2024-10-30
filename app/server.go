package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type requestLine struct {
	method string
	requestTarget string
	httpVersion string
}

func parseRequestLine(target string) (requestLine, error) {
	lineParts := strings.Split(target, " ")

	if len(lineParts) != 3 {
		return requestLine{}, fmt.Errorf("the request line could not be parsed correctly")
	}

	return requestLine{
		method: lineParts[0],
		requestTarget: lineParts[1],
		httpVersion: lineParts[2],
	}, nil

}

func extractHeaders(httpMessage []string) map[string]string {
	headers := make(map[string]string)

	for _, line := range httpMessage[1:] {
		if line == "" {
			break
		}
		currentHeader := strings.Split(line, ":")
		key := strings.TrimSpace(currentHeader[0])
		val := strings.TrimSpace(currentHeader[1])
		headers[key] = val
	}

	return headers

}

func handleEco(conn net.Conn, reqLine requestLine) {
	echoRexp := regexp.MustCompile("/echo/(?P<data>.+)")
	matches := echoRexp.FindStringSubmatch(reqLine.requestTarget)
	data := matches[echoRexp.SubexpIndex("data")]
	okResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(data), data)
	conn.Write([]byte(okResponse))
}

func handleDefault(conn net.Conn, reqLine requestLine) {
	okResponse := "HTTP/1.1 200 OK\r\n\r\n"
	errResponse := "HTTP/1.1 404 Not Found\r\n\r\n"
	if reqLine.requestTarget == "/" {
		conn.Write([]byte(okResponse))
	} else {
		conn.Write([]byte(errResponse))
	}
}

func handleUserAgent(conn net.Conn, headers map[string]string) {
	errResponse := "HTTP/1.1 400 Bad Request\r\n\r\n"
	userAgent, ok := headers["User-Agent"]
	fmt.Println("ok", ok, userAgent)
	if !ok {
		conn.Write([]byte(errResponse))
		return
	}
	okResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
	conn.Write([]byte(okResponse))
}

func reqHandler(conn net.Conn, ch chan net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)

	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
		return
	}

	httpMessage := strings.Split(string(buffer), "\r\n")

	if len(httpMessage) == 0 {
		fmt.Println("Error parsing HTTP Message")
		return
	}

	requestLine, err := parseRequestLine(httpMessage[0])

	if err != nil {
		fmt.Println("Error parsing request line: ", err.Error())
	}

	echoRexp := regexp.MustCompile("/echo/(?P<data>.+)")
	userAgentRexp := regexp.MustCompile("/user-agent")
	
	if echoRexp.MatchString(requestLine.requestTarget) {
		handleEco(conn, requestLine)
	} else if userAgentRexp.MatchString(requestLine.requestTarget) {
		headers := extractHeaders(httpMessage)
		handleUserAgent(conn, headers)
	} else {
		handleDefault(conn, requestLine)
	}

	defer func() {
		ch <- conn
	}()
}

const (
	POOL_SIZE = 5
	SLEEP_TIMEOUT = 200 * time.Millisecond
)

type ConnectionPool struct {
	connections []net.Conn
	mut sync.Mutex
	maxSize int
}

func NewConnectionPool(maxSize int) *ConnectionPool {
	return &ConnectionPool{
		connections: make([]net.Conn, 0, maxSize),
		maxSize: maxSize,
	}
}

func (cp * ConnectionPool) Add(conn net.Conn) bool {
	cp.mut.Lock()
	defer cp.mut.Unlock()

	if len(cp.connections) >= cp.maxSize {
		return false
	}

	cp.connections = append(cp.connections, conn)
	return true
}

func (cp * ConnectionPool) Remove(connToClose net.Conn) bool {
	cp.mut.Lock()
	defer cp.mut.Unlock()

	idx := -1
	for i, c := range cp.connections {
		if c == connToClose {
			idx = i
			break;
		}
	}

	if idx > -1 {
		cp.connections = append(cp.connections[:idx], cp.connections[idx+1:]...)
		fmt.Println("Connection deleted, new pool size: ", len(cp.connections))
		return true
	}

	return false
}

func (cp * ConnectionPool) HandlePool(closeCh chan net.Conn) {
	for {
		select {
		case toClose := <-closeCh:
			toClose.Close();
			cp.Remove(toClose)
		case <-time.After(SLEEP_TIMEOUT):
		}
	}
}

func handleConnections(listener net.Listener, cp * ConnectionPool, closeCh chan net.Conn) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
		}

		connAdded := cp.Add(conn)
		if !connAdded {
			fmt.Println("Connection pool is full at the moment, rejecting connection")
			conn.Close()
			continue
		}
		go reqHandler(conn, closeCh)
		time.Sleep(SLEEP_TIMEOUT)
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	pool := NewConnectionPool(POOL_SIZE)
	closeCh := make(chan net.Conn)

	// Handle removal to delete connections from the pool
	go pool.HandlePool(closeCh)

	// Wait for new connections
	handleConnections(l, pool, closeCh)
}
