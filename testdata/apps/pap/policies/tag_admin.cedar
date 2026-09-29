permit (
    principal is user,
    action,
    resource is tag
)
when {
    principal has roles &&
    (principal.roles == "admin" || principal.roles.contains("admin"))
};
