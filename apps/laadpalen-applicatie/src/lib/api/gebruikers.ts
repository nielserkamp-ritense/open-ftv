import { useQuery } from '@tanstack/react-query'
import { apiClient } from './client'
import { queryKeys } from './queryKeys'
import type {
  Gebruiker,
  ApiListResponse,
} from './types'

// Query Hooks

export function useGebruikers(params?: Record<string, string | number | boolean>) {
  return useQuery({
    queryKey: queryKeys.gebruikers.list(params),
    queryFn: async () => {
      const response = await apiClient.post<Gebruiker[] | ApiListResponse<Gebruiker>>('/v1/gebruiker', {
        params,
      })
      // Handle both direct array response and wrapped response
      if (Array.isArray(response)) {
        return response
      }
      return response.data || []
    },
  })
}
