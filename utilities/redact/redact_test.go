package redact

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "connection refused with IPv4",
			input: `Get "http://10.0.0.5:8080/x": dial tcp 10.0.0.5:8080: connect: connection refused`,
			want:  `Get "*****": dial tcp *****: connect: connection refused`,
		},
		{
			name:  "timeout with bracketed IPv6",
			input: `dial tcp [::1]:6693: connect: connection refused`,
			want:  "dial tcp *****: connect: connection refused",
		},
		{
			name:  "DNS lookup failure with bare hostname",
			input: `Get "http://internal-secrets.corp/x": dial tcp: lookup internal-secrets.corp: no such host`,
			want:  `Get "*****": dial tcp: lookup *****: no such host`,
		},
		{
			name:  "TLS hostname mismatch",
			input: `x509: certificate is valid for a.example.com, not b.example.com`,
			want:  "x509: certificate is valid for *****, not *****",
		},
		{
			name:  "field validation message is untouched",
			input: "title too long (max 80 characters)",
			want:  "title too long (max 80 characters)",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, Sensitive(tc.input))
		})
	}
}
