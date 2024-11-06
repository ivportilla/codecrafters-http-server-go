package main

import (
	"fmt"
	"strings"
)

type RequestLine struct {
	method        string
	requestTarget string
	httpVersion   string
}

type HttpRequest struct {
	requestLine RequestLine
	headers     map[string]string
	body        string
}

func sanitizeBody(target string) string {
	result := ""
	for _, char := range target {
		if char == 0x00 {
			return result
		}
		result += string(char)
	}
	return result
}

func extractBody(requestParts []string) string {
	bodyIdx := func() int {
		for i, line := range requestParts {
			if line == "" {
				return i
			}
		}
		return -1
	}()

	if bodyIdx >= len(requestParts) {
		return ""
	}

	return sanitizeBody(requestParts[bodyIdx+1])
}

func extractHeaders(requestParts []string) map[string]string {
	headers := make(map[string]string)

	for _, line := range requestParts[1:] {
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

func parseRequestLine(target string) (RequestLine, error) {
	lineParts := strings.Split(target, " ")

	if len(lineParts) != 3 {
		return RequestLine{}, fmt.Errorf("the request line could not be parsed correctly")
	}

	return RequestLine{
		method:        lineParts[0],
		requestTarget: lineParts[1],
		httpVersion:   lineParts[2],
	}, nil
}

func parseHttpRequest(data string) (HttpRequest, error) {
	requestParts := strings.Split(data, "\r\n")
	if len(requestParts) < 3 {
		return HttpRequest{}, fmt.Errorf("invalid request format")
	}

	requestLine, err := parseRequestLine(requestParts[0])
	if err != nil {
		return HttpRequest{}, fmt.Errorf("error parsing request line")
	}

	headers := extractHeaders(requestParts)

	body := extractBody(requestParts)

	return HttpRequest{
		requestLine: requestLine,
		headers:     headers,
		body:        body,
	}, nil
}