// Package management is the service layer of the management plane: one service per resource
// type, each method one business operation that asks the PDP exactly once, with the caller,
// the action and the stored object as the question, before touching the store.
//
// The order inside every operation is fixed: load the object, decide, then act. A missing
// object still gets a decision (with its id and no properties), and the decision is reported
// before the not-found, so an unauthorized caller cannot probe for ids. See
// docs/adr/0006-management-plane-authorization-model.md.
package management
