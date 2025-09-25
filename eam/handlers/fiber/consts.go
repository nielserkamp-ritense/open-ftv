package fiber

// List of API paths.
const (
	PathAttribute        = "/attribute/:key"
	PathAttributes       = "/attributes"
	PathAuthZEN          = "/authzen"
	PathAuthZenConfig    = "/authzen-configuration"
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
	PathEvaluations      = "/evaluations"
	PathHealthZ          = "/healthz"
	PathLanguage         = "/language/:language"
	PathLanguages        = "/languages"
	PathLastDeployment   = "/deployment/last"
	PathLiveZ            = "/livez"
	PathMetadata         = "/metadata"
	PathPolicies         = "/policies"
	PathPolicy           = "/policy/:id"
	PathReadyZ           = "/readyz"
	PathRelation         = "/relation/:id"
	PathRelations        = "/relations"
	PathResource         = "/resource/:resource"
	PathSearchAction     = "/search/action"
	PathSearchResource   = "/search/resource"
	PathSearchSubject    = "/search/subject"
	PathStatuses         = "/statuses"
	PathTag              = "/tag/:tag"
	PathTags             = "/tags"
	PathV1               = "/v1"
	PathWellKnown        = "/.well-known"
)

// List of header codes.
const (
	HeaderVersion   = "API-Version"
	HeaderRequestID = "X-Request-ID"
)

// List of parameter codes.
const (
	ParamForceUpsert   = "forceUpsert"
	ParamIgnoreMissing = "ignoreMissing"
)
