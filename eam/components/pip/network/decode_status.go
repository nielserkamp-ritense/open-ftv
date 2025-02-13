package network

func (r *runner) decodeStatus(dec *StatusCode, status int) bool {
	if !dec.Match(status) {
		return false
	}

	if dec.Attributes != nil {
		r.decodeAttribute(dec.Attributes)
	}
	if dec.Entity != nil {
		r.decodeEntity(dec.Entity)
	}
	if dec.Relation != nil {
		r.decodeRelation(dec.Relation)
	}

	return true
}
