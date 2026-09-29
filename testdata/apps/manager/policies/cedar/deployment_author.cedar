// The functioneel beheerder (author) publishes deployments (SCRUM-16 matrix: "Regels (un)deploy").
permit (
    principal is user,
    action == Action::"deploy",
    resource is deployment
)
when {
    principal has roles &&
    (principal.roles == "author" || principal.roles.contains("author"))
};
