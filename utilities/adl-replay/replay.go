// Command adl-replay is the ADL level-4 replay verifier (ingreep 4).
//
// It reads a level-4 authorization-decision-log record, reconstructs the PDP state it
// describes - the policies (resolved by content hash from a PAP, a local policy
// directory, or the manager's /v1/policy-hash endpoint), the PIP values, and the
// engine and request-mappings from adl.core.configuration - re-evaluates
// body.adl.core.request through a freshly constructed pdp controller, and compares the
// re-evaluated decision and obligations with body.adl.core.response. It reports
// PASS/FAIL with a diff and sets its exit code accordingly, giving the standard's
// "full, guaranteed replay" a concrete, verifiable realisation.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	odrl_geo "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl-geo"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

// Options configures a replay.
type Options struct {
	// PolicyDir loads every policy file in a directory into the reconstructed PAP.
	PolicyDir string
	// ManagerURL resolves each adl.core.policies hash via GET {url}/v1/policy-hash/{hash}.
	ManagerURL string
	// PIPStore is the PIP attribute file-store used to reconstruct information values.
	PIPStore string
	// Mappings overrides the request-mappings; when empty they are taken from the record.
	Mappings string
	// Logger receives diagnostics; defaults to a discarding logger.
	Logger *slog.Logger
}

// Result is the outcome of a replay.
type Result struct {
	Pass             bool
	ExpectedDecision bool
	ActualDecision   bool
	ResolvedHashes   map[string]bool // policy hash -> resolvable in the reconstructed PAP.
	Diff             string
}

// Replay reconstructs the PDP state described by rec and re-evaluates the request.
func Replay(ctx context.Context, rec *Record, opts Options) (*Result, error) {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	conf := mapAt(rec.Body, "adl.core.configuration")
	engine := strings.ToLower(str(conf["engine"]))
	mappings := opts.Mappings
	if mappings == "" {
		mappings = str(conf["request_mappings"])
	}

	// Reconstruct the PAP (policies) and PIP (information values).
	ap, err := reconstructPAP(ctx, engine, rec, opts)
	if err != nil {
		return nil, err
	}
	ip := pip.New(ctx, opts.Logger, pipStoreOption(opts.PIPStore)...)

	// Verify every logged policy hash resolves to a policy in the reconstructed PAP.
	resolved := map[string]bool{}
	for _, h := range policyHashes(rec) {
		_, e := ap.ReadByHash(h)
		resolved[h] = e == nil
	}

	controller, err := buildController(ctx, engine, mappings, ap, ip, opts.Logger)
	if err != nil {
		return nil, err
	}

	parc, err := parcFromRequest(mapAt(rec.Body, "adl.core.request"))
	if err != nil {
		return nil, err
	}

	resp, err := controller.Authorize("replay", parc)
	if err != nil {
		return nil, fmt.Errorf("re-evaluation failed: %w", err)
	}

	return compare(rec, resp, resolved), nil
}

// reconstructPAP builds an in-memory PAP holding the policies for the decision, either
// from a local policy directory, or by fetching each logged policy hash from the
// manager's content-addressable endpoint.
func reconstructPAP(ctx context.Context, engine string, rec *Record, opts Options) (pap.PAP, error) {
	lang := papLanguage(engine)
	ap := pap.New(ctx, opts.Logger, pap.WithLanguage(lang))

	switch {
	case opts.PolicyDir != "":
		entries, err := os.ReadDir(opts.PolicyDir)
		if err != nil {
			return nil, fmt.Errorf("read policy dir: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.HasSuffix(e.Name(), ".meta") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(opts.PolicyDir, e.Name()))
			if err != nil {
				return nil, err
			}
			id := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
			pol, err := pap.NewPolicyFromData(id, lang, "", "urn:"+id, bytes.NewReader(data))
			if err != nil {
				return nil, err
			}
			if _, err := ap.Create(pol); err != nil {
				return nil, fmt.Errorf("store policy %s: %w", id, err)
			}
		}
	case opts.ManagerURL != "":
		for _, h := range policyHashes(rec) {
			data, err := fetchPolicyByHash(ctx, opts.ManagerURL, h)
			if err != nil {
				return nil, err
			}
			pol, err := pap.NewPolicyFromData(h, lang, "", "urn:hash:"+h, bytes.NewReader(data))
			if err != nil {
				return nil, err
			}
			if _, err := ap.Create(pol); err != nil {
				return nil, fmt.Errorf("store policy %s: %w", h, err)
			}
		}
	default:
		return nil, fmt.Errorf("no policy source: set -policy-dir or -manager-url")
	}
	return ap, nil
}

// fetchPolicyByHash retrieves the exact policy source for a hash from the manager.
func fetchPolicyByHash(ctx context.Context, base, hash string) ([]byte, error) {
	url := strings.TrimRight(base, "/") + "/v1/policy-hash/" + hash
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manager returned %d resolving hash %s", resp.StatusCode, hash)
	}
	return io.ReadAll(resp.Body)
}

