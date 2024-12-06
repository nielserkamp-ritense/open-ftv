package openfga

import (
	"fmt"
	"io"

	"github.com/goccy/go-json"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
)

func readTuples(f io.Reader) ([]*openfgav1.TupleKey, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read tuples: %w", err)
	}

	t := make(tuples, 0, 32)
	if err = json.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tuples: %w", err)
	}

	out := make([]*openfgav1.TupleKey, len(t))
	for i := range t {
		out[i] = tupleFromTuple(t[i])
	}
	return out, nil
}

func tupleFromTuple(t tuple) *openfgav1.TupleKey {
	return &openfgav1.TupleKey{
		User:     fmt.Sprintf("%s:%s", t.Subject.Type, t.Subject.ID),
		Relation: t.Predicate,
		Object:   fmt.Sprintf("%s:%s", t.Object.Type, t.Object.ID),
	}
}

type tupleEntity struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type tuple struct {
	Subject   tupleEntity `json:"subject"`
	Predicate string      `json:"predicate"`
	Object    tupleEntity `json:"object"`
}

type tuples []tuple
