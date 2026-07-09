package model

import (
	"io"
	"sort"

	"github.com/deiu/rdf2go"
	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/rdf"
)

// serializeJSONLD writes the graph as expanded JSON-LD ({"@graph": [...]}),
// grouped per subject. The serializer that ships with rdf2go drops triples
// whose object is a blank node (nested permissions/refinements), so the model
// package provides its own.
func serializeJSONLD(g *rdf2go.Graph, w io.Writer) error {
	nodes := map[string]map[string]any{}
	var order []string

	node := func(id string) map[string]any {
		n, ok := nodes[id]
		if !ok {
			n = map[string]any{"@id": id}
			nodes[id] = n
			order = append(order, id)
		}
		return n
	}

	for t := range g.IterTriples() {
		subj := jsonldID(t.Subject)
		n := node(subj)
		pred := t.Predicate.RawValue()

		if pred == rdf.Type {
			if r, ok := t.Object.(*rdf2go.Resource); ok {
				types, _ := n["@type"].([]any)
				n["@type"] = append(types, r.URI)
				continue
			}
		}

		vals, _ := n[pred].([]any)
		n[pred] = append(vals, jsonldValue(t.Object))
	}

	sort.Strings(order)
	graph := make([]any, 0, len(order))
	for _, id := range order {
		graph = append(graph, nodes[id])
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]any{"@graph": graph})
}

func jsonldID(t rdf2go.Term) string {
	if b, ok := t.(*rdf2go.BlankNode); ok {
		return "_:" + b.ID
	}
	return t.RawValue()
}

func jsonldValue(t rdf2go.Term) any {
	switch x := t.(type) {
	case *rdf2go.Resource:
		return map[string]any{"@id": x.URI}
	case *rdf2go.BlankNode:
		return map[string]any{"@id": "_:" + x.ID}
	case *rdf2go.Literal:
		v := map[string]any{"@value": x.Value}
		if x.Language != "" {
			v["@language"] = x.Language
		}
		if x.Datatype != nil {
			v["@type"] = x.Datatype.RawValue()
		}
		return v
	default:
		return map[string]any{"@value": t.RawValue()}
	}
}
