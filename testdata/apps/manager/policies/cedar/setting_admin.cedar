// The systeembeheerder (admin) manages the settings (SCRUM-16 matrix: "Instellingen CUD").
permit (
    principal is user,
    action in [Action::"create", Action::"update", Action::"delete"],
    resource is setting
)
when {
    principal has roles &&
    (principal.roles == "admin" || principal.roles.contains("admin"))
};
