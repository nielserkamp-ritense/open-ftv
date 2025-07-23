package bundles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressionType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		in       CompressionType
		wantName string
	}{
		{name: "zero", in: 0, wantName: "???"},
		{name: "unknown", in: 99, wantName: "???"},
		{name: "gz", in: CompressGZ, wantName: "GZip"},
		{name: "bzip", in: CompressBZ2, wantName: "BZip2"},
		{name: "min", in: CompressMIN, wantName: "GZip"},
		{name: "max", in: CompressMAX, wantName: "BZip2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.String()
			assert.Equal(t, tc.wantName, got)
		})
	}
}

func TestCompressionTypeFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		in       string
		want     CompressionType
		wantName string
	}{
		{name: "empty", want: CompressGZ, wantName: "GZip"},
		{name: "unknown", in: "q", want: CompressGZ, wantName: "GZip"},
		{name: "gz", in: "gz", want: CompressGZ, wantName: "GZip"},
		{name: "bzip", in: "bzip", want: CompressBZ2, wantName: "BZip2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := CompressionTypeFromString(tc.in)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantName, got.String())
		})
	}
}
