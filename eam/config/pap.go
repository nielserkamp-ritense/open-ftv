package config

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// PAP contains the configuration variables for a generic PAP.
type PAP struct {
	Language     string `json:"policyLanguage,omitempty" yaml:"policies.language,omitempty"      env:"POLICIES_LANGUAGE"      flag:"policies-language,language" desc:"Language used for policy files" default:"CEDAR"`
	Store        string `json:"policyStore,omitempty"    yaml:"policies.store.path,omitempty"    env:"POLICIES_STORE"         flag:"policies-store"             desc:"Path where policy files are stored"`
	StoreRecurse bool   `json:"policyRecurse,omitempty"  yaml:"policies.store.recurse,omitempty" env:"POLICIES_STORE_RECURSE" flag:"policies-store-recurse"     desc:"Search policy file storage recursively"`
	TagsPath     string `json:"tagsPath,omitempty"       yaml:"policies.tags.path,omitempty"     env:"TAGS_PATH"              flag:"tags-path"                  desc:"Path where the policy tag file is stored"`
	// hidden fields.
	tags []*policies.Tag
}

// NewPAP instantiates a new PAP using the given configuration.
func (p *PAP) NewPAP(ctx context.Context, logger *slog.Logger) (*pap.PAP, error) {
	return pap.New(ctx, logger, pap.WithLanguage(p.Language), pap.WithFileStore(p.Store, p.StoreRecurse)), nil
}

// Tags returns the list of configured tags.
func (p *PAP) Tags() []*policies.Tag {
	return p.tags
}

// FixTags initializes the list of tags from the configured path.
func (p *PAP) FixTags() error {
	p.tags = make([]*policies.Tag, 0)
	return filepath.WalkDir(p.TagsPath, p.procesFile)
}

func (p *PAP) procesFile(path string, d os.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		return nil // always recurse subdirectories.
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return p.procesYAML(path)
	case ".json":
		return p.procesJSON(path)
	default:
		return nil
	}
}

func (p *PAP) procesYAML(path string) error {
	list := make([]*policies.Tag, 0)

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = yaml.NewDecoder(f).Decode(&list); err != nil {
		return err
	}

	p.tags = append(p.tags, list...)
	return nil
}

func (p *PAP) procesJSON(path string) error {
	list := make([]*policies.Tag, 0)

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = json.NewDecoder(f).Decode(&list); err != nil {
		return err
	}

	p.tags = append(p.tags, list...)
	return nil
}
