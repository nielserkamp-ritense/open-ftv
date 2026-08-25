// The functioneel beheerder (author) edits, accepts, reverts, deploys and restores policies
// (SCRUM-16 matrix: "Regels CUD" and "Regels (un)deploy").
permit (
    principal is user,
    action in [Action::"create", Action::"update", Action::"delete", Action::"accept", Action::"revert", Action::"deploy", Action::"restore"],
    resource is policy
)
when {
    principal has roles &&
    (principal.roles == "author" || principal.roles.contains("author"))
};
