package transforming

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

// Execute executes the given transformation using the given table data.
//
// The result of this function is the given table data, extended with the result of the transformation.
func Execute(data map[string]any, qualifiers map[string]string, qualified bool, transform *schema.Transformation, params map[string]any) (map[string]any, map[string]string) {
	p := &runner{qualified: qualified, data: data, qualifiers: qualifiers, transform: transform, params: params}
	p.addResult(p.run())
	return p.data, p.qualifiers
}

func (r *runner) run() any {
	switch r.transform.TransformationType {
	case enums.TransformCompare:
		return r.compare()
	case enums.TransformConvert:
		return r.convert()
	case enums.TransformAge:
		return r.age()
	default:
		return nil
	}
}

func (r *runner) addResult(v any) {
	// for the value nil, we must suppress the output!
	if v != nil {
		if r.qualified {
			r.data[r.transform.FQID()] = v
			r.qualifiers[r.transform.FQID()] = r.transform.FQID()
		} else {
			r.data[r.transform.ID] = v
			r.qualifiers[r.transform.ID] = r.transform.FQID()
		}
	}
}

type runner struct {
	qualified  bool
	data       map[string]any
	qualifiers map[string]string
	transform  *schema.Transformation
	params     map[string]any
}
