package transforming

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

// Execute executes the given transformation using the given table data.
//
// The result of this function is the given table data, extended with the result of the transformation.
func Execute(rec map[string]any, qualified bool, transform *schema.Transformation, params map[string]any) map[string]any {
	p := &runner{qualified: qualified, rec: rec, transform: transform, params: params}
	p.addResult(p.run())
	return p.rec
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
			r.rec[r.transform.FQID()] = v
		} else {
			r.rec[r.transform.ID] = v
		}
	}
}

type runner struct {
	qualified bool
	rec       map[string]any
	transform *schema.Transformation
	params    map[string]any
}
