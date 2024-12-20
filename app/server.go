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

var allowedEncodings = map[string]any{"gzip": compressGzip}

func handleEco(conn net.Conn, req HttpRequest) {
	echoRegExp := regexp.MustCompile("/echo/(?P<data>.+)")
	matches := echoRegExp.FindStringSubmatch(req.requestLine.requestTarget)
	data := matches[echoRegExp.SubexpIndex("data")]
	response := HttpResponse{
		code:    200,
		message: "OK",
		headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(data)),
		},
		data: []byte(data),
	}
	encoding := handleEncoding(req.headers, allowedEncodings)
	if encoding == "" {
		conn.Write(response.ToBytes())
		return
	}

	response.headers["Content-Encoding"] = encoding
	compressFn := allowedEncodings[encoding].(func([]byte) ([]byte, error))
	compressedData, err := compressFn(response.data)
	if err != nil {
		fmt.Printf("Error compressing data: %v\n", err)
		errResponse := HttpResponse{code: http.StatusInternalServerError, message: "Error compressing data"}
		conn.Write(errResponse.ToBytes())
		return
	}

	response.headers["Content-Length"] = fmt.Sprintf("%d", len(compressedData))
	response.data = compressedData
	conn.Write(response.ToBytes())
}

func handleDefault(conn net.Conn, req HttpRequest) {
	okResponse := HttpResponse{code: http.StatusOK, message: "OK"}
	errResponse := HttpResponse{code: http.StatusNotFound, message: "Not Found"}
	if req.requestLine.requestTarget == "/" {
		conn.Write(okResponse.ToBytes())
	} else {
		conn.Write(errResponse.ToBytes())
	}
}

func handleUserAgent(conn net.Conn, req HttpRequest) {
	errResponse := HttpResponse{code: http.StatusBadRequest, message: "Bad Request"}
	userAgent, ok := req.headers["User-Agent"]
	if !ok {
		conn.Write(errResponse.ToBytes())
		return
	}
	okRes := HttpResponse{
		code:    http.StatusOK,
		message: "OK",
		headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(userAgent)),
		},
		data: []byte(userAgent),
	}
	conn.Write(okRes.ToBytes())
}

func handleReadFile(conn net.Conn, req HttpRequest) {
	errResponse := HttpResponse{code: http.StatusNotFound, message: "Not Found"}
	fileName := strings.Split(req.requestLine.requestTarget, "/")[2]
	data, err := os.ReadFile(os.Args[2] + fileName)
	if err != nil {
		fmt.Printf("Error reading file %s: %v", fileName, err)
		conn.Write(errResponse.ToBytes())
		return
	}

	okRes := HttpResponse{
		code:    http.StatusOK,
		message: "OK",
		headers: map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": fmt.Sprintf("%d", len(data)),
		},
		data: data,
	}

	conn.Write(okRes.ToBytes())
}

func handleWriteFile(conn net.Conn, req HttpRequest) {
	errResponse := HttpResponse{code: http.StatusNotFound}
	fileName := strings.Split(req.requestLine.requestTarget, "/")[2]
	err := os.WriteFile(os.Args[2]+fileName, []byte(req.body), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %s", fileName, err.Error())
		conn.Write(errResponse.ToBytes())
		return
	}

	okResponse := HttpResponse{
		code:    http.StatusCreated,
		message: "Created",
	}
	conn.Write(okResponse.ToBytes())
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
