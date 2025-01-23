package xsd

// IsLiteral returns true if the given identifier can be identified as a literal.
func IsLiteral(s string) bool {
	switch s {
	case URIAny, URIAnyURI, URIBoolean, URIByte, URIDate, URIDateTime, URIDay, URIDecimal, URIDouble,
		URIDuration, URIFloat, URIInt, URIInteger, URILanguage, URILong, URIMonth,
		URIMonthDay, URINeg, URINonNeg, URINonPos, URINormalized, URIPos,
		URIShort, URISimple, URIString, URITime, URIToken,
		URIUByte, URIUInt, URIULong, URIUShort, URIYear, URIYearMonth:
		return true
	case PrefixAny, PrefixAnyURI, PrefixBoolean, PrefixByte, PrefixDate, PrefixDateTime, PrefixDay, PrefixDecimal, PrefixDouble,
		PrefixDuration, PrefixFloat, PrefixInt, PrefixInteger, PrefixLanguage, PrefixLong, PrefixMonth,
		PrefixMonthDay, PrefixNeg, PrefixNonNeg, PrefixNonPos, PrefixNormalized, PrefixPos,
		PrefixShort, PrefixSimple, PrefixString, PrefixTime, PrefixToken,
		PrefixUByte, PrefixUInt, PrefixULong, PrefixUShort, PrefixYear, PrefixYearMonth:
		return true
	default:
		return false
	}
}
