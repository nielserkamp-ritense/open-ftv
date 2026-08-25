package models

// Vocabulary of the management plane's authorization model: the resource types a
// management-plane request acts on, the actions a principal performs on them, and the
// resource properties a policy may decide on. See docs/adr/0006-management-plane-authorization-model.md.

// List of management-plane resource types, one per administered object.
const (
	EntityTypePolicy     = "policy"
	EntityTypeTag        = "tag"
	EntityTypeDeployment = "deployment"
	EntityTypeSetting    = "setting"
	EntityTypeAttribute  = "attribute"
	EntityTypeEntity     = "entity"
	EntityTypeLanguage   = "language"
	EntityTypePrincipal  = "principal"
	EntityTypeDecision   = "decision"
)

// EntityTypeManagementAction is the Cedar entity type of a management-plane action
// (`action == Action::"read"`).
const EntityTypeManagementAction = "Action"

// ResourceCollection is the resource id that addresses every object of a type, used for
// collection reads: `policy::"*"` is still `resource is policy`.
const ResourceCollection = "*"

// List of management-plane actions. The generic four apply to every type; the business
// actions name the operations the role matrix separates from a plain update.
const (
	ActionRead    = "read"
	ActionCreate  = "create"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionAccept  = "accept"  // status -> accepted.
	ActionDeploy  = "deploy"  // status -> deployed, and publishing a deployment.
	ActionRevert  = "revert"  // status -> concept: an accepted policy is withdrawn for editing.
	ActionRestore = "restore" // an older version becomes the current concept.
)

// List of resource properties a management-plane policy may decide on.
const (
	AttrStatus   = "status"
	AttrTags     = "tags"
	AttrLanguage = "language"
)
