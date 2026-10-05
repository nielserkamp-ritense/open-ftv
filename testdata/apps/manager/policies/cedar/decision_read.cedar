// Every role reads the Authorization Decision Log (ADR 0006: everyone reads, the ADL included).
// An API key does not: no decision_apikey rule exists on purpose (SCRUM-146).
// roles is a set for token users and may be a plain string for locally seeded users.
permit (
    principal is user,
    action == Action::"read",
    resource is decision
)
when {
    principal has roles &&
    (
        principal.roles == "admin" || principal.roles == "author" || principal.roles == "auditor" ||
        principal.roles.contains("admin") || principal.roles.contains("author") || principal.roles.contains("auditor")
    )
};
