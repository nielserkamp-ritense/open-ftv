package odrl_geo

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl-geo/geom"
)

// Authorize implements the pdp.Controller interface. It evaluates the ODRL-Geo
// policies against the PARC and returns a default-deny AuthZEN decision with
// "ja, mits" obligations in Response.Attributes (spec.md §7).
func (c *controller) Authorize(uid string, req *models.PARC) (*models.Response, error) {
	parc := c.Map(req)
	ctx := c.newEvalCtx(parc)

	merged := c.merge()
	rules, source := selectRules(merged, ctx)

	resp := c.decide(ctx, merged, rules)

	// Ingreep 1: stamp the content hash of the exact policy envelope that governed
	// the decision, so the ADL adl.core.policies reference {key: hash} resolves back
	// to the policy source via the PAP's content-addressable ReadByHash.
	if resp.PolicyKey != "" {
		if h := c.policyHash(resp.PolicyKey); h != "" {
			resp.PolicyHash = h
		}
	}

	// Ingreep 3: expose the exact set of PIP sources this evaluation consulted
	// (LayerSelection zone entities and the attributes read off them) so the ADL
	// level-3 information provider can report consulted-not-merely-matched sources.
	if keys := ctx.consulted.list(); len(keys) > 0 {
		if resp.Attributes == nil {
			resp.Attributes = map[string]any{}
		}
		resp.Attributes[attrConsultedSources] = keys
	}

	if c.Logger() != nil {
		c.Logger().Debug("odrl-geo decision",
			"controller", c.String(), "request-uid", uid,
			"allowed", resp.Allowed, "policy", resp.PolicyKey, "reason", resp.Message, "source", source)
	}
	return resp, nil
}

// mergedView is the deduplicated union of all loaded documents.
type mergedView struct {
	policies   map[string]*geoPolicy
	areas      map[string]*area
	selections map[string]*layerSelection
	accessURL  map[string]string
}

// merge deduplicates policies (by uid) and merges shared entities across every
// loaded envelope (each envelope keeps the full source, so the same policy can
// appear in several envelopes).
func (c *controller) merge() *mergedView {
	m := &mergedView{
		policies:   map[string]*geoPolicy{},
		areas:      map[string]*area{},
		selections: map[string]*layerSelection{},
		accessURL:  map[string]string{},
	}
	for _, d := range c.snapshot() {
		for _, p := range d.Policies {
			m.policies[p.UID] = p
		}
		for k, v := range d.Areas {
			m.areas[k] = v
		}
		for k, v := range d.Selections {
			m.selections[k] = v
		}
		for k, v := range d.AccessURL {
			m.accessURL[k] = v
		}
	}
	return m
}

// evalCtx is the resolved evaluation context of one request (spec.md §6).
type evalCtx struct {
	subjectID  string
	action     string // localname, e.g. "read".
	purpose    string
	layer      string
	resType    string
	crs        string
	attributes map[string]any

	featureGeom geom.Geometry
	geomOK      bool

	// consulted records the PIP sources actually read during this evaluation.
	consulted *consultedSet
}

// attrConsultedSources is the Response.Attributes key under which the engine reports
// the exactly-consulted PIP source keys (ingreep 3).
const attrConsultedSources = "consultedSources"

// consultedSet is an insertion-ordered set of consulted PIP source keys.
type consultedSet struct {
	keys []string
	seen map[string]struct{}
}

func newConsultedSet() *consultedSet {
	return &consultedSet{seen: map[string]struct{}{}}
}

func (s *consultedSet) add(key string) {
	if s == nil || key == "" {
		return
	}
	if _, ok := s.seen[key]; ok {
		return
	}
	s.seen[key] = struct{}{}
	s.keys = append(s.keys, key)
}

func (s *consultedSet) list() []string {
	if s == nil {
		return nil
	}
	return s.keys
}

func (c *controller) newEvalCtx(parc *models.PARC) *evalCtx {
	ctx := &evalCtx{
		subjectID: parc.Principal.ID(),
		action:    localName(parc.Action.ID()),
		resType:   strings.ToLower(parc.Resource.Type()),
		consulted: newConsultedSet(),
	}

	// action purpose (processing activity or explicit purpose).
	if pa := parc.Action.Attributes(); pa != nil {
		ctx.purpose = firstString(pa.GetAttributeValue(actionProcessingActivity), pa.GetAttributeValue(actionPurpose))
	}

	ra := parc.Resource.Attributes()
	if ra == nil {
		return ctx
	}
	ctx.layer = toString(ra.GetAttributeValue(propLayer))
	ctx.crs = toString(ra.GetAttributeValue(propCRS))
	ctx.attributes = asMap(ra.GetAttributeValue(propAttributes))

	ctx.featureGeom, ctx.geomOK = c.resolveFeatureGeometry(ra, ctx.crs)
	return ctx
}

