package server

import (
	"bufio"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

// TestADLInformationLevel3 is an integration test for finding A1: a PIP with a
// file-store (real load) plus a pulled attribute (recorded exactly as the network
// pull recorder does) must produce a non-empty adl.core.information at ADL level 3
// when a request references those values by the keys a PDP request delivers.
//
// Before the fix the PIP only registered references under the pulled/decoded entity
// UID and bare attribute name, while the ADL adapter looked up only the request
// entity UIDs - so level-3 information was virtually always empty. This test drives
// the real PIP source-ref store, the real informationProvider bridge and the real
// adl.Logger.Log write path, and asserts the references now match.
func TestADLInformationLevel3(t *testing.T) {
	dir := t.TempDir()

	// A file-store with one entity (with an attribute) and one standalone attribute.
	entDir := filepath.Join(dir, "store", "entities")
	attrDir := filepath.Join(dir, "store", "attributes")
	mustMkdir(t, entDir)
	mustMkdir(t, attrDir)
	mustWrite(t, filepath.Join(entDir, "user.yaml"),
		"- type: user\n  id: alice\n  attributes:\n    clearance: high\n")
	mustWrite(t, filepath.Join(attrDir, "attr.yaml"),
		"- key: tenant\n  value: gemeente-amsterdam\n")

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	p := pip.New(context.Background(), logger, pip.WithFileStore(filepath.Join(dir, "store"), true))

	// Simulate a pulled attribute: the network pull recorder registers a "Logged"
	// source reference for each decoded value, keyed by attribute name and by
	// "<entity-uid>/<attr>". Do exactly that through the public SourceReferencer.
	refs, ok := p.(pip.SourceReferencer)
	if !ok {
		t.Fatal("pip does not implement SourceReferencer")
	}
	refs.RecordSourceRef(pip.SourceRef{
		Kind: "attribute", Key: "risk_score",
		TraceID: "0af7651916cd43dd8448eb211c80319c", SpanID: "b7ad6b7169203331", WARCFile: "pull-0001.warc",
	})

	// Sanity: the file-store load registered a "file" reference for the entity and
	// its attribute under "<uid>/<attr>".
	if _, ok := refs.SourceRef("user::alice"); !ok {
		t.Fatal("expected file source ref for entity user::alice")
	}
	if _, ok := refs.SourceRef("user::alice/clearance"); !ok {
		t.Fatal("expected file source ref for attribute user::alice/clearance")
	}
	if _, ok := refs.SourceRef("tenant"); !ok {
		t.Fatal("expected file source ref for attribute tenant")
	}

	walPath := filepath.Join(dir, "adl.jsonl")
	logRec, err := adl.New(adl.Config{
		Level:       3,
		WALPath:     walPath,
		Resource:    map[string]any{"service.name": "pdp-test"},
		Information: newInformationProvider(p),
		Logger:      logger,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logRec.Close() }()

	// A request whose principal is the file-store entity and whose context carries
	// the pulled attribute key and the standalone file attribute key.
	ctx := models.NewAttributeSet()
	ctx.AddAttribute("risk_score", 0.2)
	ctx.AddAttribute("tenant", "gemeente-amsterdam")

	ar := &authlog.AuthRecord{
		EventName:      adl.EventAccessEvaluation,
		Principal:      models.NewEntity("user", "alice", models.NewAttributeSet(models.NewAttribute("clearance", "high"))),
		Action:         models.NewEntity(models.EntityTypeName, "read", models.NewAttributeSet()),
		Resource:       models.NewEntity("service", "brp-personen", models.NewAttributeSet()),
		Decision:       true,
		RequestContext: ctx,
	}

	if err := logRec.Log(context.Background(), true, ar); err != nil {
		t.Fatalf("log: %v", err)
	}

	info := readInformation(t, walPath)
	if len(info) == 0 {
		t.Fatal("adl.core.information is empty; level-3 references did not match (A1 regression)")
	}
	// Every request-delivered key that had a registered reference must appear.
	for _, want := range []string{"user::alice", "user::alice/clearance", "risk_score", "tenant"} {
		if _, ok := info[want]; !ok {
			t.Errorf("adl.core.information missing reference for %q; got keys %v", want, keysOf(info))
		}
	}
}

// TestADLInformationConsultedFirst proves ingreep 3: when the engine reports the exact
// PIP sources it consulted (DecisionContext.consultedSources), the level-3 information
// reports precisely that set marked "basis":"consulted" - not the broader match-based
// set - even when the request delivered additional matchable keys.
func TestADLInformationConsultedFirst(t *testing.T) {
	dir := t.TempDir()
	entDir := filepath.Join(dir, "store", "entities")
	mustMkdir(t, entDir)
	// Two zone entities of a selected layer; only zone-1 is "consulted".
	mustWrite(t, filepath.Join(entDir, "zones.yaml"),
		"- type: zone\n  id: zone-1\n  attributes:\n    geometry: POINT(1 1)\n"+
			"- type: zone\n  id: zone-2\n  attributes:\n    geometry: POINT(2 2)\n")

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	p := pip.New(context.Background(), logger, pip.WithFileStore(filepath.Join(dir, "store"), true))
	refs := p.(pip.SourceReferencer)
	if _, ok := refs.SourceRef("zone::zone-1"); !ok {
		t.Fatal("expected file source ref for zone::zone-1")
	}

	prov := newInformationProvider(p)

	dc := models.NewAttributeSet()
	dc.AddAttribute("consultedSources", []string{"zone::zone-1"})

	ar := &authlog.AuthRecord{
		Principal:       models.NewEntity("user", "alice", models.NewAttributeSet()),
		Action:          models.NewEntity(models.EntityTypeName, "read", models.NewAttributeSet()),
		Resource:        models.NewEntity("zone", "zone-2", models.NewAttributeSet()), // matchable but NOT consulted
		DecisionContext: dc,
	}

	info := prov.Information(context.Background(), ar)
	if len(info) != 1 {
		t.Fatalf("expected exactly the consulted set (1 entry), got %v", keysOf(info))
	}
	ref, ok := info["zone::zone-1"].(map[string]any)
	if !ok {
		t.Fatalf("consulted source zone::zone-1 missing; got %v", keysOf(info))
	}
	if ref["basis"] != "consulted" {
		t.Errorf("expected basis=consulted, got %v", ref["basis"])
	}
	if _, present := info["zone::zone-2"]; present {
		t.Error("matched-but-not-consulted source zone::zone-2 must not appear when a consulted set is present")
	}
}

func readInformation(t *testing.T, walPath string) map[string]any {
	t.Helper()
	f, err := os.Open(walPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var rec struct {
			Attributes map[string]any `json:"attributes"`
		}
		if json.Unmarshal(sc.Bytes(), &rec) != nil {
			continue
		}
		if info, ok := rec.Attributes[adl.KeyInformation].(map[string]any); ok {
			return info
		}
	}
	return nil
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
