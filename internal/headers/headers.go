package headers

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

type Headers map[string][]string

func NewHeaders() Headers {
	return map[string][]string{}
}

var sep = []byte("\r\n")

var (
	ErrorNoFieldName      = fmt.Errorf("no field name: malformed request")
	ErrorInvalidFieldName = fmt.Errorf("invalid field name")
	ErrorInvalidCharacter = fmt.Errorf("invalid character")
)

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	read := 0

	for {
		idx := bytes.Index(data[read:], sep)
		if idx == -1 {
			return read, false, nil
		}

		if idx == 0 {
			read += len(sep)
			return read, true, nil
		}

		line := data[read : read+idx]

		name, value, parseErr := parseHeader(line)
		if parseErr != nil {
			return read, false, parseErr
		}

		val, ok := h[strings.ToLower(name)]

		if !ok {
			h[strings.ToLower(name)] = []string{value}
		} else {
			h[strings.ToLower(name)] = append(val, value)
		}

		read += idx + len(sep)
	}
}

func isValidFieldName(name []byte) bool {
	for _, b := range name {
		// RFC 7230 token chars: ! # $ % & ' * + - . 0-9 A-Z ^ _ ` a-z | ~
		if (b < 'a' || b > 'z') && (b < 'A' || b > 'Z') && (b < '0' || b > '9') && b != '-' && b != '_' &&
			b != '.' && b != '!' && b != '#' && b != '$' &&
			b != '%' && b != '&' && b != '\'' && b != '*' &&
			b != '+' && b != '^' && b != '`' && b != '|' && b != '~' {
			return false
		}
	}
	return true
}

func parseHeader(line []byte) (string, string, error) {
	parts := bytes.SplitN(line, []byte(":"), 2)

	if len(parts) != 2 {
		log.Printf("hey")
		log.Println(len(parts))
		for i := range parts {
			log.Printf("part %v = %s", i, string(parts[i]))
		}
		return "", "", ErrorNoFieldName
	}

	fieldName := parts[0]
	fieldValue := parts[1]

	fieldName = bytes.TrimSpace(fieldName)
	fieldValue = bytes.TrimSpace(fieldValue)

	if len(fieldName) == 0 {
		return "", "", ErrorNoFieldName
	}

	if !isValidFieldName(fieldName) {
		return "", "", ErrorInvalidFieldName
	}
	return string(fieldName), string(fieldValue), nil
}
