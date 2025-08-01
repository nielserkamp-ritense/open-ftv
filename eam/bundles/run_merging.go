package bundles

func (r *runner) merging() {
	r.debug("merge stage started")

	r.mergeToGIT("merge")

	if r.advance("merge") {
		r.info("merge stage statistics")
	}
}

func (r *runner) mergeToGIT(stage string) {
	if r.policies == nil && r.attributes == nil && r.entities == nil && r.relations == nil {
		r.gatherLists(stage)
	}

	// TODO: ...

	r.gitHash = "xyz" // TODO: real GIT hash!
}
