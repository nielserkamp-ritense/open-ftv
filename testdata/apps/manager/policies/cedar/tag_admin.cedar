// The systeembeheerder (admin) manages beslispunten (SCRUM-16 matrix: "Beslispunten CUD").
permit (
    principal is user,
    action in [Action::"create", Action::"update", Action::"delete"],
    resource is tag
)
when {
    principal has roles &&
    (principal.roles == "admin" || principal.roles.contains("admin"))
};
