package pip

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/deiu/rdf2go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/rdf"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/turtle"
)

func TestLoadRDF(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		path           string
		mime           string
		wantLog        int
		wantAttributes int
		wantEntities   int
	}{
		{name: "nil - without mime", wantLog: 1},
		{name: "nil - with mime", mime: mime.MimeTypeTurtle, wantLog: 1},
		{name: "bad input (1)", path: "../../testdata/error.ttl", mime: mime.MimeTypeTurtle, wantLog: 1},
		{name: "bad input (2)", path: "../../testdata/unittest/pip/attributes/attribute.yaml", mime: mime.MimeTypeYAML, wantLog: 1},
		{name: "good input (1)", path: "../../testdata/rdf/entities/services/rvig.ttl", mime: mime.MimeTypeTurtle, wantAttributes: 0, wantEntities: 1},
		{name: "good input (2)", path: "../../testdata/rdf/entities/services/rvig.ttl", wantAttributes: 0, wantEntities: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)

			var f io.Reader
			if tc.path != "" {
				var err error
				f, err = os.Open(tc.path)
				require.NoError(t, err)
			}

			s := memory.New()
			ap := NewAttributeStore(context.Background(), s, "attribute")
			ep := NewEntityStore(context.Background(), s, "entity")

			p := &PIP{
				logger:           slog.New(h),
				store:            s,
				attributePersist: ap,
				entityPersist:    ep,
			}

			p.loadRDF(f, tc.path, tc.mime)
			assert.Equal(t, tc.wantLog, h.Count())

			var count int
			p.IterateAttributes(func(attribute *models.Attribute) {
				count++
			})
			assert.Equal(t, tc.wantAttributes, count)
		})
	}
}

