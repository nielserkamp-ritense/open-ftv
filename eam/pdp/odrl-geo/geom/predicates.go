package geom

import "math"

// The GeoSPARQL simple-features predicates. Each returns (result, ok): ok is
// false when the geometry-type combination is not supported by this minimal
// engine, which the caller MUST treat as "not evaluable" (fail-safe, spec.md
// §7.5). All predicates are planar and require matching CRS (checked by the
// caller).

// Intersects reports whether a and b share at least one point.
func Intersects(a, b Geometry) (bool, bool) {
	if !a.Bound().intersects(b.Bound()) {
		return false, true
	}
	switch {
	case a.Kind == KindPoint:
		return pointIntersects(a.Point, b), true
	case b.Kind == KindPoint:
		return pointIntersects(b.Point, a), true
	case a.Kind == KindLineString && b.Kind == KindLineString:
		return lineLineIntersects(a.Line, b.Line), true
	case a.Kind == KindLineString && isPoly(b):
		return lineAreaIntersects(a.Line, b), true
	case isPoly(a) && b.Kind == KindLineString:
		return lineAreaIntersects(b.Line, a), true
	case isPoly(a) && isPoly(b):
		return areaAreaIntersects(a, b), true
	}
	return false, false
}

// Disjoint is the negation of Intersects (spec.md §3.2: excludes touching).
func Disjoint(a, b Geometry) (bool, bool) {
	in, ok := Intersects(a, b)
	if !ok {
		return false, false
	}
	return !in, true
}

// Within reports whether a lies spatially within b (interior/boundary of a
// inside the closure of b, with interiors meeting). A point exactly on a
// polygon boundary is NOT within (spec.md §3.2).
func Within(a, b Geometry) (bool, bool) {
	switch {
	case a.Kind == KindPoint && isPoly(b):
		return pointInAreaInterior(a.Point, b), true
	case a.Kind == KindPoint && b.Kind == KindLineString:
		return pointOnLineInterior(a.Point, b.Line), true
	case a.Kind == KindPoint && b.Kind == KindPoint:
		return ptEq(a.Point, b.Point), true
	case a.Kind == KindLineString && isPoly(b):
		return lineWithinArea(a.Line, b), true
	case isPoly(a) && isPoly(b):
		return areaWithinArea(a, b), true
	}
	return false, false
}

// Contains is the converse of Within: a contains b.
func Contains(a, b Geometry) (bool, bool) { return Within(b, a) }

// Crosses reports whether a and b cross (interiors intersect with a dimension
// lower than the maximum of the two, and neither contains the other).
func Crosses(a, b Geometry) (bool, bool) {
	switch {
	case a.Kind == KindLineString && b.Kind == KindLineString:
		return lineLineCross(a.Line, b.Line), true
	case a.Kind == KindLineString && isPoly(b):
		return lineAreaCross(a.Line, b), true
	case isPoly(a) && b.Kind == KindLineString:
		return lineAreaCross(b.Line, a), true
	}
	return false, false
}

// Overlaps reports whether a and b overlap: same dimension, interiors
// intersect, but neither is within the other.
func Overlaps(a, b Geometry) (bool, bool) {
	if isPoly(a) && isPoly(b) {
		inter, _ := Intersects(a, b)
		if !inter {
			return false, true
		}
		wa, _ := Within(a, b)
		wb, _ := Within(b, a)
		if wa || wb {
			return false, true
		}
		// interior-interior overlap: a vertex of one strictly inside the other.
		if anyVertexStrictlyInside(a, b) || anyVertexStrictlyInside(b, a) {
			return true, true
		}
		return false, true
	}
	return false, false
}

