package transforming

import (
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/compare"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
)

func (p *params) compare() any {
	v1, t1 := p.getValueAndType(1)
	v2, _ := p.getValueAndType(2)

	switch p.transform.CompareType {
	case enums.InList, enums.NotInList:
		return compare.Compare(compare.Params{
			Type:        t1,
			Compare:     p.transform.CompareType,
			Insensitive: p.transform.Insensitive,
			Input:       v1,
			Values:      convertList(v2),
		})

	case enums.IsLike, enums.IsNotLike, enums.MatchRegex, enums.NotMatchRegex:
		return compare.Compare(compare.Params{
			Type:    t1,
			Compare: p.transform.CompareType,
			Input:   v1,
			RX:      p.transform.GetRX(),
		})

	default:
		return compare.Compare(compare.Params{
			Type:        t1,
			Compare:     p.transform.CompareType,
			Insensitive: p.transform.Insensitive,
			Input:       v1,
			Value:       v2,
		})
	}
}

func convertList(in any) []any {
	var list []any

	switch t := in.(type) {
	case []any:
		return t

	case []string:
		for i := range t {
			list = append(list, t[i])
		}

	case []int:
		for i := range t {
			list = append(list, t[i])
		}

	case []int8:
		for i := range t {
			list = append(list, t[i])
		}

	case []int16:
		for i := range t {
			list = append(list, t[i])
		}

	case []int32:
		for i := range t {
			list = append(list, t[i])
		}

	case []int64:
		for i := range t {
			list = append(list, t[i])
		}

	case []uint:
		for i := range t {
			list = append(list, t[i])
		}

	case []uint8:
		for i := range t {
			list = append(list, t[i])
		}

	case []uint16:
		for i := range t {
			list = append(list, t[i])
		}

	case []uint32:
		for i := range t {
			list = append(list, t[i])
		}

	case []uint64:
		for i := range t {
			list = append(list, t[i])
		}

	case []float32:
		for i := range t {
			list = append(list, t[i])
		}

	case []float64:
		for i := range t {
			list = append(list, t[i])
		}

	case []time.Time:
		for i := range t {
			list = append(list, t[i])
		}

	case []*time.Time:
		for i := range t {
			list = append(list, t[i])
		}

	default:
		return []any{t}
	}

	return list
}
