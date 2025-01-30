package openfga

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/goccy/go-json"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
)

func readTuples(f io.Reader) ([]*openfgav1.TupleKey, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read tuples: %w", err)
	}

	t := make(basicTuples, 0, 32)
	if err = json.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tuples: %w", err)
	}

	out := make([]*openfgav1.TupleKey, len(t))
	for i := range t {
		out[i] = tupleKeyFromBasicTuple(t[i])
	}
	return out, nil
}

func tupleKeyFromBasicTuple(t basicTuple) *openfgav1.TupleKey {
	return &openfgav1.TupleKey{
		User:     t.Subject.key(),
		Relation: normalize(t.Predicate),
		Object:   t.Object.key(),
	}
}

func (r basicEntity) key() string {
	return fmt.Sprintf("%s:%s", r.Type, normalize(r.ID))
}

func normalize(id string) string {
	if strings.ContainsAny(id, ":#@") {
		// some characters are prohibited for use in identifiers.
		return base64.StdEncoding.EncodeToString([]byte(id))
	}
	return id
}

type basicEntity struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type basicTuple struct {
	Subject   basicEntity `json:"subject"`
	Predicate string      `json:"predicate"`
	Object    basicEntity `json:"object"`
}

type basicTuples []basicTuple
