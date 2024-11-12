package main

import (
	"bytes"
	"fmt"
)

type HttpResponse struct {
	code    int
	message string
	headers map[string]string
	data    []byte
}

func (res *HttpResponse) ToBytes() []byte {
	var builder bytes.Buffer
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", res.code, res.message)
	builder.WriteString(statusLine)

	for key, val := range res.headers {
		header := fmt.Sprintf("%s: %s\r\n", key, val)
		builder.WriteString(header)
	}

	builder.WriteString("\r\n")
	builder.Write(res.data)
	builder.WriteString("\r\n")

	return builder.Bytes()
}
