permit (
    principal is user,
    action == Action::"read",
    resource is policy
)
when {
    principal has roles &&
    (principal.roles == "user" || principal.roles.contains("user"))
};
