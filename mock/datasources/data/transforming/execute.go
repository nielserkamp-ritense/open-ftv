package transforming

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

// Execute executes the given transformation using the given table data.
//
// The result of this function is the given table data, extended with the result of the transformation.
func Execute(rec map[string]any, qualified bool, transform *schema.Transformation) map[string]any {
	p := &params{qualified: qualified, rec: rec, transform: transform}
	p.addResult(p.run())
	return p.rec
}

func (p *params) run() any {
	switch p.transform.TransformationType {
	case enums.TransformCompare:
		return p.compare()
	case enums.TransformConvert:
		return p.convert()
	case enums.TransformAge:
		return p.age()
	default:
		return nil
	}
}

func (p *params) addResult(v any) {
	if v != nil {
		if p.qualified {
			p.rec[p.transform.FQID()] = v
		} else {
			p.rec[p.transform.ID] = v
		}
	}
}

type params struct {
	qualified bool
	rec       map[string]any
	transform *schema.Transformation
}
