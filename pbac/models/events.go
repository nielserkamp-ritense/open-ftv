package models

// EventType indicates the type of PAP/PIP event.
type EventType uint8

// ListAllKeys of possible PAP/PIP events.
const (
	PolicyAdded EventType = iota + 1
	PolicyReplaced
	PolicyRemoved
	AttributeAdded
	AttributeReplaced
	AttributeRemoved
	EntityAdded
	EntityReplaced
	EntityRemoved
	RelationAdded
	RelationReplaced
	RelationRemoved
)

// String implements the Stringer interface.
func (e EventType) String() string {
	switch e {
	case PolicyAdded:
		return "policy added"
	case PolicyReplaced:
		return "policy replaced"
	case PolicyRemoved:
		return "policy removed"
	case AttributeAdded:
		return "attribute added"
	case AttributeReplaced:
		return "attribute replaced"
	case AttributeRemoved:
		return "attribute removed"
	case EntityAdded:
		return "entity added"
	case EntityReplaced:
		return "entity replaced"
	case EntityRemoved:
		return "entity removed"
	case RelationAdded:
		return "relation added"
	case RelationReplaced:
		return "relation replaced"
	case RelationRemoved:
		return "relation removed"
	default:
		return "<invalid>"
	}
}

// EventSink represents the interface for handling events.
type EventSink interface {
	Handle(t EventType, key string)
}
