// A Burgerzaken employee must have completed the laadpaal-opleiding.
@reason("opleiding")
forbid (
    principal,
    action == name::"aanvragen",
    resource is laadpaalAanvraag
)
when {
    context has doelbinding &&
    context.doelbinding == "laadpaal-aanvraag" &&
    principal has afdeling &&
    principal.afdeling == "Burgerzaken"
}
unless {
    principal has "laadpaal-opleiding"
};
