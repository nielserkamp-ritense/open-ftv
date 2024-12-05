package models

// List of HTTP header keys.
//
// NOTE: keep values in lower-case as comparison is always done in lower-case!
const (
	HeaderApiKey           = "api-key"
	HeaderAuthorization    = "authorization"
	HeaderContentType      = "content-type"
	HeaderCoreUser         = "x-dpl-core-user"
	HeaderDoelbinding      = "doelbinding"
	HeaderFSCAuthorization = "fsc-authorization"
	HeaderForwarded        = "forwarded"
	HeaderGrondslag        = "grondslag"
	HeaderRvaActivityID    = "x-dpl-rva-activity-id"
	HeaderTaak             = "taak"
	HeaderXForwardedFor    = "x-forwarded-for"
	HeaderZaakType         = "zaak-type"
)