func TestConvertLiteral(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		term    rdf2go.Term
		want    any
		wantErr bool
	}{
		{name: "not a literal", term: rdf2go.NewResource("https://ftv.nl"), wantErr: true},
		{name: "string", term: rdf2go.NewLiteral("haha"), want: "haha"},
		{name: "string with language", term: rdf2go.NewLiteralWithLanguage("haha", "en"), want: "haha"},
		{name: "int", term: rdf2go.NewLiteralWithDatatype("123456789", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int")), want: int64(123456789)},
		{name: "boolean", term: rdf2go.NewLiteralWithDatatype("1", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#boolean")), want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := &loader{}

			got, err := l.convertLiteral(tc.term)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestAllLiterals(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		list []*rdf2go.Triple
		want bool
	}{
		{
			name: "empty list",
			list: []*rdf2go.Triple{},
			want: true,
		},
		{
			name: "not a literal",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int")),
			},
		},
		{
			name: "literals + non literal",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("hello")),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("1234", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#unsignedShort"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("-987", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int")),
			},
		},
		{
			name: "multiple strings",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("hello")),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("abc")),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("world")),
			},
			want: true,
		},
		{
			name: "mixed",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("12", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#byte"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("abc")),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("123", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#unsignedByte"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("-987", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("world")),
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := &loader{}

			got := l.allLiteral(tc.list)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConvertLiterals(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		list    []*rdf2go.Triple
		want    any
		wantErr bool
	}{
		{
			name: "empty list",
			list: []*rdf2go.Triple{},
			want: []any{},
		},
		{
			name: "not a literal",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int")),
			},
			wantErr: true,
		},
		{
			name: "literals + non literal",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("hello")),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("1234", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#unsignedShort"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("-987", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int")),
			},
			wantErr: true,
		},
		{
			name: "single string",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("abc")),
			},
			want: "abc",
		},
		{
			name: "multiple strings",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("hello")),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("abc")),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("world")),
			},
			want: []any{"hello", "abc", "world"},
		},
		{
			name: "single integer",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("98765432", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#positiveInteger"))),
			},
			want: uint64(98765432),
		},
		{
			name: "multiple integers",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("12", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#byte"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("-123456789", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#long"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("1234", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#unsignedShort"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("-987", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("98765432", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#positiveInteger"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("123", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#unsignedByte"))),
			},
			want: []any{int64(12), int64(-123456789), uint64(1234), int64(-987), uint64(98765432), uint64(123)},
		},
		{
			name: "mixed",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("12", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#byte"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("abc")),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("123", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#unsignedByte"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteralWithDatatype("-987", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int"))),
				rdf2go.NewTriple(rdf2go.NewResource("x"), rdf2go.NewResource("y"), rdf2go.NewLiteral("world")),
			},
			want: []any{int64(12), "abc", uint64(123), int64(-987), "world"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := &loader{}

			got, err := l.convertLiterals(tc.list)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestConvertObject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		path    string
		mt      string
		term    rdf2go.Term
		wantK   string
		wantV   any
		wantErr bool
	}{
		{
			name:  "string with language",
			path:  "../../testdata/rdf/entities/services/rvig.ttl",
			mt:    "text/turtle",
			term:  rdf2go.NewLiteralWithLanguage("haha", "en"),
			wantV: "haha",
		},
		{
			name:  "int",
			path:  "../../testdata/rdf/entities/services/rvig.ttl",
			mt:    "text/turtle",
			term:  rdf2go.NewLiteralWithDatatype("123456789", rdf2go.NewResource("http://www.w3.org/2001/XMLSchema#int")),
			wantV: int64(123456789),
		},
		{
			name:  "rvig parameterBSN",
			path:  "../../testdata/rdf/entities/services/rvig.ttl",
			mt:    "text/turtle",
			term:  rdf2go.NewResource("http://example.com/rvig/rdf/services#parameterBSN"),
			wantK: "burgerservicenummer",
		},
		{
			name:  "rvig parameterType",
			path:  "../../testdata/rdf/entities/services/rvig.ttl",
			mt:    "text/turtle",
			term:  rdf2go.NewResource("http://example.com/rvig/rdf/services#parameterType"),
			wantK: "type",
			wantV: map[string]any{"allowedValues": []any{
				"ZoekMetStraatHuisnummerEnGemeenteVanInschrijving",
				"ZoekMetAdresseerbaarObjectIdentificatie",
				"ZoekMetNaamEnGemeenteVanInschrijving",
				"ZoekMetPostcodeEnHuisnummer",
				"ZoekMetNummeraanduidingidentificatie",
				"RaadpleegMetBurgerservicenummer",
				"ZoekMetGeslachtsnaamEnGeboortedatum",
			}},
		},
		{
			name:    "unknown resource",
			path:    "../../testdata/rdf/entities/services/rvig.ttl",
			mt:      "text/turtle",
			term:    rdf2go.NewResource("http://example.com/rvig/rdf/services#parameterOopsie"),
			wantErr: true,
		},
		{
			name:    "bad term",
			path:    "../../testdata/rdf/entities/services/rvig.ttl",
			mt:      "text/turtle",
			term:    &fakeTerm{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)

			f, err := os.Open(tc.path)
			require.NoError(t, err)
			defer f.Close()

			graph, err3 := turtle.Load(f, tc.mt)
			require.NoError(t, err3)
			require.NotNil(t, graph)

			l := &loader{
				path:   tc.path,
				mt:     tc.mt,
				logger: slog.New(h),
				graph:  graph,
			}

			k, v, err2 := l.convertObject(tc.term)
			if tc.wantErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err2)
				assert.Equal(t, tc.wantK, k)

				switch tp := tc.wantV.(type) {
				case map[string]any:
					m, ok := v.(map[string]any)
					require.True(t, ok)
					for i := range tp {
						s1, ok1 := tp[i].([]any)
						s2, ok2 := m[i].([]any)
						if !ok1 || !ok2 {
							require.EqualValues(t, tp[i], m[i])
						} else {
							require.Equal(t, len(s1), len(s2))

							var ok3 bool
							for x := range s1 {
								for y := range s2 {
									if s1[x] == s2[y] {
										ok3 = true
									}
								}
							}

							assert.True(t, ok3)
						}
					}

				default:
					assert.EqualValues(t, tc.wantV, tp)

				}
			}
		})
	}
}

