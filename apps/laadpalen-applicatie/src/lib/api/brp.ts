import { useQuery } from '@tanstack/react-query'
import { brpApiClient } from './client'
import { queryKeys } from './queryKeys'
import { reattributeChainError } from './errors'
import type { BrpPerson } from './types'

const BRP_CHAIN_SOURCES = { inway: 'openftv-rvig-inway', pdp: 'openftv-rvig-pdp1' }

// GET /haalcentraal/api/brp/personen
export function useBrpPersons<TData = BrpPerson[]>(
  options?: { enabled?: boolean; select?: (data: BrpPerson[]) => TData, params?: Record<string, string | number | boolean> },
) {
  return useQuery<BrpPerson[], unknown, TData>({
    queryKey: queryKeys.brp.list(options?.params),
    enabled: options?.enabled,
    queryFn: async () => {
      try {
        const response = await brpApiClient.post<BrpPerson[]>(
          '/v1/haalcentraal/api/brp/personen',
          options?.params
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
export function useBrpAddress(postcode: string, huisnummer: number, options?: { enabled?: boolean }) {
  return useBrpPersons<BrpPerson[]>(
    {
      enabled: options?.enabled,
      params: {
        "@filter": "adres.postcode=" + postcode + ",adres.huisnummer=" + huisnummer
      }
    }
  )
}
