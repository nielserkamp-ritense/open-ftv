permit (
    principal,
    action == action::"GET",
    resource is service
)
when {
    resource.code == "FRK" && resource.owner == "fictief" && context.http.path like "/voertuig/*" &&
    context has doelbinding && context.doelbinding == "eigenaarschap"
};
