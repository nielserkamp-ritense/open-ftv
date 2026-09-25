package config

// PAP contains the configuration variables for a generic PAP.
type PAP struct {
	Language     string `env:"POLICIES_LANGUAGE"      flag:"policies-language,language" desc:"Language used for policy files" default:"CEDAR"`
	Store        string `env:"POLICIES_STORE"         flag:"policies-store"             desc:"Path where policy files are stored"`
	StoreRecurse bool   `env:"POLICIES_STORE_RECURSE" flag:"policies-store-recurse"     desc:"Search policy file storage recursively"`
	TagsPath     string `env:"TAGS_PATH"              flag:"tags-path"                  desc:"Path where the policy tag file is stored"`
}
