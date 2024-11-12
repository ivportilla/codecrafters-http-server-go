package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strings"
)

func handleEncoding(header map[string]string, allowedEncodings map[string]any) string {
	if len(header["Accept-Encoding"]) == 0 {
		return ""
	}

	encodings := strings.Split(header["Accept-Encoding"], ",")
	for _, encoding := range encodings {
		encoding = strings.TrimSpace(encoding)
		if _, ok := allowedEncodings[encoding]; ok {
			return encoding
		}
	}

	return ""
}

func compressGzip(data []byte) ([]byte, error) {
	var buffer bytes.Buffer

	writer := gzip.NewWriter(&buffer)
	_, err := writer.Write(data)
	if err != nil {
		return nil, fmt.Errorf("gzip writer failed: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("gzip writer failed flushing data: %v", err)
	}
	return buffer.Bytes(), nil
}
