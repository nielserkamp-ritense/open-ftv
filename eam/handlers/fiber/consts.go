package fiber

// List of API paths.
const (
	PathADL              = "/adl"
	PathAttribute        = "/attribute/:key"
	PathAttributeStatus  = "/attribute/:key/Status"
	PathAttributes       = "/attributes"
	PathAuthZEN          = "/authzen"
	PathAuthZenConfig    = "/authzen-configuration"
	PathBundle           = "/bundle"
	PathBundleID         = "/bundle/:id"
	PathCompressionTypes = "/compression-types"
	PathConfigs          = "/bundle-configurations"
	PathDeployment       = "/deployment"
	PathDeploymentID     = "/deployment/:key"
	PathDeployments      = "/deployments"
	PathEntities         = "/entities"
	PathEntity           = "/entity/:type/:id"
	PathEntityStatus     = "/entity/:type/:id/status"
	PathEntries          = "/entries"
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
	PathPolicyStatus     = "/policy/:id/status"
	PathPolicyVersions   = "/policy/:id/versions"
	PathPolicyVersion    = "/policy/:id/version/:version"
	PathPolicyRestore    = "/policy/:id/version/:version/restore"
	PathReadyZ           = "/readyz"
	PathRelation         = "/relation/:id"
	PathRelationStatus   = "/relation/:id/status"
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
