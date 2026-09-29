permit (
    principal is user,
    action == Action::"read",
    resource is entity
)
when {
    principal has roles &&
    (principal.roles == "user" || principal.roles.contains("user"))
};
