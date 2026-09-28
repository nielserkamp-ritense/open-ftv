import { useQuery } from '@tanstack/react-query'
import { rdwApiClient } from './client'
import { queryKeys } from './queryKeys'
import {RdwVehicle} from './types'

// RDW is read-only. The endpoint returns vehicle data for persons

// POST /haalcentraal/api/brv/voertuigen
// Request body: { bsns: string[] }
export function useRdwVehicles<TData = RdwVehicle[]>(
  options?: { enabled?: boolean; select?: (data: RdwVehicle[]) => TData, params?: Record<string, string | number | boolean> }
) {
  return useQuery<RdwVehicle[], unknown, TData>({
    queryKey: queryKeys.rdw.list(),
    enabled: options?.enabled,
    queryFn: async () => {
      const response = await rdwApiClient.post<RdwVehicle[]>(
        '/v1/voertuigen',
        options?.params
      )
      return response
    },
    select: options?.select,
  })
}

// Convenience hook to fetch vehicles by BSN list
// Uses the useRdwVehicles hook and filters on the frontend via React Query select
export function useRdwVehiclesByBsn(bsn: string, options?: { enabled?: boolean }) {
  return useRdwVehicles<RdwVehicle[]>({
    enabled: options?.enabled && bsn != "",
    params: {
      "@filter": "kenteken.bsn=" + bsn,
    }
  })
}

// Convenience hook to fetch vehicles for multiple BSNs at once. Once RDW shares diplomatic vehicles (demo
// scenario 2), the gateway only shows them to a Gemeentesecretaris, so it needs to know who's asking.
export function useRdwVehiclesForMultipleBsns(bsnList: string[], gebruiker: string | undefined, options?: { enabled?: boolean }) {
  return useQuery<RdwVehicle[], unknown, RdwVehicle[]>({
    queryKey: [...queryKeys.rdw.list(), 'multiple', bsnList, gebruiker],
    enabled: options?.enabled && bsnList.length > 0 && !!gebruiker,
    queryFn: async () => {
      const results = await Promise.all(
        bsnList.map(async (bsn) => {
          const response = await rdwApiClient.post<RdwVehicle[]>(
            '/v1/voertuigen',
            { "@filter": "kenteken.bsn=" + bsn },
            { headers: { 'Dpl-Core-User': gebruiker! } }
          )
          return response
        })
      )
      return results.flat()
    },
  })
}
