permit (
    principal is user,
    action,
    resource is deployment
)
when {
    principal has roles &&
    (principal.roles == "admin" || principal.roles.contains("admin"))
};
