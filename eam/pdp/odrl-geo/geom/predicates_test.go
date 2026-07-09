package geom

import "testing"

// square is the Defensie example polygon (RD New), 150000..158000 x 460000..466000.
const square = "<http://www.opengis.net/def/crs/EPSG/0/28992> POLYGON ((150000 460000, 158000 460000, 158000 466000, 150000 466000, 150000 460000))"

func mustWKT(t *testing.T, s string) Geometry {
	t.Helper()
	g, err := ParseWKT(s, "urn:default")
	if err != nil {
		t.Fatalf("ParseWKT(%q): %v", s, err)
	}
	return g
}

func TestPointWithinPolygon(t *testing.T) {
	poly := mustWKT(t, square)

	cases := []struct {
		name       string
		wkt        string
		within     bool
		disjoint   bool
		intersects bool
		touches    bool
	}{
		{"strictly inside", "POINT (155000 463000)", true, false, true, false},
		{"clearly outside", "POINT (100000 400000)", false, true, false, false},
		{"just outside edge", "POINT (149999 463000)", false, true, false, false},
		{"just inside edge", "POINT (150001 463000)", true, false, true, false},
		{"exactly on edge", "POINT (150000 463000)", false, false, true, true},
		{"exactly on corner", "POINT (150000 460000)", false, false, true, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pt := mustWKT(t, c.wkt)
			pt.CRS = poly.CRS // same CRS for the test.

			if got, ok := Within(pt, poly); !ok || got != c.within {
				t.Errorf("Within = %v (ok=%v), want %v", got, ok, c.within)
			}
			if got, ok := Disjoint(pt, poly); !ok || got != c.disjoint {
				t.Errorf("Disjoint = %v (ok=%v), want %v", got, ok, c.disjoint)
			}
			if got, ok := Intersects(pt, poly); !ok || got != c.intersects {
				t.Errorf("Intersects = %v (ok=%v), want %v", got, ok, c.intersects)
			}
			if got, ok := Touches(pt, poly); !ok || got != c.touches {
				t.Errorf("Touches = %v (ok=%v), want %v", got, ok, c.touches)
			}
		})
	}
}

func TestContainsIsConverseOfWithin(t *testing.T) {
	poly := mustWKT(t, square)
	pt := mustWKT(t, "POINT (155000 463000)")
	pt.CRS = poly.CRS
	if got, ok := Contains(poly, pt); !ok || !got {
		t.Errorf("Contains(poly, innerPoint) = %v (ok=%v), want true", got, ok)
	}
}

func TestPolygonWithinPolygon(t *testing.T) {
	outer := mustWKT(t, square)
	inner, _ := ParseWKT("POLYGON ((152000 462000, 154000 462000, 154000 464000, 152000 464000, 152000 462000))", outer.CRS)

	if got, ok := Within(inner, outer); !ok || !got {
		t.Errorf("inner Within outer = %v (ok=%v), want true", got, ok)
	}
	if got, ok := Within(outer, inner); !ok || got {
		t.Errorf("outer Within inner = %v (ok=%v), want false", got, ok)
	}
}

func TestPolygonOverlapAndDisjoint(t *testing.T) {
	a := mustWKT(t, square)
	overlap, _ := ParseWKT("POLYGON ((155000 463000, 160000 463000, 160000 470000, 155000 470000, 155000 463000))", a.CRS)
	apart, _ := ParseWKT("POLYGON ((200000 500000, 201000 500000, 201000 501000, 200000 501000, 200000 500000))", a.CRS)

	if got, ok := Overlaps(a, overlap); !ok || !got {
		t.Errorf("Overlaps = %v (ok=%v), want true", got, ok)
	}
	if got, ok := Intersects(a, overlap); !ok || !got {
		t.Errorf("Intersects = %v (ok=%v), want true", got, ok)
	}
	if got, ok := Disjoint(a, apart); !ok || !got {
		t.Errorf("Disjoint = %v (ok=%v), want true", got, ok)
	}
}

func TestLineCrossesPolygon(t *testing.T) {
	poly := mustWKT(t, square)
	crossing, _ := ParseWKT("LINESTRING (140000 463000, 170000 463000)", poly.CRS)
	outside, _ := ParseWKT("LINESTRING (100000 100000, 120000 120000)", poly.CRS)

	if got, ok := Crosses(crossing, poly); !ok || !got {
		t.Errorf("Crosses = %v (ok=%v), want true", got, ok)
	}
	if got, ok := Crosses(outside, poly); !ok || got {
		t.Errorf("Crosses(outside) = %v (ok=%v), want false", got, ok)
	}
}

func TestLineCrossesLine(t *testing.T) {
	a, _ := ParseWKT("LINESTRING (0 0, 10 10)", "c")
	b, _ := ParseWKT("LINESTRING (0 10, 10 0)", "c")
	parallel, _ := ParseWKT("LINESTRING (0 1, 10 11)", "c")

	if got, ok := Crosses(a, b); !ok || !got {
		t.Errorf("Crosses(X) = %v (ok=%v), want true", got, ok)
	}
	if got, ok := Crosses(a, parallel); !ok || got {
		t.Errorf("Crosses(parallel) = %v (ok=%v), want false", got, ok)
	}
}

func TestEquals(t *testing.T) {
	a := mustWKT(t, square)
	// same ring, rotated start + reversed orientation.
	b, _ := ParseWKT("POLYGON ((158000 466000, 158000 460000, 150000 460000, 150000 466000, 158000 466000))", a.CRS)
	if got, ok := Equals(a, b); !ok || !got {
		t.Errorf("Equals(rotated/reversed) = %v (ok=%v), want true", got, ok)
	}
}

func TestUnsupportedComboNotEvaluable(t *testing.T) {
	a, _ := ParseWKT("LINESTRING (0 0, 1 1)", "c")
	b, _ := ParseWKT("LINESTRING (2 2, 3 3)", "c")
	// Within(line, line) is not implemented -> ok=false (not evaluable).
	if _, ok := Within(a, b); ok {
		t.Errorf("Within(line,line) ok = true, want false (not evaluable)")
	}
}

func TestBBoxPolygon(t *testing.T) {
	bb := NewBBox(150000, 460000, 158000, 466000, "crs")
	pt := Geometry{Kind: KindPoint, CRS: "crs", Point: Point{155000, 463000}}
	if got, ok := Within(pt, bb); !ok || !got {
		t.Errorf("point Within bbox = %v (ok=%v), want true", got, ok)
	}
}
