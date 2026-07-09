package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	odrl_geo "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl-geo"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

// exampleTTLPath points at the shared ODRL-Geo example policies.
const exampleTTLPath = "../../eam/pdp/odrl-geo/testdata/example3.ttl"

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

// wrapEnvelope wraps raw turtle in the PAP "odrl" storage envelope, exactly as the PAP
// stores an ODRL-Geo policy.
func wrapEnvelope(t *testing.T, ttl []byte) []byte {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"uid": "replay", "mimeType": "text/turtle", "source": base64.StdEncoding.EncodeToString(ttl),
	})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// example3Request is the permit-with-obligations request from the engine's own suite.
func example3Request() *models.PARC {
	return &models.PARC{
		Principal: models.NewEntity("organization", "https://identifier.overheid.nl/tooi/id/oorg/oorg55500", models.NewAttributeSet(nil)),
		Action:    models.NewEntity(models.EntityTypeName, "read", models.NewAttributeSet(map[string]any{"dpl.core.processing_activity_id": "https://w3id.org/dpv#ResearchAndDevelopment"})),
		Resource: models.NewEntity("feature", "urn:feature:1", models.NewAttributeSet(map[string]any{
			"layer": "https://api.example.gov.nl/bro/cpt/v1/collections/sonderingen/items",
			"crs":   "http://www.opengis.net/def/crs/EPSG/0/28992",
			// point inside the CPT area.
			"geometry": "POINT (155000 463000)",
		})),
		Context: models.NewAttributeSet(nil),
	}
}

// generateRecord evaluates the request against the given policy envelope and writes a
// level-4 ADL record to a WAL, returning the WAL path. It mirrors what the PDP handler
// records (DecisionContext carries policy, policyHash and obligations).
func generateRecord(t *testing.T, dir string, envelope []byte) string {
	t.Helper()
	ctx := context.Background()
	logger := discardLogger()

	ap := pap.New(ctx, logger, pap.WithLanguage("odrl"))
	pol, err := pap.NewPolicyFromData("example3", "odrl", "", "urn:example3", bytes.NewReader(envelope))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ap.Create(pol); err != nil {
		t.Fatal(err)
	}

	c := odrl_geo.NewController(pdp.WithContext(ctx), pdp.WithLogger(logger), pdp.WithPAP(ap), pdp.WithPIP(pip.New(ctx, logger)))

	parc := example3Request()
	resp, err := c.Authorize("gen", parc)
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if !resp.Allowed {
		t.Fatalf("setup expected a permit, got deny: %q", resp.Message)
	}

	walPath := filepath.Join(dir, "adl.jsonl")
	l, err := adl.New(adl.Config{
		Level:    4,
		WALPath:  walPath,
		Resource: map[string]any{"service.name": "pdp-replay-test"},
		Engine:   "odrl",
		Logger:   logger,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()

	dc := models.NewAttributeSet()
	dc.AddAttribute("policy", resp.PolicyKey)
	dc.AddAttribute("policyHash", resp.PolicyHash)
	if obs := resp.Attributes["obligations"]; obs != nil {
		dc.AddAttribute("obligations", obs)
	}

	ar := &authlog.AuthRecord{
		EventName:       adl.EventAccessEvaluation,
		Principal:       parc.Principal,
		Action:          parc.Action,
		Resource:        parc.Resource,
		RequestContext:  parc.Context,
		DecisionContext: dc,
		Decision:        resp.Allowed,
	}
	if err := l.Log(ctx, true, ar); err != nil {
		t.Fatalf("log: %v", err)
	}
	return walPath
}

// writePolicyDir writes a single policy-source file into a fresh directory and returns it.
func writePolicyDir(t *testing.T, base, name string, envelope []byte) string {
	t.Helper()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "example3.json"), envelope, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestReplay_PassThenFailOnMutation(t *testing.T) {
	ttl, err := os.ReadFile(exampleTTLPath)
	if err != nil {
		t.Fatal(err)
	}
	envelope := wrapEnvelope(t, ttl)

	base := t.TempDir()

	// 1. Generate a fresh level-4 record from a real evaluation.
	walDir := filepath.Join(base, "wal")
	if err := os.MkdirAll(walDir, 0o755); err != nil {
		t.Fatal(err)
	}
	walPath := generateRecord(t, walDir, envelope)

	rec, err := ReadRecord(walPath)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}

	// Sanity: the record is a genuine level-4 record with a resolvable policy hash.
	if mapAt(rec.Body, "adl.core.configuration") == nil {
		t.Fatal("record has no adl.core.configuration (not level 4)")
	}
	hashes := policyHashes(rec)
	if len(hashes) == 0 {
		t.Fatal("record has no adl.core.policies hash reference")
	}
	if hashes[0] != pap.HashContent(envelope) {
		t.Fatalf("logged policy hash %s != content hash %s", hashes[0], pap.HashContent(envelope))
	}

	// Optional: materialise the artifacts so the CLI binary can be driven end-to-end.
	if dump := os.Getenv("ADL_REPLAY_DUMP"); dump != "" {
		_ = os.MkdirAll(dump, 0o755)
		good := writePolicyDir(t, dump, "policies", envelope)
		data, _ := os.ReadFile(walPath)
		_ = os.WriteFile(filepath.Join(dump, "record.jsonl"), data, 0o644)
		t.Logf("dumped record.jsonl and policy dir %s to %s", good, dump)
	}

	// 2. Replay against the unmodified policy: PASS.
	goodDir := writePolicyDir(t, base, "good", envelope)
	res, err := Replay(context.Background(), rec, Options{PolicyDir: goodDir, Logger: discardLogger()})
	if err != nil {
		t.Fatalf("replay (good): %v", err)
	}
	if !res.Pass {
		t.Fatalf("expected PASS on unmodified policy, got FAIL:\n%s", res.Diff)
	}
	if !res.ResolvedHashes[hashes[0]] {
		t.Fatalf("expected policy hash %s to resolve in the reconstructed PAP", hashes[0])
	}

	// 3. Mutate the policy (change the generalisation obligation 50000 -> 77777) and
	//    replay the same record: FAIL, because the re-evaluated obligations differ and
	//    the logged hash no longer resolves to the mutated content.
	mutatedTTL := bytes.Replace(ttl, []byte(`"50000"^^xsd:integer`), []byte(`"77777"^^xsd:integer`), 1)
	if bytes.Equal(mutatedTTL, ttl) {
		t.Fatal("mutation had no effect; example policy changed shape")
	}
	badDir := writePolicyDir(t, base, "bad", wrapEnvelope(t, mutatedTTL))
	res2, err := Replay(context.Background(), rec, Options{PolicyDir: badDir, Logger: discardLogger()})
	if err != nil {
		t.Fatalf("replay (bad): %v", err)
	}
	if res2.Pass {
		t.Fatal("expected FAIL on mutated policy, got PASS")
	}
	if !strings.Contains(res2.Diff, "obligations") && !strings.Contains(res2.Diff, "did not resolve") {
		t.Fatalf("expected a meaningful diff, got:\n%s", res2.Diff)
	}
	t.Logf("mutated-policy replay correctly FAILED:\n%s", res2.Diff)
}
