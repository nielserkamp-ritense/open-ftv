package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertParams(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
		want map[string]any
	}{
		{
			name: "empty",
			want: map[string]any{},
		},
		{
			name: "json",
			in:   `{"leeftijd": 18, "inkomen":63000.00, "kinderen": false}`,
			want: map[string]any{
				"leeftijd": 18.0,
				"inkomen":  63000.0,
				"kinderen": false,
			},
		},
		{
			name: "bad text",
			in:   "leeftijd is 18",
			want: map[string]any{},
		},
		{
			name: "single param",
			in:   "leeftijd=18",
			want: map[string]any{
				"leeftijd": "18",
			},
		},
		{
			name: "few params",
			in:   "leeftijd=18,kinderen=false,spaargeld=10000",
			want: map[string]any{
				"leeftijd":  "18",
				"kinderen":  "false",
				"spaargeld": "10000",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := convertParams(tc.in)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		query       map[string]string
		body        map[string]any
		primary     string
		wantErr     bool
		wantFilter  string
		wantMatcher string
		wantParams  map[string]any
		wantRemain  map[string]any
	}{
		{
			name:        "both empty",
			wantFilter:  "{allOf:[]}",
			wantMatcher: "{always:true}",
		},
		{
			name:    "fields in query",
			primary: "t1",
			query: map[string]string{
				"@fields": "f1,f2,b*",
			},
			wantFilter:  "{allOf:[]}",
			wantMatcher: "{exact:[f1,f2],regex:[^b.*$]}",
		},
		{
			name:    "params in query",
			primary: "persoon",
			query: map[string]string{
				"@params": "persoon.leeftijd=18,inkomen=30000",
			},
			wantFilter:  "{allOf:[]}",
			wantMatcher: "{always:true}",
			wantParams: map[string]any{
				"persoon.leeftijd": "18",
				"inkomen":          "30000",
			},
		},
		{
			name:    "filter in query",
			primary: "persoon",
			query: map[string]string{
				"@filter":        `volwassen == 1, persoon.bsn not nil`,
				"adres.postcode": "1111ZZ",
				"gemeente":       "xyz",
			},
			wantFilter:  "{allOf:[{allOf:[{level=primary,field=volwassen,compare-type=IsEqual,value=1} {level=primary,table=persoon,field=bsn,compare-type=Exists}]} {level=join,join=adres,field=postcode,compare-type=IsEqual,value=1111ZZ} {level=primary,field=gemeente,compare-type=IsEqual,value=xyz}]}",
			wantMatcher: "{always:true}",
		},
		{
			name:    "fields in body",
			primary: "t1",
			body: map[string]any{
				"@fields": "f1,f2,b*",
				"x":       123,
			},
			wantFilter:  "{allOf:[]}",
			wantMatcher: "{exact:[f1,f2],regex:[^b.*$]}",
			wantRemain: map[string]any{
				"x": 123,
			},
		},
		{
			name:    "params in body",
			primary: "t1",
			body: map[string]any{
				"@params": map[string]any{"leeftijd": 18, "inkomen": 30000},
				"x":       123,
			},
			wantFilter:  "{allOf:[]}",
			wantMatcher: "{always:true}",
			wantParams: map[string]any{
				"leeftijd": 18,
				"inkomen":  30000,
			},
			wantRemain: map[string]any{
				"x": 123,
			},
		},
		{
			name:    "filter in body",
			primary: "persoon",
			body: map[string]any{
				"@filter": map[string]any{"adres.postcode": "1111ZZ", "@filter": "leeftijd >= 18", "volwassen": true},
			},
			wantFilter:  "{allOf:[{level=primary,field=leeftijd,compare-type=IsGreaterOrEqual,value=18} {level=join,join=adres,field=postcode,compare-type=IsEqual,value=1111ZZ} {level=primary,field=volwassen,compare-type=IsEqual,value=true}]}",
			wantMatcher: "{always:true}",
			wantRemain:  map[string]any{},
		},
		{
			name:    "fields in both",
			primary: "t1",
			query: map[string]string{
				"@fields": "f1,f2,b*",
			},
			body: map[string]any{
				"@fields": "a*,f6,f4",
			},
			wantFilter:  "{allOf:[]}",
			wantMatcher: "{exact:[f1,f2,f6,f4],regex:[^b.*$,^a.*$]}",
			wantRemain:  map[string]any{},
		},
		{
			name:    "params in both",
			primary: "t1",
			query: map[string]string{
				"@params": "leeftijd=18",
			},
			body: map[string]any{
				"@params": map[string]any{"inkomen": 30000},
			},
			wantFilter:  "{allOf:[]}",
			wantMatcher: "{always:true}",
			wantParams: map[string]any{
				"leeftijd": "18",
				"inkomen":  30000,
			},
			wantRemain: map[string]any{},
		},
		{
			name:    "filter in both",
			primary: "t1",
			query: map[string]string{
				"@filter": "postcode=1111ZZ,volwassen=1,gemeente=xyz",
			},
			body: map[string]any{
				"@filter": "adres.postcode=2222YY,volwassen=false",
			},
			wantFilter:  "{allOf:[{level=join,join=adres,field=postcode,compare-type=IsEqual,value=2222YY} {level=primary,field=volwassen,compare-type=IsEqual,value=false}]}",
			wantMatcher: "{always:true}",
			wantRemain:  map[string]any{},
		},
		{
			name:    "all",
			primary: "t1",
			query: map[string]string{
				"@fields": "f1,f2,b*",
				"@filter": "postcode=1111ZZ,volwassen=1,gemeente=xyz",
				"@params": "leeftijd=18",
			},
			body: map[string]any{
				"oops":    map[string]any{"x": 123},
				"@params": map[string]any{"inkomen": 30000},
				"@filter": "postcode=2222YY,volwassen=false",
				"@fields": "a*,f6,f4",
				"x":       "y",
			},
			wantFilter:  "{allOf:[{level=primary,field=postcode,compare-type=IsEqual,value=2222YY} {level=primary,field=volwassen,compare-type=IsEqual,value=false}]}",
			wantMatcher: "{exact:[f1,f2,f6,f4],regex:[^b.*$,^a.*$]}",
			wantParams: map[string]any{
				"leeftijd": "18",
				"inkomen":  30000,
			},
			wantRemain: map[string]any{
				"oops": map[string]any{"x": 123},
				"x":    "y",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, err := New(tc.query, tc.body, tc.primary)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, ctx)
			} else {
				require.NoError(t, err)
				require.NotNil(t, ctx)

				assert.Equal(t, tc.wantFilter, ctx.Filter.String())
				assert.Equal(t, tc.wantMatcher, ctx.Matcher.String())
				assert.EqualValues(t, tc.wantParams, ctx.Params)
				assert.EqualValues(t, tc.wantRemain, ctx.RemainingBody)
			}
		})
	}
}
