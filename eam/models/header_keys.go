package models

// List of HTTP header keys.
//
// NOTE: keep values in lower-case as comparison is always done in lower-case!
const (
	HeaderAPIKey           = "api-key"
	HeaderAuthorization    = "authorization"
	HeaderContentType      = "content-type"
	HeaderCoreUser         = "dpl-core-user"
	HeaderDeviceID         = "device-id"
	HeaderDoelbinding      = "doelbinding"
	HeaderFSCAuthorization = "fsc-authorization"
	HeaderFSCTransactionID = "fsc-transaction-id" // https://gitdocumentatie.logius.nl/publicatie/fsc/logging/1.1.0/
	HeaderForwarded        = "forwarded"
	HeaderGrondslag        = "grondslag"
	HeaderObsoleteCoreUser = "x-dpl-core-user"
	HeaderObsoleteDeviceID = "x-device-id"
	HeaderObsoleteRvvaID   = "x-dpl-rva-activity-id"
	HeaderRvvaID           = "dpl-processing-activity-id"
	HeaderTaak             = "taak"
	HeaderTraceParent      = "traceparent" // https://www.w3.org/TR/trace-context/
	HeaderTraceState       = "tracestate"  // https://www.w3.org/TR/trace-context/
	HeaderXForwardedFor    = "x-forwarded-for"
	HeaderZaakType         = "zaak-type"
)