// resolveFeatureGeometry builds the feature/tile/bbox geometry from the resource
// properties: an explicit WKT geometry, else a bbox [minx,miny,maxx,maxy], else
// a tile envelope carried as geometry (spec.md §7.1).
func (c *controller) resolveFeatureGeometry(ra models.AttributeSet, crs string) (geom.Geometry, bool) {
	if wkt := toString(ra.GetAttributeValue(propGeometry)); wkt != "" {
		g, err := geom.ParseWKT(wkt, defaultCRSOr(crs))
		if err != nil {
			c.Logger().Warn("odrl-geo: cannot parse feature geometry", "error", err)
			return geom.Geometry{}, false
		}
		return g, true
	}
	if bb, ok := asFloatSlice(ra.GetAttributeValue(propBBox)); ok && len(bb) == 4 {
		return geom.NewBBox(bb[0], bb[1], bb[2], bb[3], defaultCRSOr(crs)), true
	}
	return geom.Geometry{}, false
}

// selectRules applies ODRL-AP-NL precedence: an Agreement for this assignee on
// this layer governs over the Offer (spec.md §1 / ODRL-AP-NL). It returns the
// governing rule set plus a human-readable source label.
func selectRules(m *mergedView, ctx *evalCtx) ([]ruleRef, string) {
	var agreements, offers []ruleRef

	for _, p := range m.policies {
		appliesToSubject := !p.isAgreement() || p.Assignee == ctx.subjectID
		if !appliesToSubject {
			continue
		}
		for _, r := range p.Permissions {
			if ruleMatchesLayer(r, ctx.layer, m.accessURL) {
				ref := ruleRef{policy: p, rule: r, prohibition: false}
				if p.isAgreement() {
					agreements = append(agreements, ref)
				} else {
					offers = append(offers, ref)
				}
			}
		}
		for _, r := range p.Prohibitions {
			if ruleMatchesLayer(r, ctx.layer, m.accessURL) {
				ref := ruleRef{policy: p, rule: r, prohibition: true}
				if p.isAgreement() {
					agreements = append(agreements, ref)
				} else {
					offers = append(offers, ref)
				}
			}
		}
	}

	if hasPermission(agreements) {
		return agreements, "agreement"
	}
	return offers, "offer"
}

type ruleRef struct {
	policy      *geoPolicy
	rule        *geoRule
	prohibition bool
}

func hasPermission(refs []ruleRef) bool {
	for _, r := range refs {
		if !r.prohibition {
			return true
		}
	}
	return false
}

// decide runs the default-deny decision: a prohibition that fires over the whole
// resource denies; otherwise an active permission permits (with obligations).
//
// For a tile/layer request the engine follows the spec-model "decision per tile
// with a filter-obligation" (spec.md §7.2): a rule that cannot be evaluated as a
// single whole-tile yes/no — a spatial rule that partially covers the envelope,
// or a geonl:featureProperty constraint that has no per-feature attribute to
// test — does not fail-closed to deny but permits the tile carrying a
// geonl:filterFeatures obligation whose predicate lets the PEP filter per
// feature. Only when no such predicate can be derived (e.g. a prohibition
// without a resolvable delimitation) does the tile remain a fail-closed deny.
func (c *controller) decide(ctx *evalCtx, m *mergedView, rules []ruleRef) *models.Response {
	isTile := ctx.resType == resourceTile || ctx.resType == resourceLayer
	var tilePred []map[string]any
	// tilePolicy is the policy whose rule contributed the filter predicate(s);
	// it becomes the audit policy reference when the tile is permitted with a
	// filter-obligation, so the audit records the real governing policy instead
	// of a synthetic "tile" placeholder (audit_identifiers.policy_version).
	var tilePolicy *geoPolicy

	addPred := func(policy *geoPolicy, preds []map[string]any) {
		if isTile && len(preds) > 0 {
			tilePred = append(tilePred, preds...)
			if tilePolicy == nil {
				tilePolicy = policy
			}
		}
	}

	// prohibitions win.
	for _, ref := range rules {
		if !ref.prohibition || !actionMatches(ref.rule, ctx) {
			continue
		}
		fires, preds := c.evalRule(ref.rule, ctx, m, true)
		if fires {
			return deny(ref.policy.UID, "prohibited by policy: "+shortReason(ref.rule))
		}
		addPred(ref.policy, preds)
	}

	// permissions.
	for _, ref := range rules {
		if ref.prohibition || !actionMatches(ref.rule, ctx) {
			continue
		}
		if !purposeMatch(ref.rule.Purpose, ctx) {
			continue
		}
		active, preds := c.evalRule(ref.rule, ctx, m, false)
		if active {
			return c.permit(ctx, ref, m, tilePred)
		}
		// permission is inactive; if the only reason is a non-evaluable
		// (per-feature) constraint on a tile/layer request, keep its predicate so
		// the tile can be permitted with a filter-obligation instead of denied.
		addPred(ref.policy, preds)
	}

	if len(tilePred) > 0 {
		// tile/layer with no whole-tile permission: permit but filter per feature,
		// crediting the real policy that carried the constraint (never a synthetic
		// uid) so the audit policy reference is honest.
		pol := tilePolicy
		if pol == nil {
			pol = &geoPolicy{}
		}
		return c.permit(ctx, ruleRef{policy: pol}, m, tilePred)
	}
	return deny("", "no applicable permission (default deny)")
}

