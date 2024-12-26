package io

import (
	"bytes"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// SnifStream attempts to determine the mime-type by examining the first few bytes of the given data-stream.
//
// It is kept as simple as possible, so it can potentially flag false positives.
// Use with care and a grain of salt!
func SnifStream(f io.Reader) string {
	b := make([]byte, 1024)

	l, err := f.Read(b)
	if err != nil && err != io.EOF {
		return DefaultMimeType
	}

	b = b[:l]

	for i := range order {
		if t := order[i]; sniffers[t](b) {
			return t
		}
	}

	return DefaultMimeType
}

type sniffer func([]byte) bool

var (
	sniffers = map[string]sniffer{
		MimeTypeOPA:     isRego,
		MimeTypeCedar:   isCedar,
		MimeTypeOpenFGA: isOpenFGA,
		MimeTypeTOML:    isTOML,
		MimeTypeYAML:    isYAML,
		MimeTypeXML:     isXML,
		MimeTypeJSON:    isJSON,
	}

	order = []string{
		MimeTypeOPA,
		MimeTypeCedar,
		MimeTypeOpenFGA,
		MimeTypeTOML,
		MimeTypeYAML,
		MimeTypeXML,
		MimeTypeJSON,
	}
)

func isRego(b []byte) bool {
	lines := strings.Split(string(b), "\n")
	if len(lines) < 2 {
		return false
	}

	for len(lines) > 0 {
		l := strings.TrimSpace(lines[0])
		if l == "" || strings.HasPrefix(l, "#") {
			lines = lines[1:]
		} else {
			break
		}
	}

	if len(lines) == 0 || !strings.HasPrefix(lines[0], "package") {
		return false
	}

	return bytes.Contains(b, []byte("import rego")) ||
		bytes.Contains(b, []byte("allow if {"))
}

func isCedar(b []byte) bool {
	if !bytes.HasPrefix(b, []byte("permit")) {
		return false
	}

	return bytes.Contains(b, []byte("principal")) &&
		bytes.Contains(b, []byte("action")) &&
		bytes.Contains(b, []byte("resource"))
}

func isOpenFGA(b []byte) bool {
	if !bytes.HasPrefix(b, []byte("model")) {
		return false
	}

	return bytes.Contains(b, []byte("schema"))
}

func isXML(b []byte) bool {
	return bytes.HasPrefix(b, []byte("<?xml"))
}

func isTOML(b []byte) bool {
	lines := strings.Split(string(b), "\n")
	if len(lines) < 2 {
		return false
	}

	lines = lines[:len(lines)-1]
	for i := range lines {
		l := strings.TrimSpace(lines[i])

		switch {
		case l == "",
			strings.HasPrefix(l, "#"),
			strings.HasPrefix(l, "[[") && strings.HasSuffix(l, "]]"),
			strings.HasPrefix(l, "[") && strings.HasSuffix(l, "]"),
			strings.Contains(l, "="):
		default:
			return false
		}
	}

	return true
}

func isYAML(b []byte) bool {
	lines := strings.Split(string(b), "\n")
	if len(lines) < 2 {
		return false
	}

	if l := lines[0]; !strings.HasPrefix(l, "#") && !strings.HasPrefix(l, "- ") && !strings.Contains(l, ":") {
		return false
	}

	lines = lines[:len(lines)-1]

	var m any
	err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &m)
	return err == nil
}

func isJSON(b []byte) bool {
	return bytes.HasPrefix(b, []byte("{")) || bytes.HasPrefix(b, []byte("["))
}
