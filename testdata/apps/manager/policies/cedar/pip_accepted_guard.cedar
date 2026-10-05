// An accepted attribute or entity is not edited in place: revert it to concept first.
// Accepted means "what the next deployment ships", so an edit without a new acceptance would
// change the shipped data unnoticed (SCRUM-145, decided 2026-09-29).
//
// Deployed data is deliberately not caught: the PIP has no way back from deployed to concept,
// so an edit of a deployed attribute or entity is how its next concept starts (review on
// !330, decided 2026-10-05). The deployment keeps the shipped version.
forbid (
    principal,
    action == Action::"update",
    resource
)
when {
    (resource is attribute || resource is entity) &&
    resource has status && resource.status == "accepted"
};
