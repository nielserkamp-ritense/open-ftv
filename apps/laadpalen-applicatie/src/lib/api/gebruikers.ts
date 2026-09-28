import { useQuery } from '@tanstack/react-query'
import { apiClient } from './client'
import { queryKeys } from './queryKeys'
import type {
  Gebruiker,
  ApiListResponse,
} from './types'

// Query Hooks

export function useGebruikers() {
  return useQuery({
    queryKey: queryKeys.gebruikers.list(),
    queryFn: async () => {
      const response = await apiClient.post<Gebruiker[] | ApiListResponse<Gebruiker>>('/v1/gebruiker', {})
      // Handle both direct array response and wrapped response
      if (Array.isArray(response)) {
        return response
      }
      return response.data || []
    },
  })
}