// Touches reports whether a and b share only boundary points (interiors do not
// meet).
func Touches(a, b Geometry) (bool, bool) {
	switch {
	case a.Kind == KindPoint && isPoly(b):
		return pointOnAreaBoundary(a.Point, b), true
	case b.Kind == KindPoint && isPoly(a):
		return pointOnAreaBoundary(b.Point, a), true
	case a.Kind == KindPoint && b.Kind == KindLineString:
		return pointIsLineEndpoint(a.Point, b.Line), true
	case b.Kind == KindPoint && a.Kind == KindLineString:
		return pointIsLineEndpoint(b.Point, a.Line), true
	case isPoly(a) && isPoly(b):
		inter, _ := Intersects(a, b)
		if !inter {
			return false, true
		}
		if anyVertexStrictlyInside(a, b) || anyVertexStrictlyInside(b, a) {
			return false, true
		}
		return true, true
	}
	return false, false
}

// Equals reports whether a and b are spatially equal.
func Equals(a, b Geometry) (bool, bool) {
	if a.Kind != b.Kind {
		return false, true
	}
	switch a.Kind {
	case KindPoint:
		return ptEq(a.Point, b.Point), true
	case KindLineString:
		return lineEq(a.Line, b.Line), true
	case KindPolygon:
		return ringSetEq(a.Polygon.Rings, b.Polygon.Rings), true
	}
	return false, false
}

// --- helpers ---------------------------------------------------------------

func isPoly(g Geometry) bool { return g.Kind == KindPolygon || g.Kind == KindMultiPolygon }

func polygons(g Geometry) []Polygon {
	if g.Kind == KindPolygon {
		return []Polygon{g.Polygon}
	}
	return g.Multi
}

func ptEq(a, b Point) bool { return math.Abs(a.X-b.X) <= eps && math.Abs(a.Y-b.Y) <= eps }

// orient returns the sign of the cross product (b-a)x(c-a): +1 ccw, -1 cw, 0 collinear.
func orient(a, b, c Point) int {
	v := (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
	switch {
	case v > eps:
		return 1
	case v < -eps:
		return -1
	default:
		return 0
	}
}

// onSegment reports whether collinear point p lies on segment a-b.
func onSegment(a, b, p Point) bool {
	if orient(a, b, p) != 0 {
		return false
	}
	return p.X >= math.Min(a.X, b.X)-eps && p.X <= math.Max(a.X, b.X)+eps &&
		p.Y >= math.Min(a.Y, b.Y)-eps && p.Y <= math.Max(a.Y, b.Y)+eps
}

// segIntersect reports whether segments a1-a2 and b1-b2 share any point.
func segIntersect(a1, a2, b1, b2 Point) bool {
	o1, o2 := orient(a1, a2, b1), orient(a1, a2, b2)
	o3, o4 := orient(b1, b2, a1), orient(b1, b2, a2)
	if o1 != o2 && o3 != o4 {
		return true
	}
	return onSegment(a1, a2, b1) || onSegment(a1, a2, b2) ||
		onSegment(b1, b2, a1) || onSegment(b1, b2, a2)
}

// segProperCross reports whether segments cross at a single interior point of
// both (not merely touching at an endpoint or collinearly overlapping).
func segProperCross(a1, a2, b1, b2 Point) bool {
	o1, o2 := orient(a1, a2, b1), orient(a1, a2, b2)
	o3, o4 := orient(b1, b2, a1), orient(b1, b2, a2)
	return o1 != 0 && o2 != 0 && o3 != 0 && o4 != 0 && o1 != o2 && o3 != o4
}

// --- point vs ring / area ---

func pointOnRing(p Point, ring []Point) bool {
	for i := 0; i+1 < len(ring); i++ {
		if onSegment(ring[i], ring[i+1], p) {
			return true
		}
	}
	return false
}

// pointInRing reports strict interior containment via ray casting (boundary
// excluded; test pointOnRing separately).
func pointInRing(p Point, ring []Point) bool {
	in := false
	n := len(ring)
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		yi, yj := ring[i].Y, ring[j].Y
		xi, xj := ring[i].X, ring[j].X
		if (yi > p.Y) != (yj > p.Y) {
			xCross := (xj-xi)*(p.Y-yi)/(yj-yi) + xi
			if p.X < xCross {
				in = !in
			}
		}
	}
	return in
}

