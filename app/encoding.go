package main

import "strings"

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
