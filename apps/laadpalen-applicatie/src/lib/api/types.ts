// API Response Types

export interface Aanvraag {
    id: number
    naam: string
    huisnummer: number
    postcode: string
    status: number
    omschrijving: string
    "status-datum": string
    "afgekeurd-reden"?: string
}

export interface Laadpaal {
    id: string

    [key: string]: unknown
}

export interface BrpPerson {
    'a-nummer'?: string
    afnemer?: Array<{ afnemer?: string }>
    bsn?: string
    datum?: string
    email?: string
    geboortedatum?: string
    geboorteland?: string
    geboorteplaats?: string
    geheim?: string | boolean
    gemeente?: string
    geslacht?: string
    geslachtsnaam?: string
    huisnummer?: number
    kind?: Array<{ 'bsn-kind'?: string; 'register-gemeente'?: string }>
    leeftijd?: number
    nationaliteit?: Array<{ nationaliteit?: string; volgnummer?: number }>
    postcode?: string
    reisdocument?: Array<Record<string, unknown>>
    senior?: boolean
    straatnaam?: string
    telefoon?: string
    toevoeging?: string
    verbintenis?: Array<{ land?: string; plaats?: string; volgnummer?: number }>
    volwassen?: boolean
    voornamen?: string
    woonplaats?: string

    [key: string]: unknown
}

export interface RdwVehicle {
    kenteken: string
    merk: string
    klasse: string
    diplomatiek?: boolean
    bsn: string
}

export interface Gebruiker {
    id: number
    naam: string
    afdeling: string
    wachtwoord: string
    "laadpaal-opleiding"?: string
}

// API Request Types
export interface CreateAanvraagRequest {
    postcode: string
    huisnummer: number
    status: number
    naam: string
    omschrijving: string
    "status-datum": string
    id: number
}

export interface PatchAanvraagRequest {
    [key: string]: unknown
}

// API Response Wrappers
export interface ApiListResponse<T> {
    data: T[]
}
