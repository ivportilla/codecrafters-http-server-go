package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"regexp"
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

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	
	conn, errConn := l.Accept()
	if errConn != nil {
		fmt.Println("Error accepting connection: ", errConn.Error())
		os.Exit(1)
	}

	defer conn.Close()

	buffer := make([]byte, 256)
	_, err = conn.Read(buffer)

	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
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
	
	if echoRexp.MatchString(requestLine.requestTarget) {
		handleEco(conn, requestLine)
	} else {
		handleDefault(conn, requestLine)
	}
}
