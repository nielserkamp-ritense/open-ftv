// The laadpaal-opleiding must have been completed within the last 365 days.
@reason("verlopen opleiding")
forbid (
    principal,
    action == name::"aanvragen",
    resource is laadpaalAanvraag
)
when {
    context has doelbinding &&
    context.doelbinding == "laadpaal-aanvraag" &&
    principal has afdeling &&
    principal.afdeling == "Burgerzaken" &&
    principal has "laadpaal-opleiding"
}
unless {
    context has now &&
    datetime(context.now).durationSince(datetime(principal["laadpaal-opleiding"])) <= duration("365d")
};
