// Package models contains generic definitions for working External Access Management (EAM).
package models

import (
	"net/url"
	"time"

	"github.com/google/uuid"
)

// PARC contains the details for an authorization control request.
type PARC struct {
	Principal Entity       `json:"principal,omitempty"`
	Action    Entity       `json:"action,omitempty"`
	Resource  Entity       `json:"resource,omitempty"`
	Context   AttributeSet `json:"context,omitempty"`
}

// Request contains the details of an access control request.
type Request struct {
	UID         *uuid.UUID          `json:"uid,omitempty"`
	URL         *url.URL            `json:"url,omitempty"`
	Method      string              `json:"method,omitempty"`
	RequestTime *time.Time          `json:"requestTime,omitempty"`
	Principal   Entity              `json:"principal,omitempty"`
	Action      Entity              `json:"action,omitempty"`
	Resource    Entity              `json:"resource,omitempty"`
	Headers     map[string][]string `json:"headers,omitempty"`
	Body        []byte              `json:"body,omitempty"`
	Attributes  map[string]any      `json:"attributes,omitempty"`
}

// Response contains the result of an access control request.
type Response struct {
	Allowed    bool                `json:"allowed,omitempty"`
	Message    string              `json:"message,omitempty"`
	NewURL     *url.URL            `json:"newURL,omitempty"`
	NewBody    []byte              `json:"newBody,omitempty"`
	NewHeaders map[string][]string `json:"newHeaders,omitempty"`
	Attributes map[string]any      `json:"attributes,omitempty"`
	PolicyKey  string              `json:"policyKey,omitempty"`
	PolicyHash string              `json:"policyHash,omitempty"`
}
