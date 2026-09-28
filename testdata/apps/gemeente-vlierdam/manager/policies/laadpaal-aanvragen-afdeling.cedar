// Only Burgerzaken employees may submit a laadpaal aanvraag.
// Subject attributes (afdeling, laadpaal-opleiding) must be supplied in the request's
// subject.properties - they are not resolved from PIP data.
//
// Every laadpaal-aanvragen-* forbid also requires the checks of its predecessors to pass, so
// exactly one of them fires and its @reason is returned deterministically (cedar-go iterates
// policies in map order). Missing attributes are guarded with `has`: a policy that errors is
// skipped, and the bootstrap permit-all would then allow the request.
@reason("afdeling")
forbid (
    principal,
    action == name::"aanvragen",
    resource is laadpaalAanvraag
)
when {
    context has doelbinding &&
    context.doelbinding == "laadpaal-aanvraag"
}
unless {
    principal has afdeling &&
    principal.afdeling == "Burgerzaken"
};
