// The admin, the author and the auditor read PIP data (SCRUM-16 matrix: "PIP lezen"). The
// `user` role is not among them: the manager's examples dropped it, because `auditor` already
// means "read everything" (ADR 0006).
// roles is a set for token users and may be a plain string for locally seeded users.
permit (
    principal is user,
    action == Action::"read",
    resource is attribute
)
when {
    principal has roles &&
    (
        principal.roles == "admin" || principal.roles == "author" || principal.roles == "auditor" ||
        principal.roles.contains("admin") || principal.roles.contains("author") || principal.roles.contains("auditor")
    )
};
