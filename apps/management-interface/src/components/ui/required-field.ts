import type React from 'react'

// The native validation bubble for an empty required field is rendered in the browser's own UI
// language, which no attribute (lang, title, ...) can override. setCustomValidity is the only
// standard way to force Dutch, so the Input, Select and Textarea controls wire the two handlers
// below in themselves and forms only have to set `required`.
export const REQUIRED_FIELD_MESSAGE = 'Vul dit veld in.'

type ValidatableElement = HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement

// markRequiredFieldInvalid overrides the bubble text for a field left empty. Every other
// constraint (pattern, maxLength, type="email", ...) keeps the browser's own message, which is
// more specific than REQUIRED_FIELD_MESSAGE would be.
export function markRequiredFieldInvalid(e: React.FormEvent<ValidatableElement>): void {
    const field = e.currentTarget
    if (field.validity.valueMissing) {
        field.setCustomValidity(REQUIRED_FIELD_MESSAGE)
    }
}

// clearRequiredFieldValidity drops the message set above once the field changes, since a custom
// validity keeps a field invalid until it is cleared again.
export function clearRequiredFieldValidity(e: React.FormEvent<ValidatableElement>): void {
    e.currentTarget.setCustomValidity('')
}
