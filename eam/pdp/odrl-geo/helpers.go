package odrl_geo

import "strings"

// defaultCRSOr returns the given CRS, or the profile default (EPSG:28992) when
// empty (spec.md §5).
func defaultCRSOr(crs string) string {
	if strings.TrimSpace(crs) == "" {
		return DefaultCRS
	}
	return crs
}

// crsCompatible reports whether two CRS URIs denote the same reference system.
// Empty means the profile default (RD New). Cross-CRS comparisons that would
// require a datum transformation (RDNAPTRANS) are NOT supported and are treated
// as incompatible, making the constraint not-evaluable (fail-safe). See the
// module limitations.
func crsCompatible(a, b string) bool {
	return epsgCode(a) == epsgCode(b)
}

// epsgCode extracts a normalised EPSG code from a CRS URI, defaulting to 28992.
func epsgCode(crs string) string {
	crs = strings.TrimSpace(crs)
	if crs == "" {
		crs = DefaultCRS
	}
	// e.g. http://www.opengis.net/def/crs/EPSG/0/28992 -> 28992.
	if i := strings.LastIndex(strings.ToUpper(crs), "EPSG"); i >= 0 {
		tail := crs[i+4:]
		tail = strings.TrimLeft(tail, "/:0 ")
		if j := strings.IndexAny(tail, "/# "); j >= 0 {
			tail = tail[:j]
		}
		if tail != "" {
			return tail
		}
	}
	return crs
}

// attributeKey returns the lookup key for a geonl:property value: the localname
// of an attribute URI, or the column name itself.
func attributeKey(prop string) string {
	if strings.Contains(prop, "://") || strings.HasPrefix(prop, "urn:") {
		return localName(prop)
	}
	return prop
}

// lookupAttr resolves an attribute from the resource attribute map, trying the
// normalised key, the raw property and the localname.
func lookupAttr(attrs map[string]any, key, raw string) any {
	if attrs == nil {
		return nil
	}
	if v, ok := attrs[key]; ok {
		return v
	}
	if v, ok := attrs[raw]; ok {
		return v
	}
	if v, ok := attrs[localName(raw)]; ok {
		return v
	}
	return nil
}

// asMap coerces an attribute value to a map[string]any (feature attributes).
func asMap(v any) map[string]any {
	switch m := v.(type) {
	case map[string]any:
		return m
	case *map[string]any:
		if m != nil {
			return *m
		}
	}
	return nil
}

// asFloatSlice coerces a bbox attribute value to a []float64.
func asFloatSlice(v any) ([]float64, bool) {
	switch s := v.(type) {
	case []float64:
		return s, true
	case []any:
		out := make([]float64, 0, len(s))
		for _, e := range s {
			f, ok := toFloat(e)
			if !ok {
				return nil, false
			}
			out = append(out, f)
		}
		return out, true
	}
	return nil, false
}

func firstString(vals ...any) string {
	for _, v := range vals {
		if s := toString(v); s != "" {
			return s
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
