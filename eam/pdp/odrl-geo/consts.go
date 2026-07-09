package odrl_geo

// RDF vocabulary constants for the ODRL-Geo-NL profile and the GeoSPARQL terms
// it reuses. These are owned by this engine module: the shared ODRL-AP-NL
// parser deliberately does not model geo-specific predicates, so this engine
// interprets them itself (a small geo interpretation layer on top of the reused
// parser / raw graph).

// Namespaces.
const (
	geonlNS = "https://standaarden.overheid.nl/odrl-geo-nl/"
	geoNS   = "http://www.opengis.net/ont/geosparql#"
	geofNS  = "http://www.opengis.net/def/function/geosparql/"
	odrlNS  = "http://www.w3.org/ns/odrl/2/"
	rdfNS   = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
)

// geonl left operands, classes and measures.
const (
	geonlFeatureProperty      = geonlNS + "featureProperty"
	geonlProperty             = geonlNS + "property"
	geonlFeatureGeometry      = geonlNS + "featureGeometry"
	geonlFeatureGeometryProp  = geonlNS + "featureGeometryProperty"
	geonlDefaultCRS           = geonlNS + "defaultCRS"
	geonlArea                 = geonlNS + "Area"
	geonlLayerSelection       = geonlNS + "LayerSelection"
	geonlFromLayer            = geonlNS + "fromLayer"
	geonlWhere                = geonlNS + "where"
	geonlGeoMaatregel         = geonlNS + "GeoMaatregel"
	geonlFilterFeaturesAction = geonlNS + "filterFeatures"
)

// GeoSPARQL geometry predicates.
const (
	geoHasGeometry = geoNS + "hasGeometry"
	geoAsWKT       = geoNS + "asWKT"
)

// ODRL predicates and operators used by the geo interpretation layer.
const (
	odrlLeftOperand           = odrlNS + "leftOperand"
	odrlOperator              = odrlNS + "operator"
	odrlRightOperand          = odrlNS + "rightOperand"
	odrlRightOperandReference = odrlNS + "rightOperandReference"
	odrlConstraintProp        = odrlNS + "constraint"
	odrlRefinement            = odrlNS + "refinement"
	odrlActionProp            = odrlNS + "action"
	odrlPurpose               = odrlNS + "purpose"
	rdfValue                  = rdfNS + "value"
)

// GeoSPARQL simple-features operators (typed as odrl:Operator by the profile).
const (
	sfWithin     = geofNS + "sfWithin"
	sfDisjoint   = geofNS + "sfDisjoint"
	sfCrosses    = geofNS + "sfCrosses"
	sfIntersects = geofNS + "sfIntersects"
	sfContains   = geofNS + "sfContains"
	sfOverlaps   = geofNS + "sfOverlaps"
	sfTouches    = geofNS + "sfTouches"
	sfEquals     = geofNS + "sfEquals"
)

// ODRL comparison operators for feature-property constraints.
const (
	opLt      = odrlNS + "lt"
	opLteq    = odrlNS + "lteq"
	opEq      = odrlNS + "eq"
	opNeq     = odrlNS + "neq"
	opGt      = odrlNS + "gt"
	opGteq    = odrlNS + "gteq"
	opIsAnyOf = odrlNS + "isAnyOf"
	opIsAllOf = odrlNS + "isAllOf"
	opIsA     = odrlNS + "isA"
)

// DefaultCRS is the ODRL-Geo-NL profile default coordinate reference system
// (EPSG:28992 / RD New), used when a geometry literal carries no CRS URI and
// the policy declares no geonl:defaultCRS (spec.md §5).
const DefaultCRS = "http://www.opengis.net/def/crs/EPSG/0/28992"

// Resource property keys (AuthZEN resource.properties, spec.md §7.1).
const (
	propLayer      = "layer"
	propCRS        = "crs"
	propGeometry   = "geometry"
	propBBox       = "bbox"
	propTile       = "tile"
	propAttributes = "attributes"
)

// Action property key carrying the ODRL-AP-NL purpose (processing activity).
const (
	actionProcessingActivity = "dpl.core.processing_activity_id"
	actionPurpose            = "purpose"
)

// Resource types (AuthZEN resource.type).
const (
	resourceFeature = "feature"
	resourceTile    = "tile"
	resourceLayer   = "layer"
)
