package server

import (
	"testing"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/inzicht/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl/query"
)

func TestBuildSourceSelection(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*config.Config)
		want    any
		wantErr bool
	}{
		{
			name:   "default is wal",
			mutate: func(c *config.Config) { c.ADL.Path = "/tmp/adl.jsonl" },
			want:   (*query.WALSource)(nil),
		},
		{
			name:   "explicit wal",
			mutate: func(c *config.Config) { c.Inzicht.Backend = "wal"; c.ADL.Path = "/tmp/adl.jsonl" },
			want:   (*query.WALSource)(nil),
		},
		{
			name:   "loki",
			mutate: func(c *config.Config) { c.Inzicht.Backend = "loki"; c.Inzicht.LokiURL = "http://loki:3100" },
			want:   (*query.LokiSource)(nil),
		},
		{
			name:    "loki without url errors",
			mutate:  func(c *config.Config) { c.Inzicht.Backend = "loki" },
			wantErr: true,
		},
		{
			name:   "opensearch",
			mutate: func(c *config.Config) { c.Inzicht.Backend = "opensearch"; c.Inzicht.OpenSearchURL = "http://os:9200" },
			want:   (*query.OpenSearchSource)(nil),
		},
		{
			name:    "opensearch without url errors",
			mutate:  func(c *config.Config) { c.Inzicht.Backend = "opensearch" },
			wantErr: true,
		},
		{
			name:    "unknown backend errors",
			mutate:  func(c *config.Config) { c.Inzicht.Backend = "elasticsearch" },
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{}
			tc.mutate(cfg)

			src, err := buildSource(cfg)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got source %T", src)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotType, wantType := typeName(src), typeName(tc.want); gotType != wantType {
				t.Fatalf("source type = %s, want %s", gotType, wantType)
			}
		})
	}
}

func typeName(v any) string {
	switch v.(type) {
	case *query.WALSource:
		return "WALSource"
	case *query.LokiSource:
		return "LokiSource"
	case *query.OpenSearchSource:
		return "OpenSearchSource"
	default:
		return "unknown"
	}
}
