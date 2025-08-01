package bundles

func (r *runner) creating() {
	r.debug("create stage started")

	r.initCountsAndSlices()

	if r.advance("create") {
		r.info("create stage statistics", "bundles", r.bundleCount, "targets", r.targetCount)
	}
}

func (r *runner) initCountsAndSlices() {
	r.bundles = make(map[string]*Bundle, len(r.m.bundles))
	for k := range r.m.bundles {
		cfg := r.m.bundles[k]

		r.bundleCount++
		r.targetCount += uint64(len(cfg.Targets))
		r.bundles[k] = NewBundle(r.d.version, cfg.Language, cfg.Tags...)
	}
}
