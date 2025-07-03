package transforming

import (
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

func (r *runner) age() any {
	v, tp := r.getValueAndType(1)
	if v == nil {
		return nil // no need for further calculations.
	}

	now := today()
	y1, m1, d1 := now.Year(), int(now.Month()), now.Day()

	var y2, m2, d2 int
	switch tp {
	case enums.StringType, enums.IntegerType, enums.UnsignedIntegerType, enums.FloatType:
		y2, m2, d2 = dmyFromNumber(v)
		if m2 == 0 {
			m2 = 1
		}
		if d2 == 0 {
			d2 = 1
		}

	default:
		ts := convert.AnyToDateTime(v)
		y2, m2, d2 = ts.Year(), int(ts.Month()), ts.Day()
	}

	if y2 < 1850 {
		return nil // unrealistic date of birth.
	}

	// age is the difference in years.
	age := y1 - y2

	if m1 < m2 || (m1 == m2 && d1 < d2) {
		age-- // minus one if the day of birth is before the current day of year.
	}

	if age < 0 {
		return nil // future date of birth.
	}
	return r.fixResultType(age)
}

func dmyFromNumber(in any) (int, int, int) {
	d := int(convert.AnyToInt64(in))
	// a numeric value must be formatted as YYYYMMDD.
	// note that in some systems one or more date-components may be zero.
	return d / 10000, d / 100 % 100, d % 100
}

var today = func() time.Time { return time.Now().Truncate(24 * time.Hour) }