func TestConvertObjects(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		path    string
		list    []*rdf2go.Triple
		want    any
		wantErr bool
	}{
		{
			name: "empty list",
			path: "../../testdata/unittest/rdf/errors.ttl",
			want: map[string]any{},
		},
		{
			name: "single object",
			path: "../../testdata/unittest/rdf/errors.ttl",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("a"), rdf2go.NewResource("b"), rdf2go.NewResource("http://example.com/rdf/services#singleObject")),
			},
			want: map[string]any{
				"singleObject": map[string]any{"int": int64(123)},
			},
		},
		{
			name: "few object",
			path: "../../testdata/unittest/rdf/errors.ttl",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("a"), rdf2go.NewResource("b"), rdf2go.NewResource("http://example.com/rdf/services#fewObjects")),
			},
			want: map[string]any{
				"fewObjects": map[string]any{
					"int":   int64(123),
					"hello": "world",
					"bool":  true,
				},
			},
		},
		{
			name: "mixed",
			path: "../../testdata/unittest/rdf/errors.ttl",
			list: []*rdf2go.Triple{
				rdf2go.NewTriple(rdf2go.NewResource("a"), rdf2go.NewResource("b"), rdf2go.NewResource("http://example.com/rdf/services#mixedValues")),
			},
			want: map[string]any{
				"mixedValues": map[string]any{
					"int":     int64(123),
					"hello":   "world",
					"__obj_1": int64(987),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)

			f, err := os.Open(tc.path)
			require.NoError(t, err)
			defer f.Close()

			graph, err3 := turtle.Load(f, "text/turtle")
			require.NoError(t, err3)
			require.NotNil(t, graph)

			l := &loader{
				path:   tc.path,
				mt:     "text/turtle",
				logger: slog.New(h),
				graph:  graph,
			}

			got, err2 := l.convertObjects(tc.list)
			if tc.wantErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err2)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestGetKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		path    string
		sub     rdf2go.Term
		want    string
		wantErr bool
	}{
		{
			name: "rvig parameterBSN",
			path: "../../testdata/rdf/entities/services/rvig.ttl",
			sub:  rdf2go.NewResource("http://example.com/rvig/rdf/services#parameterBSN"),
			want: "burgerservicenummer",
		},
		{
			name:    "bad resource",
			path:    "../../testdata/rdf/entities/services/rvig.ttl",
			sub:     rdf2go.NewResource("http://example.com/rvig/rdf/services#oopsie"),
			wantErr: true,
		},
		{
			name:    "bad key type",
			path:    "../../testdata/unittest/rdf/errors.ttl",
			sub:     rdf2go.NewResource("http://example.com/rdf/services#badKey"),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)

			f, err := os.Open(tc.path)
			require.NoError(t, err)
			defer f.Close()

			graph, err3 := turtle.Load(f, "text/turtle")
			require.NoError(t, err3)
			require.NotNil(t, graph)

			l := &loader{
				path:   tc.path,
				mt:     "test/turtle",
				logger: slog.New(h),
				graph:  graph,
			}

			got, err2 := l.getString(tc.sub, rdf.FTVAttributeKey)
			if tc.wantErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestGetKeyValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		path    string
		sub     rdf2go.Term
		wantK   string
		wantV   any
		wantErr bool
	}{
		{
			name:    "bad key",
			path:    "../../testdata/unittest/rdf/errors.ttl",
			sub:     rdf2go.NewResource("http://example.com/rdf/services#badKey"),
			wantErr: true,
		},
		{
			name:    "bad value",
			path:    "../../testdata/unittest/rdf/errors.ttl",
			sub:     rdf2go.NewResource("http://example.com/rdf/services#badValue"),
			wantErr: true,
		},
		{
			name:  "good",
			path:  "../../testdata/unittest/rdf/errors.ttl",
			sub:   rdf2go.NewResource("http://example.com/rdf/services#goodKV"),
			wantK: "hello",
			wantV: "world",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)

			f, err := os.Open(tc.path)
			require.NoError(t, err)
			defer f.Close()

			graph, err3 := turtle.Load(f, "text/turtle")
			require.NoError(t, err3)
			require.NotNil(t, graph)

			l := &loader{
				path:   tc.path,
				mt:     "test/turtle",
				logger: slog.New(h),
				graph:  graph,
			}

			k, v, err2 := l.getKeyValue(tc.sub)
			if tc.wantErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantK, k)
				assert.Equal(t, tc.wantV, v)
			}
		})
	}
}

type fakeTerm struct {
	URI string
}

func (tt *fakeTerm) Equal(rdf2go.Term) bool {
	return false
}
func (tt *fakeTerm) String() string {
	return ""
}
func (tt *fakeTerm) RawValue() string {
	return ""
}
