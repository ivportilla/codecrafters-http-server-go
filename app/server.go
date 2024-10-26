package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

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

	okResponse := "HTTP/1.1 200 OK\r\n\r\n"
	errorResponse := "HTTP/1.1 404 Not Found\r\n\r\n"

	buffer := make([]byte, 256)
	_, err = conn.Read(buffer)

	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
	}

	httpMessage := strings.Split(string(buffer), "\r\n")

	if len(httpMessage) == 0 {
		fmt.Println("Error parsing HTTP Message")
		conn.Close()
	}

	requestLine := httpMessage[0]
	requestPath := strings.Split(requestLine, " ")[1]

	if requestPath != "/" {
		conn.Write([]byte(errorResponse))
	} else {
		conn.Write([]byte(okResponse))
	}

	conn.Close()	
}
