// API Error Handling Utilities

export class ApiError extends Error {
  constructor(
    message: string,
    // undefined when the request never got an HTTP response at all (e.g. connection refused).
    public status: number | undefined,
    public data?: unknown,
    // docker container name of the backend this request went to, for local debugging.
    public source?: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}

// Backends return an RFC9457-style body: { status, title, detail }.
export function getErrorMessage(error: unknown): string {
  if (isApiError(error)) {
    if (typeof error.data === 'object' && error.data !== null) {
      const data = error.data as Record<string, unknown>
      if ('detail' in data && typeof data.detail === 'string' && data.detail) {
        return data.detail
      }
      if ('message' in data && typeof data.message === 'string') {
        return data.message
      }
      if ('error' in data && typeof data.error === 'string') {
        return data.error
      }
    }
    return error.message
  }

  if (error instanceof Error) {
    return error.message
  }

  return 'An unknown error occurred'
}

// Dutch fallback per HTTP status, since backends only return a status and an English "detail".
const DUTCH_MESSAGE_BY_STATUS: Record<number, string> = {
  400: 'De aanvraag bevat onjuiste of onvolledige gegevens.',
  401: 'U bent niet aangemeld.',
  403: 'U heeft geen toegang tot deze gegevens.',
  404: 'De gevraagde gegevens zijn niet gevonden.',
  409: 'De gegevens zijn in strijd met bestaande gegevens.',
  500: 'Er is een interne fout opgetreden.',
  501: 'Deze functie wordt niet ondersteund.',
  503: 'De dienst is tijdelijk niet bereikbaar.',
}

const DEFAULT_DUTCH_MESSAGE = 'Er is een fout opgetreden bij het ophalen of opslaan van gegevens.'

export function getDutchErrorMessage(error: unknown): string {
  if (isApiError(error) && error.status !== undefined) {
    return DUTCH_MESSAGE_BY_STATUS[error.status] ?? DEFAULT_DUTCH_MESSAGE
  }
  return DEFAULT_DUTCH_MESSAGE
}

export function isNotFoundError(error: unknown): boolean {
  return isApiError(error) && error.status === 404
}

export function isUnauthorizedError(error: unknown): boolean {
  return isApiError(error) && error.status === 401
}

export function isForbiddenError(error: unknown): boolean {
  return isApiError(error) && error.status === 403
}

export function isValidationError(error: unknown): boolean {
  return isApiError(error) && (error.status === 400 || error.status === 422)
}

// A client that goes through an inway/PDP/dataspace chain carries only one static `source`
// label (the dataspace at the end of the chain) - but Kong can reject a request before it ever
// gets there. Re-labels the error to whichever hop actually produced it, so toasts/health
// displays point at the real failure point instead of always blaming the dataspace.
export function reattributeChainError(error: unknown, sources: { inway: string; pdp: string }): never {
  if (isApiError(error)) {
    // key-auth rejects at the inway itself, before the PDP or dataspace are ever reached.
    if (error.status === 401) {
      throw new ApiError(error.message, error.status, error.data, sources.inway)
    }
    // Kong returns a bare, bodyless 500 when its own call to the PDP fails.
    if (error.status === 500 && error.data === undefined) {
      throw new ApiError(error.message, error.status, error.data, sources.pdp)
    }
  }
  throw error
}
