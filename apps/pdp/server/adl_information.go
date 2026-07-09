package server

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

// newInformationProvider bridges the PIP's source references to the ADL level-3 hook
// (adl.core.information).
//
// Two bases are distinguished (ingreep 3):
//   - "consulted": the policy engine reported the exact set of PIP sources it read for
//     this decision (DecisionContext.consultedSources). This is the accurate,
//     causally-consulted set and is preferred whenever present.
//   - "matched": fallback for black-box engines (OPA/Cedar/OpenFGA/Cerbos) that cannot
//     report their reads. It reports every registered source reference whose key matches
//     something the request delivered (the entity UIDs, their attributes keyed
//     "<uid>/<attr>", and the context attribute keys). This over-reports (a value that
//     was available but perhaps unused) rather than under-reports, so it never hides a
//     consulted source, but it does not prove causal consultation.
func newInformationProvider(ip any) adl.InformationProvider {
	refs, ok := ip.(pip.SourceReferencer)
	if !ok {
		return nil
	}
	return &informationProvider{refs: refs}
}

type informationProvider struct{ refs pip.SourceReferencer }

// Information implements the adl.InformationProvider interface.
func (p *informationProvider) Information(_ context.Context, ar *authlog.AuthRecord) map[string]any {
	if ar == nil {
		return nil
	}

	// Consulted-first: when the engine reported exactly which PIP sources it consulted,
	// report that exact set marked "consulted".
	if consulted := consultedKeys(ar); len(consulted) > 0 {
		out := map[string]any{}
		for _, key := range consulted {
			if _, seen := out[key]; seen {
				continue
			}
			if ref, ok := p.refs.SourceRef(key); ok {
				out[key] = sourceRefMap(ref, "consulted")
			}
		}
		if len(out) > 0 {
			return out
		}
		// consulted keys carried no registered source references; fall through to matched.
	}

	out := map[string]any{}
	add := func(key string) {
		if key == "" {
			return
		}
		if _, seen := out[key]; seen {
			return
		}
		if ref, ok := p.refs.SourceRef(key); ok {
			out[key] = sourceRefMap(ref, "matched")
		}
	}

	for _, e := range []models.Entity{ar.Principal, ar.Action, ar.Resource} {
		if e == nil || e.UID() == "" {
			continue
		}
		// The entity itself, plus each of its attributes under "<uid>/<attr>".
		add(e.UID())
		if attrs := e.Attributes(); attrs != nil {
			attrs.IterateAttributes(func(a models.Attribute) {
				add(e.UID() + "/" + a.Key())
			})
		}
	}

	// Attribute keys delivered in the request context (a PDP request keys attributes
	// by their bare name, which is how the PIP registers pulled/file attributes).
	if ar.RequestContext != nil {
		ar.RequestContext.IterateAttributes(func(a models.Attribute) {
			add(a.Key())
		})
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

// consultedKeys returns the exact set of PIP source keys the engine reported consulting
// for this decision, taken from DecisionContext.consultedSources. Returns nil when the
// engine did not report a consulted set (black-box engines).
func consultedKeys(ar *authlog.AuthRecord) []string {
	if ar.DecisionContext == nil {
		return nil
	}
	switch v := ar.DecisionContext.GetAttributeValue("consultedSources").(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, e := range v {
			if s, ok := e.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// sourceRefMap renders a PIP source reference as an ADL source reference (never a raw
// payload), tagged with the basis on which it is reported ("consulted" or "matched").
func sourceRefMap(ref pip.SourceRef, basis string) map[string]any {
	m := map[string]any{"span_id": ref.SpanID, "basis": basis}
	if ref.TraceID != "" {
		m["trace_id"] = ref.TraceID
	}
	if ref.WARCFile != "" {
		m["warc_file"] = ref.WARCFile
	}
	if ref.Version != "" {
		m["version"] = ref.Version
	}
	if ref.Sequence != "" {
		m["sequence"] = ref.Sequence
	}
	return m
}
