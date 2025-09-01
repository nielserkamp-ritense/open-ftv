package fiber

// List of API paths.
const (
	PathLanguages        = "/languages"
	PathLanguage         = "/language/:language"
	PathTags             = "/tags"
	PathTag              = "/tag/:tag"
	PathPolicies         = "/policies"
	PathPolicy           = "/policy/:id"
	PathAttributes       = "/attributes"
	PathAttribute        = "/attribute/:key"
	PathEntities         = "/entities"
	PathEntity           = "/entity/:type/:id"
	PathRelations        = "/relations"
	PathRelation         = "/relation/:id"
	PathStatuses         = "/statuses"
	PathCompressionTypes = "/compression-types"
	PathConfigs          = "/bundle-configurations"
	PathDeployments      = "/deployments"
	PathDeployment       = "/deployment"
	PathDeploymentID     = "/deployment/:key"
	PathLastDeployment   = "/deployment/last"
	PathBundleID         = "/bundle/:id"
	PathBundle           = "/bundle"
	PathAuthlog          = "/authlog"
	PathResource         = "/resource/:resource"
)

// HeaderVersion is the header for reporting the full semantic API version.
const HeaderVersion = "API-Version"
