package fiber

// List of API paths.
const (
	PathAttribute        = "/attribute/:key"
	PathAttributes       = "/attributes"
	PathAuthZEN          = "/authzen"
	PathAuthlog          = "/authlog"
	PathBundle           = "/bundle"
	PathBundleID         = "/bundle/:id"
	PathCompressionTypes = "/compression-types"
	PathConfigs          = "/bundle-configurations"
	PathDeployment       = "/deployment"
	PathDeploymentID     = "/deployment/:key"
	PathDeployments      = "/deployments"
	PathEntities         = "/entities"
	PathEntity           = "/entity/:type/:id"
	PathEvaluation       = "/evaluation"
	PathHealthZ          = "/healthz"
	PathLanguage         = "/language/:language"
	PathLanguages        = "/languages"
	PathLastDeployment   = "/deployment/last"
	PathLiveZ            = "/livez"
	PathPolicies         = "/policies"
	PathPolicy           = "/policy/:id"
	PathReadyZ           = "/readyz"
	PathRelation         = "/relation/:id"
	PathRelations        = "/relations"
	PathResource         = "/resource/:resource"
	PathStatuses         = "/statuses"
	PathTag              = "/tag/:tag"
	PathTags             = "/tags"
	PathV1               = "/v1"
)

// HeaderVersion is the header for reporting the full semantic API version.
const HeaderVersion = "API-Version"
