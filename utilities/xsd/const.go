package xsd

// Prefix is the well-known prefix for XSD.
const Prefix = "xsd"

// URI is the well-known URI for XSD.
const URI = "http://www.w3.org/2001/XMLSchema#"

// List of standard XSD identifiers.
const (
	Any              = "anyType"
	AnyURI           = "anyURI"
	Boolean          = "boolean"
	Byte             = "byte"
	Date             = "date"
	DateTime         = "dateTime"
	Day              = "gDay"
	Decimal          = "decimal"
	Double           = "double"
	Duration         = "duration"
	Float            = "float"
	Int              = "int"
	Integer          = "integer"
	Language         = "language"
	Long             = "long"
	Month            = "gMonth"
	MonthDay         = "gMonthDay"
	Neg              = "negativeInteger"
	NonNeg           = "nonNegativeInteger"
	NonPos           = "nonPositiveInteger"
	NormalizedString = "normalizedString"
	Pos              = "positiveInteger"
	Short            = "short"
	Simple           = "simpleType"
	String           = "string"
	Time             = "time"
	Token            = "token"
	UByte            = "unsignedByte"
	UInt             = "unsignedInt"
	ULong            = "unsignedLong"
	UShort           = "unsignedShort"
	Year             = "gYear"
	YearMonth        = "gYearMonth"
)

// List of standard XSD URI-identifiers.
const (
	URIAny        = URI + Any
	URIAnyURI     = URI + AnyURI
	URIBoolean    = URI + Boolean
	URIByte       = URI + Byte
	URIDate       = URI + Date
	URIDateTime   = URI + DateTime
	URIDay        = URI + Day
	URIDecimal    = URI + Decimal
	URIDouble     = URI + Double
	URIDuration   = URI + Duration
	URIFloat      = URI + Float
	URIInt        = URI + Int
	URIInteger    = URI + Integer
	URILanguage   = URI + Language
	URILong       = URI + Long
	URIMonth      = URI + Month
	URIMonthDay   = URI + MonthDay
	URINeg        = URI + Neg
	URINonNeg     = URI + NonNeg
	URINonPos     = URI + NonPos
	URINormalized = URI + NormalizedString
	URIPos        = URI + Pos
	URIShort      = URI + Short
	URISimple     = URI + Simple
	URIString     = URI + String
	URITime       = URI + Time
	URIToken      = URI + Token
	URIUByte      = URI + UByte
	URIUInt       = URI + UInt
	URIULong      = URI + ULong
	URIUShort     = URI + UShort
	URIYear       = URI + Year
	URIYearMonth  = URI + YearMonth
)

// List of standard XSD prefix-identifiers.
const (
	PrefixAny              = prefixColon + Any
	PrefixAnyURI           = prefixColon + AnyURI
	PrefixBoolean          = prefixColon + Boolean
	PrefixByte             = prefixColon + Byte
	PrefixDate             = prefixColon + Date
	PrefixDateTime         = prefixColon + DateTime
	PrefixDay              = prefixColon + Day
	PrefixDecimal          = prefixColon + Decimal
	PrefixDouble           = prefixColon + Double
	PrefixDuration         = prefixColon + Duration
	PrefixFloat            = prefixColon + Float
	PrefixInt              = prefixColon + Int
	PrefixInteger          = prefixColon + Integer
	PrefixLanguage         = prefixColon + Language
	PrefixLong             = prefixColon + Long
	PrefixMonth            = prefixColon + Month
	PrefixMonthDay         = prefixColon + MonthDay
	PrefixNeg              = prefixColon + Neg
	PrefixNonNeg           = prefixColon + NonNeg
	PrefixNonPos           = prefixColon + NonPos
	PrefixNormalizedString = prefixColon + NormalizedString
	PrefixPos              = prefixColon + Pos
	PrefixShort            = prefixColon + Short
	PrefixSimple           = prefixColon + Simple
	PrefixString           = prefixColon + String
	PrefixTime             = prefixColon + Time
	PrefixToken            = prefixColon + Token
	PrefixUByte            = prefixColon + UByte
	PrefixUInt             = prefixColon + UInt
	PrefixULong            = prefixColon + ULong
	PrefixUShort           = prefixColon + UShort
	PrefixYear             = prefixColon + Year
	PrefixYearMonth        = prefixColon + YearMonth
)

const prefixColon = Prefix + ":"
