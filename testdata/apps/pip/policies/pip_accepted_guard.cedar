// An accepted attribute or entity is not edited in place: revert it to concept first, or
// restore an older version. Accepted means "what the next deployment ships", so an edit
// without a new acceptance would change the shipped data unnoticed (SCRUM-145, decided
// 2026-09-29). Restore is its own action and is not caught here.
forbid (
    principal,
    action == Action::"update",
    resource
)
when {
    (resource is attribute || resource is entity) &&
    resource has status && resource.status != "concept"
};
