package config

import "flag"

// PAP contains the configuration variables for a generic PAP.
type PAP struct {
	Language     string
	Store        string
	StoreRecurse bool
	TagsPath     string
}

func (p *PAP) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&p.Language, "policies-language", "CEDAR", "Language used for policy files")
	fs.StringVar(&p.Store, "policies-store", "", "Path where policy files are stored")
	fs.BoolVar(&p.StoreRecurse, "policies-store-recurse", false, "Search policy file storage recursively")
	fs.StringVar(&p.TagsPath, "tags-path", "", "Path where the policy tag file is stored")
}
