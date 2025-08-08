package bundles

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func addPolicies(b *Bundle) {
	p1, _ := models.NewPolicyFromData("myP1", "opa", "", "", bytes.NewBufferString("haha"))
	p2, _ := models.NewPolicyFromData("myP2", "cedar", "", "", bytes.NewBufferString("haha"))
	p3, _ := models.NewPolicyFromData("myP3", "opa/rego", "", "", bytes.NewBufferString("haha"))
	p4, _ := models.NewPolicyFromData("myP4", "cerbos", "", "", bytes.NewBufferString("haha"))
	p5, _ := models.NewPolicyFromData("myP5", "openFGA", "", "", bytes.NewBufferString("haha"))
	p6, _ := models.NewPolicyFromData("p6", "cedar", "", "", bytes.NewBufferString("haha"))
	p7, _ := models.NewPolicyFromData("p7", "cel", "", "", bytes.NewBufferString("haha"))

	p1.WithTags("x", "y", "z")
	p2.WithTags("x")
	p3.WithTags("a", "b")
	p4.WithTags("y", "z")
	p5.WithTags("x", "y")
	p6.WithTags("a", "b", "x")
	p7.WithTags("a", "c", "z")

	policies := []*models.Policy{p1, p2, p3, p4, p5, p6, p7}

	for i := range policies {
		b.AddPolicy(policies[i])
	}
}

func addAttributes(b *Bundle) {
	a1 := models.NewAttribute("myA1", 1)
	a2 := models.NewAttribute("myA2", 1)
	a3 := models.NewAttribute("myA3", 1)
	a4 := models.NewAttribute("myA4", 1)
	a5 := models.NewAttribute("a5", 1)
	a6 := models.NewAttribute("a6", 1)
	a7 := models.NewAttribute("a7", 1)
	a8 := models.NewAttribute("a8", 1)

	a1.WithTags("x", "y", "z")
	a2.WithTags("x", "y", "z")
	a3.WithTags("x", "y")
	a4.WithTags("x", "y")
	a5.WithTags("a", "b", "c")
	a6.WithTags("a", "b", "c")
	a7.WithTags("b", "c")
	a8.WithTags("a", "c")

	attributes := []*models.Attribute{a1, a2, a3, a4, a5, a6, a7, a8}

	for i := range attributes {
		b.AddAttribute(attributes[i])
	}
}

func addEntities(b *Bundle) {
	e1 := models.NewEntity("user", "e1", models.NewAttributeSet(models.NewAttribute("myA1", 1)))
	e2 := models.NewEntity("user", "e2", nil)
	e3 := models.NewEntity("user", "e3", models.NewAttributeSet(models.NewAttribute("myA3", 1)))
	e4 := models.NewEntity("user", "e4", nil)
	e5 := models.NewEntity("user", "e5", models.NewAttributeSet(models.NewAttribute("myA1", 1)))
	e6 := models.NewEntity("user", "e6", nil)
	e7 := models.NewEntity("user", "e7", models.NewAttributeSet(models.NewAttribute("myA3", 1)))
	e8 := models.NewEntity("user", "e8", nil)

	e1.WithTags("x", "y", "z")
	e2.WithTags("x", "y", "z")
	e3.WithTags("x", "y")
	e4.WithTags("x", "y")
	e5.WithTags("a", "b", "c")
	e6.WithTags("a", "b", "c")
	e7.WithTags("b", "c")
	e8.WithTags("a", "c")

	entities := []*models.Entity{e1, e2, e3, e4, e5, e6, e7, e8}

	for i := range entities {
		b.AddEntity(entities[i])
	}
}

