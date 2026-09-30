import { vlierdamPdp1Client } from './client'
import { VITE_DOELBINDING } from '@/lib/env.ts'

interface AuthzenEntity {
  type: string
  id: string
  properties?: Record<string, unknown>
}

interface AuthzenEvaluationRequest {
  subject: AuthzenEntity
  resource: AuthzenEntity
  action: { name: string; properties?: Record<string, unknown> }
  context?: Record<string, unknown>
}

interface AuthzenEvaluationResponse {
  decision: boolean
  context?: {
    reason?: string
    [key: string]: unknown
  }
}

// PDP1 (Cedar) - Scenario 1: may `gebruiker` submit a laadpaal aanvraag at all. Verified
// empirically against the running stack: pdp1's AuthZEN /evaluation handler builds the Cedar
// principal straight from this request's subject.properties - it does not resolve PIP-pulled
// entity data (that only applies to entities pre-published as PAP "data", not runtime PIP
// pulls). So afdeling/laadpaal-opleiding must be sent here.
export async function evaluateLaadpaalIndienen(params: {
  gebruiker: string
  afdeling?: string
  laadpaalOpleiding?: string
  postcode: string
  huisnummer: number
}): Promise<AuthzenEvaluationResponse> {
  const { gebruiker, afdeling, laadpaalOpleiding, postcode, huisnummer } = params
  return vlierdamPdp1Client!.post<AuthzenEvaluationResponse>('/authzen/v1/evaluation', {
    subject: {
      type: 'medewerker',
      id: gebruiker,
      properties: { afdeling, 'laadpaal-opleiding': laadpaalOpleiding },
    },
    resource: { type: 'laadpaalAanvraag', id: `${postcode}-${huisnummer}` },
    action: { name: 'aanvragen' },
    context: { now: new Date().toISOString(), doelbinding: VITE_DOELBINDING },
  } satisfies AuthzenEvaluationRequest)
}
