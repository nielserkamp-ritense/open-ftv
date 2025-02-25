package models

import (
	"net/url"
	"time"
)

// HTTPRequest contains the details of an HTTP request.
type HTTPRequest struct {
	RequestTime *time.Time          `json:"requestTime,omitempty"`
	URL         *url.URL            `json:"url,omitempty"`
	Method      string              `json:"method,omitempty"`
	Headers     map[string][]string `json:"headers,omitempty"`
	Body        []byte              `json:"body,omitempty"`
}
