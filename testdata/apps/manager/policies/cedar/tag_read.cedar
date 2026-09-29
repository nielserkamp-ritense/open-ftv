// Every role reads beslispunten (SCRUM-16 matrix: "Beslispunten lezen").
// roles is a set for token users and may be a plain string for locally seeded users.
permit (
    principal is user,
    action == Action::"read",
    resource is tag
)
when {
    principal has roles &&
    (
        principal.roles == "admin" || principal.roles == "author" || principal.roles == "auditor" ||
        principal.roles.contains("admin") || principal.roles.contains("author") || principal.roles.contains("auditor")
    )
};
