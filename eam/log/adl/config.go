package adl

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"runtime/debug"
)

// Config configures an ADL Logger.
type Config struct {
	// Level is the level of detail (1-4). Values below 1 are treated as 1, above 4 as 4.
	//   1: request/response only.
	//   2: + adl.core.policies version references.
	//   3: + adl.core.information source references (requires an InformationProvider).
	//   4: + adl.core.configuration (engine/version/config-hash/request-mappings).
	Level int

	// WALPath is the file path of the durable write-ahead log (JSONL). When empty no WAL
	// is written; the standard RECOMMENDS a WAL, so a path SHOULD be configured.
	WALPath string

	// Resource identifies the producer of the records. It MUST be set such that records
	// can be unambiguously attributed to their producer when aggregated across organisations.
	Resource map[string]any

	// Engine is the policy-engine language (e.g. "odrl"); it feeds body.adl.core.configuration
	// and the derived config hash.
	Engine string

	// EngineVersion is an optional operator-supplied engine/build label. The recorded
	// configuration version is derived from the process build info (runtime/debug) and,
	// when set, this label; a container image digest is expected to be pinned at the
	// deployment level, not self-reported here.
	EngineVersion string

	// ConfigHash, when set, overrides the automatically derived configuration hash
	// (ADL_CONFIG_HASH). When empty, config_hash is derived from a canonical JSON of the
	// evaluation-determining configuration (engine language, request mappings, PIP and
	// PAP configuration, ADL level).
	ConfigHash string

	// RequestMappings is the PDP request-mapping configuration string. It transforms the
	// PARC before evaluation and therefore co-determines the decision, so it is recorded
	// verbatim in adl.core.configuration and folded into the derived config hash.
	RequestMappings string

	// PIPStore, PIPPull and PIPWARC describe the information-point configuration
	// (attribute store path, pull-config path, WARC directory). PAPStore describes the
	// policy-point store path. They co-determine evaluation and feed the derived config
	// hash; they are not emitted verbatim (paths are deployment detail) but are hashed.
	PIPStore string
	PIPPull  string
	PIPWARC  string
	PAPStore string

	// Logger receives diagnostics from the asynchronous flush path. Optional.
	Logger *slog.Logger

	// Information is the level-3 PIP hook. Optional; when nil, no information sources are
	// recorded even at level >= 3.
	Information InformationProvider
}

func (c *Config) normalise() {
	if c.Level < 1 {
		c.Level = 1
	}
	if c.Level > 4 {
		c.Level = 4
	}
}

func (c *Config) resourceCopy() map[string]any {
	if len(c.Resource) == 0 {
		return nil
	}
	out := make(map[string]any, len(c.Resource))
	for k, v := range c.Resource {
		out[k] = v
	}
	return out
}

// configurationBody builds the raw body.adl.core.configuration payload for level 4.
//
// It captures the machine-readable configuration a verifier needs to reconstruct the
// evaluation environment: the engine language, a real version derived from the build
// info, the request-mappings string (which transforms the request before evaluation),
// and a config_hash that is either the operator override or automatically derived from
// the full evaluation-determining configuration.
func (c *Config) configurationBody() map[string]any {
	conf := map[string]any{}
	if c.Engine != "" {
		conf["engine"] = c.Engine
	}
	if v := c.resolveVersion(); v != "" {
		conf["version"] = v
	}
	if c.RequestMappings != "" {
		conf["request_mappings"] = c.RequestMappings
	}
	if h := c.resolveConfigHash(); h != "" {
		conf["config_hash"] = h
	}
	return conf
}

// resolveVersion returns the recorded configuration version: the process build info
// (module version and, when present, the VCS revision), optionally suffixed with the
// operator-supplied EngineVersion label. A container image digest belongs at the
// deployment level and is not self-reported here.
func (c *Config) resolveVersion() string {
	v := buildVersion()
	if c.EngineVersion == "" {
		return v
	}
	if v == "" {
		return c.EngineVersion
	}
	return v + " (" + c.EngineVersion + ")"
}

// buildVersion reads the module version and VCS revision from the build info embedded
// by the Go toolchain. It returns "" when no build info is available (e.g. `go test`
// without module info).
func buildVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	v := bi.Main.Version
	if v == "(devel)" {
		v = "" // no meaningful module version (e.g. `go test`, unversioned build).
	}
	var rev string
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" {
			rev = s.Value
			break
		}
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	switch {
	case v != "" && rev != "":
		return v + "+" + rev
	case rev != "":
		return rev
	default:
		return v
	}
}

// resolveConfigHash returns the operator override when set, otherwise the derived hash.
func (c *Config) resolveConfigHash() string {
	if c.ConfigHash != "" {
		return c.ConfigHash
	}
	return c.derivedConfigHash()
}

// derivedConfigHash computes a SHA-256 over a canonical JSON of the configuration that
// determines the evaluation outcome. Struct field order makes the JSON canonical.
func (c *Config) derivedConfigHash() string {
	canon := struct {
		Engine          string `json:"engine"`
		Version         string `json:"version"`
		RequestMappings string `json:"request_mappings"`
		PIPStore        string `json:"pip_store"`
		PIPPull         string `json:"pip_pull"`
		PIPWARC         string `json:"pip_warc"`
		PAPStore        string `json:"pap_store"`
		Level           int    `json:"adl_level"`
	}{
		Engine:          c.Engine,
		Version:         c.resolveVersion(),
		RequestMappings: c.RequestMappings,
		PIPStore:        c.PIPStore,
		PIPPull:         c.PIPPull,
		PIPWARC:         c.PIPWARC,
		PAPStore:        c.PAPStore,
		Level:           c.Level,
	}
	b, err := json.Marshal(canon)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
