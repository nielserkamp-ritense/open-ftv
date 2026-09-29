// A PDP fetches its bundle from the last deployment with an API key: read only.
permit (
    principal is app,
    action == Action::"read",
    resource is deployment
);
