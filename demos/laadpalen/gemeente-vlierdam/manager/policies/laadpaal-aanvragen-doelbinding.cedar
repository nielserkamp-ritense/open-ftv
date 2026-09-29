// A laadpaal aanvraag may only be submitted under the "laadpaal-aanvraag" doelbinding.
@reason("doelbinding")
forbid (
    principal,
    action == name::"aanvragen",
    resource is laadpaalAanvraag
)
unless {
    context has doelbinding &&
    context.doelbinding == "laadpaal-aanvraag"
};
