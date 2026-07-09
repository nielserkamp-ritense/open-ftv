// Package config handles the configurable options for the Inzicht service.
package config

import (
	"log/slog"
	"os"
	"time"

	"gitlab.com/gjuyn/go-config/config"
	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
)

const (
	// AppName defines the name and version of this application.
	AppName   = "Inzicht 0.1"
	envPrefix = "INZICHT_"
	cfg1      = "/etc/inzicht/default.conf"
	cfg2      = "./etc/inzicht.yaml"
	cfg3      = "./inzicht.yaml"
)

// Config represents the full set of configuration variables for the Inzicht app.
//
// It reuses the shared EAM config building blocks (server, ADL, persistence) so
// the Inzicht service has the same configuration surface as the other apps, plus
// an Inzicht-specific block for the statistics/push and Inzicht-API behaviour.
type Config struct {
	config2.ServerApp
	config2.ADL
	config2.Persist
	Inzicht
}

// Inzicht holds the Inzicht-service-specific configuration.
type Inzicht struct {
	// Source identifies this afnemer as the producer of statistics events and of
	// ADL access records. Recommended: the organisation/service identifier.
	Source string `yaml:"inzicht.source,omitempty" env:"SOURCE" flag:"inzicht-source" default:"nl.overheid.ftv.inzicht" desc:"Producer/source identifier for statistics CloudEvents and ADL access records"`

	// Backend selects the read-side query.Source that serves the statistics and
	// Inzicht endpoints: "wal" (default) reads the local write-ahead log, "loki"
	// queries a Grafana Loki backend, "opensearch" queries an OpenSearch index.
	//
	// NOTE: the backend selector is INZICHT_BACKEND, not INZICHT_SOURCE - the latter
	// is already the producer/source identifier above.
	Backend string `yaml:"inzicht.backend,omitempty" env:"BACKEND" flag:"inzicht-backend" default:"wal" desc:"Read-side backend serving the endpoints: wal, loki or opensearch"`

	// LokiURL is the Grafana Loki base URL, used when Backend=loki.
	LokiURL string `yaml:"inzicht.loki.url,omitempty" env:"LOKI_URL" flag:"inzicht-loki-url" desc:"Grafana Loki base URL (Backend=loki)"`
	// LokiOrgID sets X-Scope-OrgID for multi-tenant Loki (optional).
	LokiOrgID string `yaml:"inzicht.loki.orgID,omitempty" env:"LOKI_ORG_ID" flag:"inzicht-loki-org-id" desc:"Loki X-Scope-OrgID tenant header (optional)"`
	// LokiUser and LokiPassword enable HTTP basic auth against Loki (optional).
	LokiUser     string `yaml:"inzicht.loki.user,omitempty" env:"LOKI_USER" flag:"inzicht-loki-user" desc:"Loki basic-auth user (optional)"`
	LokiPassword string `yaml:"inzicht.loki.password,omitempty" env:"LOKI_PASSWORD" flag:"inzicht-loki-password" desc:"Loki basic-auth password (optional)"`

	// OpenSearchURL is the OpenSearch base URL, used when Backend=opensearch.
	OpenSearchURL string `yaml:"inzicht.opensearch.url,omitempty" env:"OPENSEARCH_URL" flag:"inzicht-opensearch-url" desc:"OpenSearch base URL (Backend=opensearch)"`
	// OpenSearchIndex is the index/alias holding ADL records (default "adl").
	OpenSearchIndex string `yaml:"inzicht.opensearch.index,omitempty" env:"OPENSEARCH_INDEX" flag:"inzicht-opensearch-index" default:"adl" desc:"OpenSearch index/alias holding ADL records"`
	// OpenSearchUser and OpenSearchPassword enable basic auth (optional).
	OpenSearchUser     string `yaml:"inzicht.opensearch.user,omitempty" env:"OPENSEARCH_USER" flag:"inzicht-opensearch-user" desc:"OpenSearch basic-auth user (optional)"`
	OpenSearchPassword string `yaml:"inzicht.opensearch.password,omitempty" env:"OPENSEARCH_PASSWORD" flag:"inzicht-opensearch-password" desc:"OpenSearch basic-auth password (optional)"`

	// K is the k-anonymity threshold applied to aggregate statistics: buckets with
	// fewer than K decisions are suppressed.
	K int `yaml:"inzicht.k,omitempty" env:"K" flag:"inzicht-k" default:"5" desc:"k-anonymity threshold for aggregate statistics (buckets with n<k are suppressed)"`

	// Bucket is the default period bucket for statistics (day/week/month).
	Bucket string `yaml:"inzicht.bucket,omitempty" env:"BUCKET" flag:"inzicht-bucket" default:"day" desc:"Default period bucket for statistics: day, week or month"`

	// PushInterval is the interval at which aggregate statistics are pushed to the
	// configured verstrekker endpoints. Zero disables the push component.
	PushInterval time.Duration `yaml:"inzicht.push.interval,omitempty" env:"PUSH_INTERVAL" flag:"inzicht-push-interval" desc:"Interval for pushing aggregate statistics to verstrekkers (0 disables)"`

	// PushTargets is a comma-separated list of verstrekker endpoint URLs that
	// receive the statistics CloudEvents.
	PushTargets string `yaml:"inzicht.push.targets,omitempty" env:"PUSH_TARGETS" flag:"inzicht-push-targets" desc:"Comma-separated verstrekker endpoint URLs receiving statistics CloudEvents"`

	// PushWindow is how far back each push aggregates. Defaults to 24h when zero.
	PushWindow time.Duration `yaml:"inzicht.push.window,omitempty" env:"PUSH_WINDOW" flag:"inzicht-push-window" default:"24h" desc:"Look-back window aggregated on each statistics push"`

	// AuthTokens configures the bearer-token access control for the Inzicht API.
	//
	// It is a comma-separated list of "token:role:verstrekker" triples, where role
	// is "verstrekker" or "beheerder" and verstrekker is the identity bound to the
	// token. Example: "s3cr3t:verstrekker:rvig,adm1n:beheerder:".
	//
	// This static token list is a deliberate SIMULATION stand-in for a production
	// OAuth2/mTLS token service: in a real deployment the bearer token would be an
	// OAuth2 access token (or the identity would come from an mTLS client cert),
	// verified against an authorization server, and the role/verstrekker identity
	// would be derived from verified claims - never from a configured constant or a
	// spoofable request header. When no token is configured the API is closed (401),
	// unless AuthDisabled is set for local play.
	AuthTokens string `yaml:"inzicht.auth.tokens,omitempty" env:"AUTH_TOKENS" flag:"inzicht-auth-tokens" desc:"Bearer tokens as comma-separated token:role:verstrekker triples (simulation stand-in for OAuth2/mTLS)"`

	// AuthDisabled turns off Inzicht API authentication entirely, for local play.
	// It is an explicit, dangerous opt-out: with it set, the caller identity falls
	// back to the (spoofable) X-Verstrekker header and every role check is bypassed.
	AuthDisabled bool `yaml:"inzicht.auth.disabled,omitempty" env:"AUTH_DISABLED" flag:"inzicht-auth-disabled" desc:"Disable Inzicht API auth for local play (identity from X-Verstrekker, all role checks bypassed)"`
}

// New instantiates a new configuration from config files, environment variables
// and/or command-line flags, following the same loading pattern as the other apps.
func New(opts ...config.Option) (*Config, *slog.Logger) {
	cfg := &Config{}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	opts = append(
		append(
			[]config.Option{
				yaml.FilesYAML(cfg1, cfg2, cfg3),
				config.EnvironmentPrefix(envPrefix),
				config.AppName(AppName),
			},
			opts...,
		),
		config.NoHelpOnError(),
	)

	logger = config2.Load(cfg, opts...)
	return cfg, logger
}

// LogSanitized implements the config.Printer interface.
func (c *Config) LogSanitized(logger *slog.Logger) {
	sanitized := *c
	sanitized.Persist.Sanitized()
	logger.Info("configuration loaded successfully", "config", sanitized)
}
