package io

import (
	"os"
	"path/filepath"
)

// SnifFile attempts to determine the mime-type by examining the given file path.
//
// If the path contains a distinguishable extension that can be converted to a supported mime-type,
// that mime-type will be returned.
// In other cases, if the file can be opened, the mime-type will be determined based on the first few bytes
// of the file using the stream sniffer function.
// If the file cannot be opened, DefaultMimeType is returned.
func SnifFile(path string) string {
	ext := filepath.Ext(path)

	if t := ConvertExt(ext); IsSupported(t) {
		return t
	}

	if f, err := os.Open(path); err == nil {
		defer f.Close()
		return SnifStream(f)
	}

	return DefaultMimeType
}
