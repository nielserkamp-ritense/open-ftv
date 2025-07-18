package bundles

import "strings"

// CompressionType represents a type of compression.
type CompressionType uint8

// List of supported compression types.
const (
	CompressGZ CompressionType = iota + 1
	CompressBZ2
)

// String implements the Stringer interface.
func (c CompressionType) String() string {
	switch c {
	case CompressGZ:
		return "GZip"
	case CompressBZ2:
		return "BZip2"
	default:
		return "???"
	}
}

// CompressionTypeFromString converts the given string to a compression type.
//
// Unrecognized values default to GZip.
func CompressionTypeFromString(s string) CompressionType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "bz2", "bzip", "bzip2":
		return CompressBZ2
	default:
		return CompressGZ
	}
}