// pointInPolygonInterior: strictly inside the outer ring and outside all holes,
// and not on any boundary.
func pointInPolygonInterior(p Point, poly Polygon) bool {
	if len(poly.Rings) == 0 {
		return false
	}
	if pointOnPolygonBoundary(p, poly) {
		return false
	}
	if !pointInRing(p, poly.Rings[0]) {
		return false
	}
	for _, hole := range poly.Rings[1:] {
		if pointInRing(p, hole) {
			return false
		}
	}
	return true
}

func pointOnPolygonBoundary(p Point, poly Polygon) bool {
	for _, ring := range poly.Rings {
		if pointOnRing(p, ring) {
			return true
		}
	}
	return false
}

func pointInAreaInterior(p Point, area Geometry) bool {
	for _, poly := range polygons(area) {
		if pointInPolygonInterior(p, poly) {
			return true
		}
	}
	return false
}

func pointOnAreaBoundary(p Point, area Geometry) bool {
	for _, poly := range polygons(area) {
		if pointOnPolygonBoundary(p, poly) {
			return true
		}
	}
	return false
}

func pointIntersects(p Point, g Geometry) bool {
	switch {
	case g.Kind == KindPoint:
		return ptEq(p, g.Point)
	case g.Kind == KindLineString:
		return pointOnLine(p, g.Line)
	case isPoly(g):
		return pointInAreaInterior(p, g) || pointOnAreaBoundary(p, g)
	}
	return false
}

// --- point vs line ---

func pointOnLine(p Point, line []Point) bool {
	for i := 0; i+1 < len(line); i++ {
		if onSegment(line[i], line[i+1], p) {
			return true
		}
	}
	return false
}

func pointOnLineInterior(p Point, line []Point) bool {
	return pointOnLine(p, line) && !pointIsLineEndpoint(p, line)
}

func pointIsLineEndpoint(p Point, line []Point) bool {
	if len(line) == 0 {
		return false
	}
	return ptEq(p, line[0]) || ptEq(p, line[len(line)-1])
}

// --- line vs line ---

func lineLineIntersects(a, b []Point) bool {
	for i := 0; i+1 < len(a); i++ {
		for j := 0; j+1 < len(b); j++ {
			if segIntersect(a[i], a[i+1], b[j], b[j+1]) {
				return true
			}
		}
	}
	return false
}

func lineLineCross(a, b []Point) bool {
	for i := 0; i+1 < len(a); i++ {
		for j := 0; j+1 < len(b); j++ {
			if segProperCross(a[i], a[i+1], b[j], b[j+1]) {
				return true
			}
		}
	}
	return false
}

// --- line vs area ---

func lineAreaIntersects(line []Point, area Geometry) bool {
	for _, p := range line {
		if pointInAreaInterior(p, area) || pointOnAreaBoundary(p, area) {
			return true
		}
	}
	for _, poly := range polygons(area) {
		for _, ring := range poly.Rings {
			for i := 0; i+1 < len(line); i++ {
				for j := 0; j+1 < len(ring); j++ {
					if segIntersect(line[i], line[i+1], ring[j], ring[j+1]) {
						return true
					}
				}
			}
		}
	}
	return false
}

func lineWithinArea(line []Point, area Geometry) bool {
	interior := false
	for _, p := range line {
		if pointInAreaInterior(p, area) {
			interior = true
		} else if !pointOnAreaBoundary(p, area) {
			return false // a vertex outside the closure.
		}
	}
	// midpoints must stay inside the closure (no segment leaves and re-enters).
	for i := 0; i+1 < len(line); i++ {
		mid := Point{(line[i].X + line[i+1].X) / 2, (line[i].Y + line[i+1].Y) / 2}
		if !pointInAreaInterior(mid, area) && !pointOnAreaBoundary(mid, area) {
			return false
		}
		if pointInAreaInterior(mid, area) {
			interior = true
		}
	}
	return interior
}

