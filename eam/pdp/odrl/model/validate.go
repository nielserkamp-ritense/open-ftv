package model

import (
	"errors"
	"fmt"
	"strings"
)

// KnownExtraProfiles lists additional odrl:profile IRIs that are recognised
// alongside the ODRL-AP-NL base profile and accepted without a warning.
var KnownExtraProfiles = []string{GeoNLProfile, BRPProfile}

// ValidateOptions configures policy validation, in particular how additional
// odrl:profile IRIs (beyond the ODRL-AP-NL base profile) are treated.
type ValidateOptions struct {
	// RejectUnknownProfiles rejects a policy that carries an odrl:profile IRI
	// which is neither the ODRL-AP-NL base profile nor listed in KnownExtra.
	// The broad default (false) accepts any additional profile and only reports
	// it as a warning.
	RejectUnknownProfiles bool
	// KnownExtra lists extra profile IRIs accepted without a warning. When nil,
	// KnownExtraProfiles is used.
	KnownExtra []string
}

func (o ValidateOptions) knownExtra() []string {
	if o.KnownExtra != nil {
		return o.KnownExtra
	}
	return KnownExtraProfiles
}

// profileSet returns the full odrl:profile set of the policy, falling back to
// the single Profile field for models built without the multi-valued field.
func (p *Policy) profileSet() []string {
	if len(p.Profiles) > 0 {
		return p.Profiles
	}
	if p.Profile != "" {
		return []string{p.Profile}
	}
	return nil
}

func containsProfile(set []string, prof string) bool {
	for _, s := range set {
		if s == prof {
			return true
		}
	}
	return false
}

// ValidatePolicy checks the mandatory ODRL-AP-NL base elements of a policy
// (conformance level 1) using the default (permissive) options.
func ValidatePolicy(p *Policy) error {
	_, err := ValidatePolicyOpts(p, ValidateOptions{})
	return err
}

// ValidatePolicyOpts checks the mandatory ODRL-AP-NL base elements of a policy:
// uid, title, publisher and an odrl:profile set that contains the ODRL-AP-NL
// base profile. Additional profiles (e.g. geonl, brp) are accepted; unknown
// extra profiles are returned as warnings unless RejectUnknownProfiles is set.
func ValidatePolicyOpts(p *Policy, opts ValidateOptions) (warnings []string, err error) {
	var errs []error

	if p.UID == "" || strings.HasPrefix(p.UID, "_:") {
		errs = append(errs, errors.New("policy is missing a resolvable odrl:uid"))
	}
	if len(p.Titles) == 0 {
		errs = append(errs, fmt.Errorf("policy %s is missing dct:title", p.UID))
	}
	if p.Publisher == "" {
		errs = append(errs, fmt.Errorf("policy %s is missing dct:publisher", p.UID))
	}

	profiles := p.profileSet()
	switch {
	case len(profiles) == 0:
		errs = append(errs, fmt.Errorf("policy %s is missing odrl:profile", p.UID))
	case !containsProfile(profiles, APNLProfile):
		errs = append(errs, fmt.Errorf("policy %s does not declare the ODRL-AP-NL base profile %s", p.UID, APNLProfile))
	default:
		for _, prof := range profiles {
			if prof == APNLProfile || containsProfile(opts.knownExtra(), prof) {
				continue
			}
			if opts.RejectUnknownProfiles {
				errs = append(errs, fmt.Errorf("policy %s declares unsupported profile %s", p.UID, prof))
			} else {
				warnings = append(warnings, fmt.Sprintf("policy %s declares additional unrecognised profile %s (accepted)", p.UID, prof))
			}
		}
	}

	return warnings, errors.Join(errs...)
}

// Validate checks all policies in the document against the mandatory
// ODRL-AP-NL base elements using the default (permissive) options. Documents
// without any policy are rejected.
func (d *Document) Validate() error {
	_, err := d.ValidateOpts(ValidateOptions{})
	return err
}

// ValidateOpts validates every policy in the document with the given options,
// returning the accumulated warnings (e.g. unrecognised extra profiles) next to
// the first fatal error.
func (d *Document) ValidateOpts(opts ValidateOptions) ([]string, error) {
	if len(d.Policies) == 0 {
		return nil, errors.New("document contains no odrl:Set/Offer/Agreement")
	}

	var errs []error
	var warnings []string
	for _, p := range d.Policies {
		w, err := ValidatePolicyOpts(p, opts)
		warnings = append(warnings, w...)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return warnings, errors.Join(errs...)
}
