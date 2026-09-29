permit (
    principal is user,
    action,
    resource is attribute
)
when {
    principal has roles &&
    (principal.roles == "admin" || principal.roles.contains("admin"))
};