func addRelations(b *Bundle) {
	s1 := models.NewEntity("user", "s1", nil)
	s2 := models.NewEntity("user", "s2", nil)
	p1 := models.NewEntity("action", "POST", nil)
	p2 := models.NewEntity("action", "GET", nil)
	o1 := models.NewEntity("resource", "o1", nil)
	o2 := models.NewEntity("resource", "o2", nil)

	r1 := models.NewRelation(s1, p1, o1)
	r2 := models.NewRelation(s1, p1, o2)
	r3 := models.NewRelation(s1, p2, o1)
	r4 := models.NewRelation(s1, p2, o2)
	r5 := models.NewRelation(s2, p1, o1)
	r6 := models.NewRelation(s2, p1, o2)
	r7 := models.NewRelation(s2, p2, o1)
	r8 := models.NewRelation(s2, p2, o2)

	r1.WithTags("x", "y", "z")
	r2.WithTags("x", "y", "z")
	r3.WithTags("x", "y")
	r4.WithTags("x", "y")
	r5.WithTags("a", "b", "c")
	r6.WithTags("a", "b", "c")
	r7.WithTags("b", "c")
	r8.WithTags("a", "c")

	relations := []*models.Relation{r1, r2, r3, r4, r5, r6, r7, r8}

	for i := range relations {
		b.AddRelation(relations[i])
	}
}

func TestNewBundle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		version  uint64
		language string
		tags     []string
	}{
		{
			name:     "opa",
			version:  1,
			language: "opa/rego",
			tags:     []string{"x", "y", "z"},
		},
		{
			name:     "cedar",
			version:  2,
			language: "cedar",
			tags:     []string{"a", "b"},
		},
		{
			name:     "openfga",
			version:  3,
			language: "OpenFGA",
			tags:     []string{"x", "a"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewBundle(tc.version, tc.language, tc.tags...)
			require.NotNil(t, got)
			assert.Equal(t, tc.version, got.Version)
			assert.Equal(t, tc.language, got.Language)
			assert.EqualValues(t, tc.tags, got.tags)

			assert.NotNil(t, got.Policies)
			assert.NotNil(t, got.Attributes)
			assert.NotNil(t, got.Entities)
			assert.NotNil(t, got.Relations)
		})
	}
}

func TestBundle_AddPolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		language string
		tags     []string
		want     int
	}{
		{name: "rego x", language: "rego", tags: []string{"x"}, want: 1},
		{name: "rego a+x", language: "rego", tags: []string{"a", "x"}, want: 2},
		{name: "rego a+b+c", language: "rego", tags: []string{"a", "b", "c"}, want: 1},
		{name: "cedar x", language: "cedar", tags: []string{"x"}, want: 2},
		{name: "cerbos z", language: "Cerbos", tags: []string{"z"}, want: 2},
		{name: "openfga a+b+c+x+y+z", language: "OpenFGA", tags: []string{"a", "b", "c", "x", "y", "z"}, want: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewBundle(1, tc.language, tc.tags...)
			require.NotNil(t, got)

			addPolicies(got)

			assert.NotNil(t, got.Policies)
			assert.Equal(t, tc.want, len(got.Policies))

			assert.NotNil(t, got.Attributes)
			assert.Zero(t, len(got.Attributes))

			assert.NotNil(t, got.Entities)
			assert.Zero(t, len(got.Entities))

			assert.NotNil(t, got.Relations)
			assert.Zero(t, len(got.Relations))
		})
	}
}

func TestBundle_AddAttribute(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		tags []string
		want int
	}{
		{name: "none", tags: []string{}, want: 0},
		{name: "x", tags: []string{"x"}, want: 4},
		{name: "b", tags: []string{"b"}, want: 3},
		{name: "a+y", tags: []string{"a", "y"}, want: 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewBundle(1, "?", tc.tags...)
			require.NotNil(t, got)

			addAttributes(got)

			assert.NotNil(t, got.Policies)
			assert.Zero(t, len(got.Policies))

			assert.NotNil(t, got.Attributes)
			assert.Equal(t, tc.want, len(got.Attributes))

			assert.NotNil(t, got.Entities)
			assert.Zero(t, len(got.Entities))

			assert.NotNil(t, got.Relations)
			assert.Zero(t, len(got.Relations))
		})
	}
}

