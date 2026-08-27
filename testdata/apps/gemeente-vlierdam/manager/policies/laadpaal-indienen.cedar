// Only Burgerzaken employees with a laadpaal-opleiding completed within the last 365 days
// may submit a laadpaal aanvraag. Subject attributes (afdeling, laadpaal-opleiding) must be
// supplied in the request's subject.properties - they are not resolved from PIP data.
forbid (
    principal,
    action == name::"indienen",
    resource is laadpaalAanvraag
)
unless {
    principal.afdeling == "Burgerzaken" &&
    principal has "laadpaal-opleiding" &&
    datetime(context.now).durationSince(datetime(principal["laadpaal-opleiding"])) <= duration("365d")
};