// evalRule evaluates all constraints of a rule. For a prohibition it returns
// whether the prohibition fires (fail-safe: a non-evaluable constraint counts
// as firing). For a permission it returns whether the permission is active
// (fail-safe: a non-evaluable constraint deactivates it).
//
// The second return value carries, for a tile/layer request, the
// geonl:filterFeatures predicates derived from constraints that cannot be
// decided as a single whole-tile yes/no: a spatial constraint whose delimiting
// geometry straddles the tile envelope, or a geonl:featureProperty constraint
// that has no per-feature attribute to test. The caller uses these to permit
// the tile with a filter-obligation instead of failing closed (spec.md §7.2).
func (c *controller) evalRule(r *geoRule, ctx *evalCtx, m *mergedView, prohibition bool) (bool, []map[string]any) {
	isTile := ctx.resType == resourceTile || ctx.resType == resourceLayer
	result := true // prohibition: fires; permission: active.
	var preds []map[string]any

	for _, cons := range r.Constraints {
		state := c.evalConstraint(cons, ctx, m)

		if isTile {
			if cons.isSpatial() {
				if p := c.spatialTilePredicate(cons, ctx, m); p != nil {
					preds = append(preds, p)
				}
			} else if state == tUnknown {
				// no per-feature attribute in a tile/layer request: filter per feature.
				if p := propertyTilePredicate(cons); p != nil {
					preds = append(preds, p)
				}
			}
		}

		if prohibition {
			// fires unless the constraint is definitively false.
			if state == tFalse {
				result = false
			}
		} else {
			// active only if every constraint is definitively true.
			if state != tTrue {
				result = false
			}
		}
	}
	return result, preds
}

// evalConstraint dispatches to the property or spatial evaluator.
func (c *controller) evalConstraint(cons *geoConstraint, ctx *evalCtx, m *mergedView) triState {
	if cons.isSpatial() {
		return c.evalSpatial(cons, ctx, m)
	}
	return evalProperty(cons, ctx)
}

// evalProperty evaluates a geonl:featureProperty constraint against the
// resource attributes (spec.md §3.1 / §7.4).
func evalProperty(cons *geoConstraint, ctx *evalCtx) triState {
	prop := cons.Property
	if prop == "" && cons.LeftOperand != geonlFeatureProperty {
		// BRP-style: the attribute URI is the left operand itself.
		prop = cons.LeftOperand
	}
	if prop == "" {
		return tUnknown
	}
	val := lookupAttr(ctx.attributes, attributeKey(prop), prop)
	if val == nil {
		return tUnknown
	}
	return compareProperty(cons.Operator, val, cons.RightLiterals)
}

