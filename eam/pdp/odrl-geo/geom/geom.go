// Package geom implements the minimal 2D planar geometry primitives, a WKT
// reader and the GeoSPARQL simple-features topological predicates required to
// evaluate ODRL-Geo-NL spatial constraints.
//
// The implementation is deliberately small and dependency-free: it covers
// points, linestrings and (multi)polygons with holes, which is sufficient for
// the BRO simulation. Comparisons are strictly planar (2D) and assume both
// operands share a coordinate reference system; see the package limitations in
// the module documentation. Boundary semantics follow the OGC simple-features /
// DE-9IM reading used by the profile (spec.md §3.2): a point exactly on a
// polygon boundary "touches" but is not "within"; sfDisjoint therefore also
// excludes touching geometries.
package geom

import "math"

// eps is the coordinate tolerance used for on-boundary / equality tests. In RD
// New (metres) this is well below a millimetre.
const eps = 1e-9

// Kind enumerates the supported geometry types.
type Kind uint8

// The supported geometry kinds.
const (
	KindPoint Kind = iota + 1
	KindLineString
	KindPolygon
	KindMultiPolygon
)

// Point is a 2D coordinate.
type Point struct {
	X, Y float64
}

// Polygon is a single polygon: the first ring is the outer boundary, any
// further rings are holes. Rings are closed (first point == last point); the
// reader closes them if needed.
type Polygon struct {
	Rings [][]Point
}

// Geometry is a tagged union of the supported geometry types plus its CRS URI
// (empty means the profile default, EPSG:28992 / RD New).
type Geometry struct {
	Kind    Kind
	CRS     string
	Point   Point
	Line    []Point
	Polygon Polygon
	Multi   []Polygon
}

// Bound is an axis-aligned bounding box.
type Bound struct {
	Min, Max Point
}

// Bound returns the bounding box of the geometry.
func (g Geometry) Bound() Bound {
	switch g.Kind {
	case KindPoint:
		return Bound{g.Point, g.Point}
	case KindLineString:
		return boundOfPoints(g.Line)
	case KindPolygon:
		return boundOfPolygon(g.Polygon)
	case KindMultiPolygon:
		b := empty()
		for _, p := range g.Multi {
			b = b.union(boundOfPolygon(p))
		}
		return b
	}
	return Bound{}
}

func empty() Bound {
	return Bound{
		Min: Point{math.Inf(1), math.Inf(1)},
		Max: Point{math.Inf(-1), math.Inf(-1)},
	}
}

func (b Bound) union(o Bound) Bound {
	return Bound{
		Min: Point{math.Min(b.Min.X, o.Min.X), math.Min(b.Min.Y, o.Min.Y)},
		Max: Point{math.Max(b.Max.X, o.Max.X), math.Max(b.Max.Y, o.Max.Y)},
	}
}

func (b Bound) intersects(o Bound) bool {
	return b.Min.X <= o.Max.X+eps && b.Max.X >= o.Min.X-eps &&
		b.Min.Y <= o.Max.Y+eps && b.Max.Y >= o.Min.Y-eps
}

func boundOfPoints(ps []Point) Bound {
	b := empty()
	for _, p := range ps {
		b = b.union(Bound{p, p})
	}
	return b
}

func boundOfPolygon(p Polygon) Bound {
	if len(p.Rings) == 0 {
		return Bound{}
	}
	return boundOfPoints(p.Rings[0])
}

// NewBBox builds a rectangular polygon geometry from a bounding box in the
// given CRS. It is used to project a bbox request or a tile envelope to a
// polygon (spec.md §6/§7.1).
func NewBBox(minX, minY, maxX, maxY float64, crs string) Geometry {
	ring := []Point{
		{minX, minY}, {maxX, minY}, {maxX, maxY}, {minX, maxY}, {minX, minY},
	}
	return Geometry{Kind: KindPolygon, CRS: crs, Polygon: Polygon{Rings: [][]Point{ring}}}
}
