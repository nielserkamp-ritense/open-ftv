permit (
    principal == activity::"31953ef2-5a52-421a-9d2d-2c11e6dcb86f",
    action == action::"POST",
    resource is service
)
when {
    principal has brpPersonen &&
    resource.code == "BRP" && resource.owner == "RvIG" &&
    context.body has type && principal.brpPersonen.types.contains(context.body.type) &&
    context.body has fields &&
    context.body.type like "Raadpleeg*" && principal.brpPersonen.raadplegen.allowedFields.containsAll(context.body.fields)
};
