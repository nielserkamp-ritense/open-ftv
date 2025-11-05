package models

// EventType indicates the type of PAP/PIP event.
type EventType uint8

// List of possible PAP/PIP events.
const (
	PolicyAdded EventType = iota + 1
	PolicyReplaced
	PolicyRemoved
	AttributeAdded
	AttributeReplaced
	AttributeRemoved
	AttributesAdded
	AttributesReplaced
	AttributesRemoved
	EntityAdded
	EntityReplaced
	EntityRemoved
	EntitiesAdded
	EntitiesReplaced
	EntitiesRemoved
	RelationAdded
	RelationReplaced
	RelationRemoved
	RelationsAdded
	RelationsReplaced
	RelationsRemoved
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
	case AttributesAdded:
		return "attributes added"
	case AttributesReplaced:
		return "attributes replaced"
	case AttributesRemoved:
		return "attributes removed"
	case EntityAdded:
		return "entity added"
	case EntityReplaced:
		return "entity replaced"
	case EntityRemoved:
		return "entity removed"
	case EntitiesAdded:
		return "entities added"
	case EntitiesReplaced:
		return "entities replaced"
	case EntitiesRemoved:
		return "entities removed"
	case RelationAdded:
		return "relation added"
	case RelationReplaced:
		return "relation replaced"
	case RelationRemoved:
		return "relation removed"
	case RelationsAdded:
		return "relations added"
	case RelationsReplaced:
		return "relations replaced"
	case RelationsRemoved:
		return "relations removed"
	default:
		return "<invalid>"
	}
}

// EventSink represents the interface for handling events.
type EventSink interface {
	Handle(t EventType, key string)
}