// evalSpatial evaluates a geonl:featureGeometry constraint (spec.md §7.4). The
// spatial relation must hold against every resolved right geometry (universal
// quantification).
func (c *controller) evalSpatial(cons *geoConstraint, ctx *evalCtx, m *mergedView) triState {
	if !ctx.geomOK {
		return tUnknown
	}
	rights, resolved := c.resolveRight(cons, ctx, m)
	if !resolved {
		return tUnknown
	}
	if len(rights) == 0 {
		// universal quantification over the empty set is vacuously true.
		return tTrue
	}
	for _, right := range rights {
		if !crsCompatible(ctx.featureGeom.CRS, right.CRS) {
			return tUnknown
		}
		ok, evaluable := applyOperator(cons.Operator, ctx.featureGeom, right)
		if !evaluable {
			return tUnknown
		}
		if !ok {
			return tFalse
		}
	}
	return tTrue
}

// resolveRight resolves the right operand(s) of a spatial constraint: an inline
// WKT literal, a named geonl:Area, or a geonl:LayerSelection resolved via PIP.
func (c *controller) resolveRight(cons *geoConstraint, ctx *evalCtx, m *mergedView) ([]geom.Geometry, bool) {
	def := defaultCRSOr(ctx.crs)

	if cons.RightGeometry != "" {
		g, err := geom.ParseWKT(cons.RightGeometry, def)
		if err != nil {
			return nil, false
		}
		return []geom.Geometry{g}, true
	}

	if cons.RightRef == "" {
		return nil, false
	}
	if a, ok := m.areas[cons.RightRef]; ok {
		g, err := geom.ParseWKT(a.WKT, def)
		if err != nil {
			return nil, false
		}
		return []geom.Geometry{g}, true
	}
	if sel, ok := m.selections[cons.RightRef]; ok {
		return c.resolveLayerSelection(ctx, sel, def)
	}
	return nil, false
}

// spatialTilePredicate derives a geonl:filterFeatures predicate from a spatial
// constraint whose delimiting geometry cannot be decided whole-tile: the tile
// envelope straddles the boundary (some features inside, some outside), or the
// request is a whole layer (no envelope). It returns nil when no filter can be
// expressed for the PEP.
//
// The predicate is always phrased as "notWithin <geometry>" — the only spatial
// form the BRO PEP's implement (enforce.go applyFilterFeatures): keep the
// features that lie outside the delimiting geometry. This is exactly right for a
// prohibition on features within an area (keep the ones not within it) and for a
// permission requiring disjointness (same set). The geometry is serialised to
// bare WKT in RD New (EPSG:28992); geometries in another CRS are skipped because
// the PEP reads the WKT in the layer's native (RD) CRS.
func (c *controller) spatialTilePredicate(cons *geoConstraint, ctx *evalCtx, m *mergedView) map[string]any {
	rights, resolved := c.resolveRight(cons, ctx, m)
	if !resolved {
		return nil
	}
	for _, right := range rights {
		if epsgCode(right.CRS) != epsgCode(DefaultCRS) {
			continue
		}
		if ctx.geomOK && !envelopeStraddles(ctx.featureGeom, right) {
			// concrete envelope fully inside or fully outside: decided whole-tile.
			continue
		}
		return map[string]any{
			"spatial": "notWithin",
			"wkt":     geom.ToWKT(right),
			"crs":     DefaultCRS,
		}
	}
	return nil
}

// envelopeStraddles reports whether a tile envelope crosses the boundary of the
// delimiting geometry (partly inside, partly outside).
func envelopeStraddles(envelope, area geom.Geometry) bool {
	inter, iok := geom.Intersects(envelope, area)
	within, wok := geom.Within(envelope, area)
	contains, cok := geom.Contains(envelope, area)
	return iok && inter && wok && !within && cok && !contains
}

// propertyTilePredicate derives a geonl:filterFeatures predicate from a
// geonl:featureProperty constraint so the PEP can apply it per feature (as a SQL
// / CQL filter). It returns nil when the constraint has no usable
// property/operator/value triple.
func propertyTilePredicate(cons *geoConstraint) map[string]any {
	prop := cons.Property
	if prop == "" && cons.LeftOperand != geonlFeatureProperty {
		// BRP-style: the attribute URI is the left operand itself.
		prop = cons.LeftOperand
	}
	if prop == "" || len(cons.RightLiterals) == 0 {
		return nil
	}
	op := localName(cons.Operator)
	if !pepPropertyOperator(op) {
		return nil
	}
	return map[string]any{
		"property": attributeKey(prop),
		"operator": op,
		"value":    coerceValue(cons.RightLiterals[0].Value),
	}
}

