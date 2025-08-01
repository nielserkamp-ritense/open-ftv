package bundles

func (r *runner) bundling() {
	r.debug("bundle stage started")

	r.createBundles("bundle")

	if r.advance("bundle") {
		r.info("bundle stage statistics", "bundles", r.bundleCount, "initialized", r.bundledCount)
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
