import { useQuery } from '@tanstack/react-query'
import { brpApiClient } from './client'
import { queryKeys } from './queryKeys'
import { reattributeChainError } from './errors'
import { traceparent } from './trace'
import type { BrpPerson } from './types'

const BRP_CHAIN_SOURCES = { inway: 'openftv-rvig-inway', pdp: 'openftv-rvig-pdp1' }

// GET /haalcentraal/api/brp/personen
export function useBrpPersons<TData = BrpPerson[]>(
  traceId: string | null,
  options?: { enabled?: boolean; select?: (data: BrpPerson[]) => TData, params?: Record<string, string | number | boolean> },
) {
  return useQuery<BrpPerson[], unknown, TData>({
    queryKey: [...queryKeys.brp.list(options?.params), traceId],
    enabled: (options?.enabled ?? true) && !!traceId,
    // A run keeps its trace ID, so it never runs again by itself, such as when the tab regains focus. Unlike
    // Infinity, 'static' also holds for a run that failed.
    staleTime: 'static',
    queryFn: async () => {
      try {
        const response = await brpApiClient.post<BrpPerson[]>(
          '/v1/haalcentraal/api/brp/personen',
          options?.params,
          { headers: { traceparent: traceparent(traceId!) } }
        )
        // API returns a plain array
        return response
      } catch (err) {
        reattributeChainError(err, BRP_CHAIN_SOURCES)
      }
    },
    select: options?.select,
  })
}

// Convenience hook to fetch persons by address
export function useBrpAddress(postcode: string, huisnummer: number, traceId: string | null, options?: { enabled?: boolean }) {
  return useBrpPersons<BrpPerson[]>(
    traceId,
    {
      enabled: options?.enabled,
      params: {
        "@filter": "adres.postcode=" + postcode + ",adres.huisnummer=" + huisnummer
      }
    }
  )
}