func lineAreaCross(line []Point, area Geometry) bool {
	inside, outside := false, false
	for _, p := range line {
		switch {
		case pointInAreaInterior(p, area):
			inside = true
		case !pointOnAreaBoundary(p, area):
			outside = true
		}
	}
	// a proper crossing of a boundary edge means the line passes through.
	for _, poly := range polygons(area) {
		for _, ring := range poly.Rings {
			for i := 0; i+1 < len(line); i++ {
				for j := 0; j+1 < len(ring); j++ {
					if segProperCross(line[i], line[i+1], ring[j], ring[j+1]) {
						inside, outside = true, true
					}
				}
			}
		}
	}
	return inside && outside
}

// --- area vs area ---

func areaAreaIntersects(a, b Geometry) bool {
	if anyVertexInClosure(a, b) || anyVertexInClosure(b, a) {
		return true
	}
	for _, pa := range polygons(a) {
		for _, ra := range pa.Rings {
			for _, pb := range polygons(b) {
				for _, rb := range pb.Rings {
					for i := 0; i+1 < len(ra); i++ {
						for j := 0; j+1 < len(rb); j++ {
							if segIntersect(ra[i], ra[i+1], rb[j], rb[j+1]) {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func areaWithinArea(a, b Geometry) bool {
	// every vertex of a inside the closure of b ...
	for _, poly := range polygons(a) {
		for _, ring := range poly.Rings {
			for _, v := range ring {
				if !pointInAreaInterior(v, b) && !pointOnAreaBoundary(v, b) {
					return false
				}
			}
		}
	}
	// ... and no edge of a properly crosses an edge of b (which would exit b).
	for _, pa := range polygons(a) {
		for _, ra := range pa.Rings {
			for _, pb := range polygons(b) {
				for _, rb := range pb.Rings {
					for i := 0; i+1 < len(ra); i++ {
						for j := 0; j+1 < len(rb); j++ {
							if segProperCross(ra[i], ra[i+1], rb[j], rb[j+1]) {
								return false
							}
						}
					}
				}
			}
		}
	}
	return true
}

func anyVertexInClosure(a, b Geometry) bool {
	for _, poly := range polygons(a) {
		for _, ring := range poly.Rings {
			for _, v := range ring {
				if pointInAreaInterior(v, b) || pointOnAreaBoundary(v, b) {
					return true
				}
			}
		}
	}
	return false
}

func anyVertexStrictlyInside(a, b Geometry) bool {
	for _, poly := range polygons(a) {
		for _, ring := range poly.Rings {
			for _, v := range ring {
				if pointInAreaInterior(v, b) {
					return true
				}
			}
		}
	}
	return false
}

// --- equality helpers ---

func lineEq(a, b []Point) bool {
	if len(a) != len(b) {
		return false
	}
	fwd, rev := true, true
	n := len(a)
	for i := 0; i < n; i++ {
		if !ptEq(a[i], b[i]) {
			fwd = false
		}
		if !ptEq(a[i], b[n-1-i]) {
			rev = false
		}
	}
	return fwd || rev
}

func ringSetEq(a, b [][]Point) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	return sameRing(a[0], b[0])
}

// sameRing compares two closed rings ignoring start vertex and orientation.
func sameRing(a, b []Point) bool {
	na, nb := len(a)-1, len(b)-1 // drop closing duplicate.
	if na != nb || na <= 0 {
		return len(a) == len(b) && na == nb
	}
	as, bs := a[:na], b[:nb]
	for start := 0; start < na; start++ {
		if rotEq(as, bs, start, false) || rotEq(as, bs, start, true) {
			return true
		}
	}
	return false
}

func rotEq(a, b []Point, start int, reverse bool) bool {
	n := len(a)
	for i := 0; i < n; i++ {
		j := (start + i) % n
		if reverse {
			j = (start - i + n*n) % n
		}
		if !ptEq(a[i], b[j]) {
			return false
		}
	}
	return true
}