func TestBundle_AddEntity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		tags []string
		want int
	}{
		{name: "none", tags: []string{}, want: 0},
		{name: "x", tags: []string{"x"}, want: 4},
		{name: "b", tags: []string{"b"}, want: 3},
		{name: "a+y", tags: []string{"a", "y"}, want: 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewBundle(1, "?", tc.tags...)
			require.NotNil(t, got)

			addEntities(got)

			assert.NotNil(t, got.Policies)
			assert.Zero(t, len(got.Policies))

			assert.NotNil(t, got.Attributes)
			assert.Zero(t, len(got.Attributes))

			assert.NotNil(t, got.Entities)
			assert.Equal(t, tc.want, len(got.Entities))

			assert.NotNil(t, got.Relations)
			assert.Zero(t, len(got.Relations))
		})
	}
}

func TestBundle_AddRelation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		tags []string
		want int
	}{
		{name: "none", tags: []string{}, want: 0},
		{name: "x", tags: []string{"x"}, want: 4},
		{name: "b", tags: []string{"b"}, want: 3},
		{name: "a+y", tags: []string{"a", "y"}, want: 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewBundle(1, "?", tc.tags...)
			require.NotNil(t, got)

			addRelations(got)

			assert.NotNil(t, got.Policies)
			assert.Zero(t, len(got.Policies))

			assert.NotNil(t, got.Attributes)
			assert.Zero(t, len(got.Attributes))

			assert.NotNil(t, got.Entities)
			assert.Zero(t, len(got.Entities))

			assert.NotNil(t, got.Relations)
			assert.Equal(t, tc.want, len(got.Relations))
		})
	}
}

func TestBundle_Transfer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		version  uint64
		language string
		tags     []string
		compress CompressionType
		headers  map[string][]string
		wantErr  bool
		wantErr2 bool
	}{
		{
			name:     "bad compression type",
			version:  1,
			language: "OPA",
			tags:     []string{"x"},
			compress: 0,
			wantErr:  true,
		},
		{
			name:     "bad compression header (1)",
			version:  1,
			language: "OPA",
			tags:     []string{"x"},
			compress: CompressBZ2,
			headers:  map[string][]string{},
			wantErr2: true,
		},
		{
			name:     "bad compression header (2)",
			version:  1,
			language: "OPA",
			tags:     []string{"x"},
			compress: CompressGZ,
			headers:  map[string][]string{"Content-Encoding": {"bz2"}},
			wantErr2: true,
		},
		{
			name:     "good gzip - no header",
			version:  1,
			language: "OPA",
			tags:     []string{"x"},
			compress: CompressGZ,
			headers:  map[string][]string{},
		},
		{
			name:     "good gzip - with header",
			version:  1,
			language: "OPA",
			tags:     []string{"x"},
			compress: CompressGZ,
			headers:  map[string][]string{"Content-Encoding": {"gzip"}},
		},
		{
			name:     "good bzip2",
			version:  1,
			language: "OPA",
			tags:     []string{"x"},
			compress: CompressBZ2,
			headers:  map[string][]string{"Content-Encoding": {"bz2"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b := NewBundle(tc.version, tc.language, tc.tags...)
			require.NotNil(t, b)

			addPolicies(b)
			addAttributes(b)
			addEntities(b)
			addRelations(b)

			buf := &bytes.Buffer{}
			err := b.Compress(tc.compress, buf)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				b2, err2 := BundleFromAPI(buf, tc.headers)
				if tc.wantErr2 {
					require.Error(t, err2)
					require.Nil(t, b2)
				} else {
					require.NoError(t, err2)
					require.NotNil(t, b2)

					assert.Equal(t, b.Version, b2.Version)
					assert.Equal(t, b.Language, b2.Language)
					assert.Equal(t, len(b.Policies), len(b2.Policies))
					assert.Equal(t, len(b.Attributes), len(b2.Attributes))
					assert.Equal(t, len(b.Entities), len(b2.Entities))
					assert.Equal(t, len(b.Relations), len(b2.Relations))

					assert.Zero(t, len(b2.tags))
					assert.Zero(t, b2.l)
				}
			}
		})
	}
}
