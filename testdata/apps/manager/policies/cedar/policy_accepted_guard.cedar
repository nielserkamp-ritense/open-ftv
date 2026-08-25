// An accepted or deployed policy is not edited in place: restore an older version, or
// accept a new concept. Restore is its own action and is not caught here.
forbid (
    principal,
    action == Action::"update",
    resource is policy
)
when {
    resource has status && resource.status != "concept"
};
