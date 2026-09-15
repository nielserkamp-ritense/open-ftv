permit (
    principal is user,
    action,
    resource is policy
)
when {
    principal has roles &&
    (principal.roles == "admin" || principal.roles.contains("admin"))
};
