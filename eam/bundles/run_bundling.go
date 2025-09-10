package bundles

func (r *runner) bundling() {
	r.debug("bundle stage started")

	r.createBundles("bundle")

	if r.createBundleAudit() {
		if r.advance("bundle") {
			r.info("bundle stage statistics", "bundles", r.bundleCount, "initialized", r.bundledCount)
		}
	}
}

func (r *runner) createBundles(stage string) {
	if r.gitHash == "" {
		r.mergeToGIT(stage)
	}

	for _, b := range r.bundles {
		for i := range r.policies {
			b.AddPolicy(r.policies[i])
		}
		for i := range r.attributes {
			b.AddAttribute(r.attributes[i])
		}
		for i := range r.entities {
			b.AddEntity(r.entities[i])
		}
		for i := range r.relations {
			b.AddRelation(r.relations[i])
		}

		r.bundledCount++
	}
}

func (r *runner) createBundleAudit() bool {
	for k, b := range r.bundles {
		cfg := r.m.bundles[k]
		version := r.d.version

		if err := r.handler.CreateBundleAudit(r.ctx, version, cfg, b); err != nil {
			r.error("failed to create bundle audit", "version", version, "bundle", cfg.ID, "error", err)
			return false
		}
	}

	return true
}