// pepPropertyOperator reports whether an ODRL comparison operator localname is
// one the BRO PEP can translate to SQL/CQL (enforce.go sqlOperator).
func pepPropertyOperator(op string) bool {
	switch op {
	case "lt", "lteq", "eq", "neq", "gt", "gteq":
		return true
	}
	return false
}

// applyOperator applies a GeoSPARQL simple-features operator.
func applyOperator(op string, a, b geom.Geometry) (bool, bool) {
	switch op {
	case sfWithin:
		return geom.Within(a, b)
	case sfDisjoint:
		return geom.Disjoint(a, b)
	case sfCrosses:
		return geom.Crosses(a, b)
	case sfIntersects:
		return geom.Intersects(a, b)
	case sfContains:
		return geom.Contains(a, b)
	case sfOverlaps:
		return geom.Overlaps(a, b)
	case sfTouches:
		return geom.Touches(a, b)
	case sfEquals:
		return geom.Equals(a, b)
	}
	return false, false
}

// permit builds a permit response, mapping the rule's duties (and, for a
// tile/layer request, one geonl:filterFeatures obligation per derived predicate)
// to AuthZEN obligations (spec.md §7.3). Each filterFeatures obligation carries a
// non-empty predicate the BRO PEP can enforce: a spatial ("spatial"/"wkt"/"crs")
// or a per-feature ("property"/"operator"/"value") form.
func (c *controller) permit(ctx *evalCtx, ref ruleRef, m *mergedView, filterPreds []map[string]any) *models.Response {
	var obligations []map[string]any
	if ref.rule != nil {
		for _, d := range ref.rule.Duties {
			obligations = append(obligations, dutyToObligation(d))
		}
	}
	seen := map[string]bool{}
	for _, props := range filterPreds {
		key := fmt.Sprintf("%v", props)
		if seen[key] {
			continue
		}
		seen[key] = true
		obligations = append(obligations, map[string]any{
			"id":         geonlFilterFeaturesAction,
			"properties": props,
		})
	}

	attrs := map[string]any{
		"decision_reason": "permitted by policy",
		"policy_uid":      ref.policy.UID,
	}
	if len(obligations) > 0 {
		attrs["obligations"] = obligations
	}

	return &models.Response{
		Allowed:    true,
		Message:    fmt.Sprintf("permitted by %s", ref.policy.UID),
		PolicyKey:  ref.policy.UID,
		Attributes: attrs,
	}
}

func deny(policyUID, reason string) *models.Response {
	return &models.Response{
		Allowed:   false,
		Message:   reason,
		PolicyKey: policyUID,
		Attributes: map[string]any{
			"decision_reason": reason,
			"policy_uid":      policyUID,
		},
	}
}

// dutyToObligation flattens a duty to an AuthZEN obligation (spec.md §7.3 /
// pseudocode map_duty).
func dutyToObligation(d *geoDuty) map[string]any {
	props := map[string]any{}
	for _, r := range d.Refinements {
		props[r.Key] = coerceValue(r.Value)
	}
	return map[string]any{"id": d.Action, "properties": props}
}

// --- matching helpers ---

func actionMatches(r *geoRule, ctx *evalCtx) bool {
	if r.Action == "" {
		return true
	}
	return strings.EqualFold(localName(r.Action), ctx.action)
}

func ruleMatchesLayer(r *geoRule, layer string, accessURL map[string]string) bool {
	if layer == "" {
		return true // no layer supplied -> do not filter by target.
	}
	for _, t := range r.Targets {
		if t == layer {
			return true
		}
		if url, ok := accessURL[t]; ok && url == layer {
			return true
		}
	}
	return false
}

func purposeMatch(pr *purposeRefinement, ctx *evalCtx) bool {
	if pr == nil {
		return true
	}
	if ctx.purpose == "" {
		return false
	}
	for _, v := range pr.Values {
		if v == ctx.purpose ||
			strings.EqualFold(localName(v), localName(ctx.purpose)) ||
			strings.Contains(ctx.purpose, localName(v)) {
			return true
		}
	}
	return false
}

func shortReason(r *geoRule) string {
	var parts []string
	for _, cons := range r.Constraints {
		if cons.isSpatial() {
			parts = append(parts, localName(cons.Operator)+" "+localName(firstNonEmpty(cons.RightRef, cons.RightGeometry)))
		} else {
			parts = append(parts, attributeKey(cons.Property)+" "+localName(cons.Operator))
		}
	}
	return strings.Join(parts, ", ")
}
