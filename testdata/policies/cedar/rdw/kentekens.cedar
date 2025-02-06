permit (
    principal,
    action == action::"GET",
    resource is service
)
when {
    resource.code == "FRK" && resource.owner == "fictief" && context.http.path == "/voertuigen" &&
    context has doelbinding && context.doelbinding == "lijst"
};
