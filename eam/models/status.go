package models

import "strings"

// Status represents the status of an object.
type Status uint8

// List of supported statuses.
const (
	StatusConcept Status = iota + 1
	StatusAccepted
	StatusDeployed
)

// String implements the Stringer interface.
func (s Status) String() string {
	switch s {
	case StatusConcept:
		return "concept"
	case StatusAccepted:
		return "accepted"
	case StatusDeployed:
		return "deployed"
	default:
		return "???"
	}
}

// StatusFromString determines the status from the given string.
func StatusFromString(s string) Status {
	switch strings.ToLower(s) {
	case "accepted", "accept":
		return StatusAccepted
	case "deployed", "deploy":
		return StatusDeployed
	default:
		return StatusConcept
	}
}
