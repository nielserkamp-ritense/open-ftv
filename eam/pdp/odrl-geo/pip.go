package odrl_geo

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl-geo/geom"
)

// resolveLayerSelection resolves a geonl:LayerSelection to the set of zone
// geometries it selects, via the PIP (spec.md §3.2 / §7.4).
//
// Expected PIP data model: each zone of the referenced layer is a PIP entity
// whose Type() equals the geonl:fromLayer URI. Each such entity carries:
//   - a "geometry" attribute holding a WKT literal (optionally CRS-prefixed);
//   - the attributes referenced by the selection's geonl:where filter
//     (e.g. "classificatie").
//
// The selection's where-filter (a featureProperty constraint) is applied to
// each entity's attributes; only matching zones are returned. The boolean
// reports whether the layer could be resolved at all (an unresolved layer makes
// the spatial constraint not-evaluable / fail-safe, rather than vacuously
// true).
func (c *controller) resolveLayerSelection(ctx *evalCtx, sel *layerSelection, defaultCRS string) ([]geom.Geometry, bool) {
	if c.PIP() == nil || sel == nil || sel.FromLayer == "" {
		return nil, false
	}

	var geoms []geom.Geometry
	resolved := false

	// Record the layer this selection consulted (ingreep 3); it identifies the PIP
	// source set the engine actually read for this decision.
	ctx.consulted.add(sel.FromLayer)

	c.PIP().IterateEntities(func(e models.Entity) {
		if !strings.EqualFold(e.Type(), sel.FromLayer) {
			return
		}
		resolved = true

		attrs := models.MapFromAttributes(e.Attributes())

		// Each zone entity of the selected layer is genuinely consulted: the engine
		// reads its where-filter attribute and its geometry. Record the entity and the
		// specific attribute keys read, matching the key shapes the PIP SourceReferencer
		// registers ("<uid>" and "<uid>/<attr>").
		if uid := e.UID(); uid != "" {
			ctx.consulted.add(uid)
			ctx.consulted.add(uid + "/geometry")
			if sel.Where != nil {
				ctx.consulted.add(uid + "/" + attributeKey(sel.Where.Property))
			}
		}

		// apply the where-filter, if any.
		if sel.Where != nil {
			key := attributeKey(sel.Where.Property)
			val := lookupAttr(attrs, key, sel.Where.Property)
			if compareProperty(sel.Where.Operator, val, sel.Where.RightLiterals) != tTrue {
				return
			}
		}

		wkt := stringAttr(attrs, "geometry")
		if wkt == "" {
			wkt = stringAttr(attrs, "wkt")
		}
		if wkt == "" {
			return
		}
		g, err := geom.ParseWKT(wkt, defaultCRS)
		if err == nil {
			geoms = append(geoms, g)
		}
	})

	return geoms, resolved
}

func stringAttr(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return toString(v)
	}
	return ""
}
