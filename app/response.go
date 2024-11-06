package main

import (
	"fmt"
	"strings"
)

type HttpResponse struct {
	code    int
	message string
	headers map[string]string
	data    string
}

func (res *HttpResponse) ToString() string {
	var strBuilder strings.Builder
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", res.code, res.message)
	strBuilder.WriteString(statusLine)

	for key, val := range res.headers {
		header := fmt.Sprintf("%s: %s\r\n", key, val)
		strBuilder.WriteString(header)
	}

	strBuilder.WriteString("\r\n")
	strBuilder.WriteString(res.data)
	strBuilder.WriteString("\r\n")

	return strBuilder.String()
}