// buildController constructs the reference PDP controller for the recorded engine.
func buildController(ctx context.Context, engine, mappings string, ap pap.PAP, ip pip.PIP, logger *slog.Logger) (pdp.Controller, error) {
	opts := []pdp.Option{pdp.WithContext(ctx), pdp.WithLogger(logger), pdp.WithPAP(ap), pdp.WithPIP(ip)}
	if mappings != "" {
		opts = append(opts, pdp.WithMappings(mapping.MappingsFromConfig(mappings)...))
	}

	switch engine {
	case "odrl", "odrl-geo", "odrlgeo", "":
		return odrl_geo.NewController(opts...), nil
	default:
		return nil, fmt.Errorf("replay does not support engine %q (only the odrl-geo reference engine)", engine)
	}
}

// compare checks the re-evaluated decision and obligations against the recorded response.
func compare(rec *Record, resp *models.Response, resolved map[string]bool) *Result {
	expResp := mapAt(rec.Body, "adl.core.response")
	expDecision, _ := expResp["decision"].(bool)

	expOblig := canonical(obligationsFromResponse(expResp))
	actOblig := canonical(resp.Attributes["obligations"])

	res := &Result{
		ExpectedDecision: expDecision,
		ActualDecision:   resp.Allowed,
		ResolvedHashes:   resolved,
	}

	var diffs []string
	if expDecision != resp.Allowed {
		diffs = append(diffs, fmt.Sprintf("decision: expected %v, replayed %v", expDecision, resp.Allowed))
	}
	if !reflect.DeepEqual(expOblig, actOblig) {
		diffs = append(diffs, fmt.Sprintf("obligations:\n  expected: %s\n  replayed: %s", jsonString(expOblig), jsonString(actOblig)))
	}
	for h, ok := range resolved {
		if !ok {
			diffs = append(diffs, fmt.Sprintf("policy hash %s did not resolve in the reconstructed PAP", h))
		}
	}

	res.Pass = len(diffs) == 0
	res.Diff = strings.Join(diffs, "\n")
	return res
}

// --- record -> evaluation input helpers -------------------------------------

func parcFromRequest(req map[string]any) (*models.PARC, error) {
	if len(req) == 0 {
		return nil, fmt.Errorf("record has no body.adl.core.request")
	}

	subj := mapOf(req["subject"])
	act := mapOf(req["action"])
	res := mapOf(req["resource"])
	if subj == nil || act == nil || res == nil {
		return nil, fmt.Errorf("request missing subject/action/resource")
	}

	principal := models.NewEntity(str(subj["type"]), str(subj["id"]), models.NewAttributeSet(mapOf(subj["properties"])))
	action := models.NewEntity(models.EntityTypeName, str(act["name"]), models.NewAttributeSet(mapOf(act["properties"])))
	resource := models.NewEntity(str(res["type"]), str(res["id"]), models.NewAttributeSet(mapOf(res["properties"])))
	rctx := models.NewAttributeSet(mapOf(req["context"]))

	return &models.PARC{Principal: principal, Action: action, Resource: resource, Context: rctx}, nil
}

// policyHashes returns the hash values from attributes["adl.core.policies"] ({key: hash}).
func policyHashes(rec *Record) []string {
	pol := mapAt(rec.Attributes, "adl.core.policies")
	var out []string
	for _, v := range pol {
		if h := str(v); h != "" {
			out = append(out, h)
		}
	}
	sort.Strings(out)
	return out
}

// obligationsFromResponse extracts the obligations array from a recorded response's
// context ({decision, context:{obligations:[...]}}).
func obligationsFromResponse(resp map[string]any) any {
	c := mapOf(resp["context"])
	if c == nil {
		return nil
	}
	return c["obligations"]
}

// canonical normalises a value through JSON so int/int64/float compare equal and
// obligation lists compare order-independently.
func canonical(v any) any {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	if json.Unmarshal(b, &out) != nil {
		return v
	}
	sortLists(out)
	return out
}

// sortLists sorts every []any it finds by the JSON encoding of its elements, so two
// obligation lists that differ only in order are treated as equal.
func sortLists(v any) {
	switch t := v.(type) {
	case []any:
		for _, e := range t {
			sortLists(e)
		}
		sort.Slice(t, func(i, j int) bool { return jsonString(t[i]) < jsonString(t[j]) })
	case map[string]any:
		for _, e := range t {
			sortLists(e)
		}
	}
}

func pipStoreOption(store string) []pip.Option {
	if store == "" {
		return nil
	}
	return []pip.Option{pip.WithFileStore(store, true)}
}

func papLanguage(engine string) string {
	if engine == "" {
		return "odrl"
	}
	if engine == "odrl-geo" || engine == "odrlgeo" {
		return "odrl"
	}
	return engine
}

func mapAt(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	return mapOf(m[key])
}

func mapOf(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func jsonString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}
