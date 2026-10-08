import { useQuery } from '@tanstack/react-query'
import { rdwApiClient } from './client'
import { queryKeys } from './queryKeys'
import {RdwVehicle} from './types'
import { traceparent } from './trace'

// RDW is read-only. The endpoint returns vehicle data for persons.

// Convenience hook to fetch vehicles for multiple BSNs at once. Once RDW shares diplomatic vehicles (demo
// scenario 2), the gateway only shows them to a Gemeentesecretaris, so it needs to know who's asking.
export function useRdwVehiclesForMultipleBsns(bsnList: string[], gebruiker: string | undefined, traceId: string | null, options?: { enabled?: boolean }) {
  return useQuery<RdwVehicle[], unknown, RdwVehicle[]>({
    queryKey: [...queryKeys.rdw.list(), 'multiple', bsnList, gebruiker, traceId],
    enabled: (options?.enabled ?? true) && bsnList.length > 0 && !!gebruiker && !!traceId,
    // A run keeps its trace ID, so it never runs again by itself, such as when the tab regains focus. Unlike
    // Infinity, 'static' also holds for a run that failed.
    staleTime: 'static',
    queryFn: async () => {
      const results = await Promise.all(
        bsnList.map(async (bsn) => {
          const response = await rdwApiClient.post<RdwVehicle[]>(
            '/v1/voertuigen',
            { "@filter": "kenteken.bsn=" + bsn },
            { headers: { 'Dpl-Core-User': gebruiker!, traceparent: traceparent(traceId!) } }
          )
          return response
        })
      )
      return results.flat()
    },
  })
}
