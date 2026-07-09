package cloudevents

import "encoding/base64"

// decodeBase64 decodes a data_base64 value; it returns nil when the input is invalid.
func decodeBase64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil
	}
	return b
}
