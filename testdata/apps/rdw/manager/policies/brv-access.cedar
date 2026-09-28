// Only requests carrying the "laadinfra-verificatie" doelbinding may access RDW services.
permit (
    principal,
    action,
    resource is service
)
when {
    context has doelbinding &&
    context.doelbinding == "laadinfra-verificatie"
};
