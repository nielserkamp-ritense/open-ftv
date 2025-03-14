package pem

import (
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

// LoadFile returns the first PEM block from the given file, or an error if it fails.
func LoadFile(path string) (*pem.Block, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer r.Close()
	return LoadStream(r)
}

// LoadStream returns the first PEM block from the given stream, or an error if it fails.
func LoadStream(r io.Reader) (*pem.Block, error) {
	d, err2 := io.ReadAll(r)
	if err2 != nil {
		return nil, fmt.Errorf("failed to read file: %w", err2)
	}

	b, _ := pem.Decode(d)
	return b, nil
}
