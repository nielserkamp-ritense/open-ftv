package query

import (
	"fmt"
	"sort"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
)

// Bucket is a period granularity for aggregation.
type Bucket string

// The supported period buckets.
const (
	BucketDay   Bucket = "day"
	BucketWeek  Bucket = "week"
	BucketMonth Bucket = "month"
)

// DefaultK is the default k-anonymity threshold: aggregate buckets with fewer
// than this many decisions are suppressed from the output.
const DefaultK = 5

// AggregateOptions configures the aggregation.
type AggregateOptions struct {
	// Bucket is the period granularity (day/week/month). Defaults to day.
	Bucket Bucket
	// K is the k-anonymity threshold. Buckets with a total decision count below
	// K are suppressed. Values <= 0 fall back to DefaultK.
	K int
}

// Key identifies one aggregate bucket: a consumer, a purpose and a period.
type Key struct {
	Afnemer string `json:"afnemer"`
	Doel    string `json:"doel"`
	Periode string `json:"periode"`
}

// Count is the aggregate for one Key: permit/deny/total decision counts. It
// carries no personal data - only the grouping key and the tallies.
type Count struct {
	Key
	Permit int `json:"permit"`
	Deny   int `json:"deny"`
	Total  int `json:"total"`
}

// Aggregate groups the records per (afnemer, doel, period-bucket) and counts
// permit/deny decisions. It is a data-minimisation / anonymisation step as
// endorsed by the ADL standard (§5a): subject ids are dropped and only tallies
// remain. Buckets whose total falls below the k-anonymity threshold are
// suppressed entirely, so no small (potentially re-identifiable) group is
// exposed. The result is sorted deterministically.
func Aggregate(records []adl.Record, opts AggregateOptions) []Count {
	bucket := opts.Bucket
	if bucket == "" {
		bucket = BucketDay
	}
	k := opts.K
	if k <= 0 {
		k = DefaultK
	}

	tally := make(map[Key]*Count)
	for i := range records {
		r := &records[i]
		decision, ok := Decision(r)
		if !ok {
			// Error records carry no decision; they are not permit/deny and are
			// excluded from usage statistics.
			continue
		}
		key := Key{
			Afnemer: Afnemer(r),
			Doel:    Doel(r),
			Periode: periodLabel(recordTime(r), bucket),
		}
		c := tally[key]
		if c == nil {
			c = &Count{Key: key}
			tally[key] = c
		}
		if decision {
			c.Permit++
		} else {
			c.Deny++
		}
		c.Total++
	}

	out := make([]Count, 0, len(tally))
	for _, c := range tally {
		if c.Total < k {
			continue // k-anonymity suppression.
		}
		out = append(out, *c)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Afnemer != out[j].Afnemer {
			return out[i].Afnemer < out[j].Afnemer
		}
		if out[i].Doel != out[j].Doel {
			return out[i].Doel < out[j].Doel
		}
		return out[i].Periode < out[j].Periode
	})
	return out
}

// periodLabel renders the period bucket label for a timestamp:
//   - day:   "2006-01-02"
//   - week:  "2006-Www" (ISO-8601 year and week)
//   - month: "2006-01"
func periodLabel(t time.Time, b Bucket) string {
	switch b {
	case BucketWeek:
		y, w := t.ISOWeek()
		return isoWeekLabel(y, w)
	case BucketMonth:
		return t.Format("2006-01")
	default:
		return t.Format("2006-01-02")
	}
}

func isoWeekLabel(year, week int) string {
	return fmt.Sprintf("%04d-W%02d", year, week)
}
