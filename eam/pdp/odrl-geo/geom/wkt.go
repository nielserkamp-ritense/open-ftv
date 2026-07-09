package geom

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseWKT parses a (GeoSPARQL) WKT literal into a Geometry. The literal may be
// prefixed with an explicit CRS URI in angle brackets, e.g.
//
//	<http://www.opengis.net/def/crs/EPSG/0/28992> POLYGON ((...))
//
// When no CRS prefix is present, defaultCRS is used (which the caller sets to
// the policy / profile default, EPSG:28992).
//
// Supported geometry types: POINT, LINESTRING, POLYGON and MULTIPOLYGON.
func ParseWKT(literal, defaultCRS string) (Geometry, error) {
	s := strings.TrimSpace(literal)
	crs := defaultCRS

	if strings.HasPrefix(s, "<") {
		end := strings.IndexByte(s, '>')
		if end < 0 {
			return Geometry{}, fmt.Errorf("geom: unterminated CRS URI in %q", literal)
		}
		crs = strings.TrimSpace(s[1:end])
		s = strings.TrimSpace(s[end+1:])
	}

	upper := strings.ToUpper(s)
	switch {
	case strings.HasPrefix(upper, "POINT"):
		pts, err := readCoordGroup(s[len("POINT"):])
		if err != nil || len(pts) != 1 {
			return Geometry{}, fmt.Errorf("geom: invalid POINT %q", literal)
		}
		return Geometry{Kind: KindPoint, CRS: crs, Point: pts[0]}, nil

	case strings.HasPrefix(upper, "LINESTRING"):
		pts, err := readCoordGroup(s[len("LINESTRING"):])
		if err != nil || len(pts) < 2 {
			return Geometry{}, fmt.Errorf("geom: invalid LINESTRING %q", literal)
		}
		return Geometry{Kind: KindLineString, CRS: crs, Line: pts}, nil

	case strings.HasPrefix(upper, "MULTIPOLYGON"):
		poly, err := readMultiPolygon(s[len("MULTIPOLYGON"):])
		if err != nil {
			return Geometry{}, fmt.Errorf("geom: invalid MULTIPOLYGON %q: %w", literal, err)
		}
		return Geometry{Kind: KindMultiPolygon, CRS: crs, Multi: poly}, nil

	case strings.HasPrefix(upper, "POLYGON"):
		rings, err := readPolygon(s[len("POLYGON"):])
		if err != nil {
			return Geometry{}, fmt.Errorf("geom: invalid POLYGON %q: %w", literal, err)
		}
		return Geometry{Kind: KindPolygon, CRS: crs, Polygon: Polygon{Rings: rings}}, nil
	}

	return Geometry{}, fmt.Errorf("geom: unsupported WKT geometry %q", literal)
}

// readCoordGroup parses a single parenthesised list of coordinates such as
// "(155000 463000)" or "(x1 y1, x2 y2, ...)".
func readCoordGroup(s string) ([]Point, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "(")
	s = strings.TrimSuffix(strings.TrimSpace(s), ")")
	return readCoords(s)
}

func readCoords(s string) ([]Point, error) {
	var out []Point
	for _, part := range strings.Split(s, ",") {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) < 2 {
			return nil, fmt.Errorf("geom: coordinate needs x and y, got %q", part)
		}
		x, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return nil, err
		}
		y, err2 := strconv.ParseFloat(fields[1], 64)
		if err2 != nil {
			return nil, err2
		}
		out = append(out, Point{x, y})
	}
	return out, nil
}

// readPolygon parses the body after the POLYGON keyword: "((ring1), (ring2))".
func readPolygon(s string) ([][]Point, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "(") || !strings.HasSuffix(s, ")") {
		return nil, fmt.Errorf("geom: malformed polygon %q", s)
	}
	inner := strings.TrimSpace(s[1 : len(s)-1]) // strip outermost parens.

	var rings [][]Point
	for _, ringStr := range splitRings(inner) {
		pts, err := readCoordGroup(ringStr)
		if err != nil {
			return nil, err
		}
		rings = append(rings, closeRing(pts))
	}
	if len(rings) == 0 {
		return nil, fmt.Errorf("geom: polygon has no rings")
	}
	return rings, nil
}

// readMultiPolygon parses "(((ring)),((ring),(hole)))".
func readMultiPolygon(s string) ([]Polygon, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "(") || !strings.HasSuffix(s, ")") {
		return nil, fmt.Errorf("geom: malformed multipolygon %q", s)
	}
	inner := strings.TrimSpace(s[1 : len(s)-1])

	var out []Polygon
	for _, polyStr := range splitParenGroups(inner) {
		rings, err := readPolygon(polyStr)
		if err != nil {
			return nil, err
		}
		out = append(out, Polygon{Rings: rings})
	}
	return out, nil
}

// splitRings splits a polygon body "(r1),(r2)" into "(r1)" and "(r2)".
func splitRings(s string) []string { return splitParenGroups(s) }

// splitParenGroups splits a comma-separated list of parenthesised groups at
// depth zero, keeping the parentheses of each group.
func splitParenGroups(s string) []string {
	var out []string
	depth, start := 0, -1
	for i, r := range s {
		switch r {
		case '(':
			if depth == 0 {
				start = i
			}
			depth++
		case ')':
			depth--
			if depth == 0 && start >= 0 {
				out = append(out, s[start:i+1])
				start = -1
			}
		}
	}
	return out
}

func closeRing(pts []Point) []Point {
	if len(pts) == 0 {
		return pts
	}
	first, last := pts[0], pts[len(pts)-1]
	if first.X != last.X || first.Y != last.Y {
		pts = append(pts, first)
	}
	return pts
}

// ToWKT serialises a geometry to a bare WKT literal (no CRS prefix), so it can
// be carried in a geonl:filterFeatures obligation and handed to a PEP that reads
// the geometry in the layer's native CRS (RD New). The coordinate order and
// formatting match the reader in ParseWKT.
func ToWKT(g Geometry) string {
	switch g.Kind {
	case KindPoint:
		return "POINT (" + fmtPoint(g.Point) + ")"
	case KindLineString:
		return "LINESTRING " + fmtCoordList(g.Line)
	case KindPolygon:
		return "POLYGON " + fmtRings(g.Polygon.Rings)
	case KindMultiPolygon:
		parts := make([]string, 0, len(g.Multi))
		for _, p := range g.Multi {
			parts = append(parts, fmtRings(p.Rings))
		}
		return "MULTIPOLYGON (" + strings.Join(parts, ", ") + ")"
	}
	return ""
}

func fmtPoint(p Point) string {
	return strconv.FormatFloat(p.X, 'f', -1, 64) + " " + strconv.FormatFloat(p.Y, 'f', -1, 64)
}

func fmtCoordList(pts []Point) string {
	parts := make([]string, 0, len(pts))
	for _, p := range pts {
		parts = append(parts, fmtPoint(p))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func fmtRings(rings [][]Point) string {
	parts := make([]string, 0, len(rings))
	for _, r := range rings {
		parts = append(parts, fmtCoordList(r))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
