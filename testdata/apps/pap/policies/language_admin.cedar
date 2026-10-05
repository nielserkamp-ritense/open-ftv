permit (
    principal is user,
    action,
    resource is language
)
when {
    principal has roles &&
    (principal.roles == "admin" || principal.roles.contains("admin"))
};
