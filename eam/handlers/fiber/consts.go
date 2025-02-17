package fiber

// List of API paths.
const (
	PathPolicies   = "/policies"
	PathPolicy     = "/policy/:id"
	PathAttributes = "/attributes"
	PathAttribute  = "/attribute/:key"
	PathEntities   = "/entities"
	PathEntity     = "/entity/:type/:id"
	PathRelations  = "/relations"
	PathRelation   = "/relation/:subjectType/:subjectId/:relation/:objectType/:objectId"
	PathAuthlog    = "/authlog"
	PathResource   = "/resource/:resource"
)
