package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func handleEco(conn net.Conn, req HttpRequest) {
	echoRexp := regexp.MustCompile("/echo/(?P<data>.+)")
	matches := echoRexp.FindStringSubmatch(req.requestLine.requestTarget)
	data := matches[echoRexp.SubexpIndex("data")]
	response := HttpResponse{
		code:    200,
		message: "OK",
		headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(data)),
		},
		data: data,
	}
	if req.headers["Accept-Encoding"] == "gzip" {
		response.headers["Content-Encoding"] = "gzip"
	}
	conn.Write([]byte(response.ToString()))
}

func handleDefault(conn net.Conn, req HttpRequest) {
	okResponse := HttpResponse{code: http.StatusOK, message: "OK"}
	errResponse := HttpResponse{code: http.StatusNotFound, message: "Not Found"}
	if req.requestLine.requestTarget == "/" {
		conn.Write([]byte(okResponse.ToString()))
	} else {
		conn.Write([]byte(errResponse.ToString()))
	}
}

func handleUserAgent(conn net.Conn, req HttpRequest) {
	errResponse := HttpResponse{code: http.StatusBadRequest, message: "Bad Request"}
	userAgent, ok := req.headers["User-Agent"]
	if !ok {
		conn.Write([]byte(errResponse.ToString()))
		return
	}
	okRes := HttpResponse{
		code:    http.StatusOK,
		message: "OK",
		headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(userAgent)),
		},
		data: userAgent,
	}
	conn.Write([]byte(okRes.ToString()))
}

func handleReadFile(conn net.Conn, req HttpRequest) {
	errResponse := HttpResponse{code: http.StatusNotFound, message: "Not Found"}
	fileName := strings.Split(req.requestLine.requestTarget, "/")[2]
	data, err := os.ReadFile(os.Args[2] + fileName)
	if err != nil {
		fmt.Printf("Error reading file %s: %v", fileName, err)
		conn.Write([]byte(errResponse.ToString()))
		return
	}

	okRes := HttpResponse{
		code:    http.StatusOK,
		message: "OK",
		headers: map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": fmt.Sprintf("%d", len(data)),
		},
		data: string(data),
	}

	conn.Write([]byte(okRes.ToString()))
}

func handleWriteFile(conn net.Conn, req HttpRequest) {
	errResponse := HttpResponse{code: http.StatusNotFound}
	fileName := strings.Split(req.requestLine.requestTarget, "/")[2]
	err := os.WriteFile(os.Args[2]+fileName, []byte(req.body), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %s", fileName, err.Error())
		conn.Write([]byte(errResponse.ToString()))
		return
	}

	okResponse := HttpResponse{
		code:    http.StatusCreated,
		message: "Created",
	}
	conn.Write([]byte(okResponse.ToString()))
}

func reqHandler(conn net.Conn, ch chan net.Conn) {
	defer func() {
		ch <- conn
	}()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)

	if err != nil {
		if err == io.EOF {
			return
		}
		fmt.Printf("Error reading connection: %v\n", err)
		return
	}

	req, err := parseHttpRequest(string(buffer))
	if err != nil {
		fmt.Printf("Error parsing http request: %v\n", err)
		return
	}

	fmt.Printf("Request: %v\n", req)

	httpMessage := strings.Split(string(buffer), "\r\n")

	if len(httpMessage) == 0 {
		fmt.Println("Error parsing HTTP Message")
		return
	}

	requestLine, err := parseRequestLine(httpMessage[0])
	if err != nil {
		fmt.Println("Error parsing request line: ", err.Error())
		return
	}

	echoRexp := regexp.MustCompile("/echo/(?P<data>.+)")
	userAgentRexp := regexp.MustCompile("/user-agent")

	if echoRexp.MatchString(requestLine.requestTarget) {
		handleEco(conn, req)
	} else if userAgentRexp.MatchString(requestLine.requestTarget) {
		handleUserAgent(conn, req)
	} else if regexp.MustCompile("/files/.+").MatchString(requestLine.requestTarget) && requestLine.method == "GET" {
		handleReadFile(conn, req)
	} else if regexp.MustCompile("/files/.+").MatchString(requestLine.requestTarget) && requestLine.method == "POST" {
		handleWriteFile(conn, req)
	} else {
		handleDefault(conn, req)
	}
}

func main() {
	fmt.Println("Server listening at port 4221")
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
