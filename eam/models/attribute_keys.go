package models

// List of standard attribute keys.
const (
	AttrAction          = "action"
	AttrActieURI        = "actie_uri" // canonical FTV action URI, added alongside the short action name.
	AttrAPIKey          = "api_key"
	AttrAVGHandeling    = "nl.avg.verwerkingshandeling" // AVG verwerkingshandeling name (informatiemodel).
	AttrBasicUser       = "basic_user"
	AttrBasicPswd       = "basic_pswd"
	AttrBody            = "body"
	AttrClaims          = "claims"
	AttrClientIP        = "ip_address"
	AttrClientPrincipal = "client_principal"
	AttrContentType     = "content_type"
	AttrCoreUser        = "core_user"
	AttrDeviceID        = "device_id"
	AttrDoelbinding     = "doelbinding"
	AttrFSC             = "fsc"
	AttrGrondslag       = "grondslag"
	AttrHeaders         = "headers"
	AttrHost            = "host"
	AttrHTTP            = "http"
	AttrJWT             = "jwt"
	AttrLDContext       = "ld-context" // NLGov: Linked Data context (URL or object) for the request.
	AttrMethod          = "method"
	AttrMIM             = "mim" // NLGov: URL to the meta-informatiemodel (self-describing request).
	AttrPath            = "path"
	AttrPathParts       = "path_parts"
	AttrPrincipal       = "principal"
	// AttrProcessingActivityID and AttrProcessingActivityDPL both reference a processing
	// activity (verwerking) in a Verwerkingenregister; the DPL variant is the informatiemodel form.
	AttrProcessingActivityID  = "processing_activity_id"
	AttrProcessingActivityDPL = "dpl.core.processing_activity_id"
	AttrQuery                 = "query"
	AttrRequestTime           = "request_time"
	AttrResource              = "resource"
	AttrRvvaID                = "rvva_id"
	AttrScheme                = "scheme"
	AttrTaak                  = "taak"
	AttrTime                  = "time"
	AttrTraceParent           = "traceparent" // https://www.w3.org/TR/trace-context/
	AttrTraceState            = "tracestate"  // https://www.w3.org/TR/trace-context/
	AttrTypeURI               = "type_uri"    // canonical FTV subject/resource type URI, added alongside the short type.
	AttrValid                 = "valid"
	AttrZaakType              = "zaak_type"
)
