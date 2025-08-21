package pap

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// ListLanguages returns a sorted list of all supported policy languages.
func (p *PAP) ListLanguages() (out policies.Languages, err error) {
	return p.languageDB.ListLanguages(p.ctx)
}
