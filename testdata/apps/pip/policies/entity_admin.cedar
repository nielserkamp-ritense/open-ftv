permit (
    principal is user,
    action,
    resource is entity
)
when {
    principal has roles &&
    (principal.roles == "admin" || principal.roles.contains("admin"))
};
