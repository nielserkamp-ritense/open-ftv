// Maximum lengths enforced by the backend field checks, mirrored here so the Dutch
// messages below can't drift from the actual limits (see eam/handlers/fiber/policies.go
// and eam/handlers/fiber/common_checks.go for the source of truth).
const POLICY_LANGUAGE_MAX_LENGTH = 40
const POLICY_RVVA_ID_MAX_LENGTH = 80
const POLICY_URL_MAX_LENGTH = 400
const SETTINGS_TITLE_MAX_LENGTH = 200
const COMMON_STATUS_MAX_LENGTH = 40
const COMMON_TITLE_MAX_LENGTH = 80
const COMMON_ATTR_TYPE_MAX_LENGTH = 40
const COMMON_ENTITY_TYPE_MAX_LENGTH = 80
const COMMON_ATTR_KEY_MAX_LENGTH = 200
const COMMON_TAG_MAX_LENGTH = 40

// Dutch messages for backend error codes.
const ERROR_MESSAGES: Record<string, string> = {
    // Policy messages (see eam/handlers/fiber/policies.go)
    E02005: 'Regeltaal is verplicht.',
    E02010: `Regeltaal is te lang (maximaal ${POLICY_LANGUAGE_MAX_LENGTH} tekens).`,
    E02015: `Register van verwerkingsactiviteiten ID is te lang (maximaal ${POLICY_RVVA_ID_MAX_LENGTH} tekens).`,
    E02020: 'URL of broncode moet worden ingevuld.',
    E02025: 'Vul óf URL óf broncode in, niet beide.',
    E02030: `URL is te lang (maximaal ${POLICY_URL_MAX_LENGTH} tekens).`,
    E02035: 'ID in URL en body komen niet overeen.',

    // Entity messages (see eam/handlers/fiber/entities.go)
    E03005: 'Type in URL en body komen niet overeen.',
    E03010: 'ID in URL en body komen niet overeen.',

    // Attribute messages (see eam/handlers/fiber/attributes.go)
    E04005: 'Key in URL en body komen niet overeen.',

    // Tag messages (see eam/handlers/fiber/tags.go)
    E05005: 'ID van het beslispunt in URL en body komen niet overeen.',

    // Settings messages
    E08005: 'Titel is verplicht.',
    E08010: `Titel is te lang (maximaal ${SETTINGS_TITLE_MAX_LENGTH} tekens).`,
    E08015: 'Kleur van de header is verplicht.',
    E08020: 'Kleur van de header moet een geldige kleurcode zijn, bijv. #3B82F6.',
    E08025: 'Kleur van de titel is verplicht.',
    E08030: 'Kleur van de titel moet een geldige kleurcode zijn, bijv. #fcfcfc.',
    E08035: 'Logo moet een PNG, JPEG, SVG of WebP-afbeelding zijn.',

    // Shared field-check messages (attributen, beleidsregels, entiteiten, tags; see
    // eam/handlers/fiber/common_checks.go)
    E95005: `Status is te lang (maximaal ${COMMON_STATUS_MAX_LENGTH} tekens).`,
    E95010: 'Ongeldige status.',
    E95015: `Titel is te lang (maximaal ${COMMON_TITLE_MAX_LENGTH} tekens).`,
    E95020: `Type is te lang (maximaal ${COMMON_ATTR_TYPE_MAX_LENGTH} tekens).`,
    E95025: `Type is te lang (maximaal ${COMMON_ENTITY_TYPE_MAX_LENGTH} tekens).`,
    E95030: 'Key is verplicht.',
    E95035: `Key is te lang (maximaal ${COMMON_ATTR_KEY_MAX_LENGTH} tekens).`,
    E95040: 'Beslispunt mag niet leeg zijn.',
    E95045: `Beslispunt is te lang (maximaal ${COMMON_TAG_MAX_LENGTH} tekens).`,
}

// Monaco/textarea source-code fields don't enforce "required" the way native inputs do (a
// browser's required check treats whitespace-only text as filled in), so policy forms check
// and report this themselves before submitting.
export const SOURCE_REQUIRED_MESSAGE = 'Vul de broncode in.'

const DEFAULT_ERROR_MESSAGE = 'Er is een onverwachte fout opgetreden.'

// Dutch messages for generic errors that carry no Code, keyed by HTTP status instead.
const STATUS_MESSAGES: Record<number, string> = {
    400: 'Ongeldig verzoek.',
    401: 'Niet ingelogd.',
    403: 'Niet geautoriseerd.',
    404: 'Niet gevonden.',
    409: 'Conflict: de gegevens zijn ondertussen gewijzigd.',
    500: 'Er is een onverwachte fout opgetreden.',
}

// lookupErrorMessage looks up the message for a failed API request: by Code first then by HTTP status.
export function lookupErrorMessage(err: unknown): string {
    const code = errorCodeFrom(err)
    if (code && ERROR_MESSAGES[code]) {
        return ERROR_MESSAGES[code]
    }

    const status = errorStatusFrom(err)
    if (status !== undefined && STATUS_MESSAGES[status]) {
        return STATUS_MESSAGES[status]
    }

    return DEFAULT_ERROR_MESSAGE
}

// errorCodeFrom extracts the backend's error code from a failed API request, if present.
export function errorCodeFrom(err: unknown): string | undefined {
    if (typeof err !== 'object' || err === null || !('response' in err)) {
        return undefined
    }

    const response = (err as { response?: { data?: { code?: unknown } } }).response
    const code = response?.data?.code

    return typeof code === 'string' ? code : undefined
}

// errorStatusFrom extracts the backend's HTTP status from a failed API request, if present.
export function errorStatusFrom(err: unknown): number | undefined {
    if (typeof err !== 'object' || err === null || !('response' in err)) {
        return undefined
    }

    const response = (err as { response?: { data?: { status?: unknown } } }).response
    const status = response?.data?.status

    return typeof status === 'number' ? status : undefined
}
