package fiber

// List of API paths.
const (
	PathADL               = "/adl"
	PathAttribute         = "/attribute/:key"
	PathAttributeRestore  = "/attribute/:key/version/:version/restore"
	PathAttributeStatus   = "/attribute/:key/Status"
	PathAttributeVersion  = "/attribute/:key/version/:version"
	PathAttributeVersions = "/attribute/:key/versions"
	PathAttributes        = "/attributes"
	PathAuthZEN           = "/authzen"
	PathAuthZenConfig     = "/authzen-configuration"
	PathBundle            = "/bundle"
	PathBundleID          = "/bundle/:id"
	PathCompressionTypes  = "/compression-types"
	PathConfigs           = "/bundle-configurations"
	PathDeployment        = "/deployment"
	PathDeploymentID      = "/deployment/:key"
	PathDeployments       = "/deployments"
	PathEntities          = "/entities"
	PathEntity            = "/entity/:type/:id"
	PathEntityRestore     = "/entity/:type/:id/version/:version/restore"
	PathEntityStatus      = "/entity/:type/:id/status"
	PathEntityVersion     = "/entity/:type/:id/version/:version"
	PathEntityVersions    = "/entity/:type/:id/versions"
	PathEntries           = "/entries"
	PathEvaluation        = "/evaluation"
	PathEvaluations       = "/evaluations"
	PathHealthZ           = "/healthz"
	PathLanguage          = "/language/:language"
	PathLanguages         = "/languages"
	PathLastDeployment    = "/deployment/last"
	PathLiveZ             = "/livez"
	PathMetadata          = "/metadata"
	PathPolicies          = "/policies"
	PathPolicy            = "/policy/:id"
	PathPolicyRestore     = "/policy/:id/version/:version/restore"
	PathPolicyStatus      = "/policy/:id/status"
	PathPolicyVersion     = "/policy/:id/version/:version"
	PathPolicyVersions    = "/policy/:id/versions"
	PathReadyZ            = "/readyz"
	PathRelation          = "/relation/:id"
	PathRelationStatus    = "/relation/:id/status"
	PathRelations         = "/relations"
	PathResource          = "/resource/:resource"
	PathSearchAction      = "/search/action"
	PathSearchResource    = "/search/resource"
	PathSearchSubject     = "/search/subject"
	PathStatuses          = "/statuses"
	PathTag               = "/tag/:tag"
	PathTags              = "/tags"
	PathV1                = "/v1"
	PathWellKnown         = "/.well-known"
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
