// The functioneel beheerder (author) edits, accepts, reverts and restores PIP data
// (SCRUM-16 matrix: "PIP CUD").
permit (
    principal is user,
    action in [Action::"create", Action::"update", Action::"delete", Action::"accept", Action::"revert", Action::"restore"],
    resource is entity
)
when {
    principal has roles &&
    (principal.roles == "author" || principal.roles.contains("author"))
};
