import {useQuery} from '@tanstack/react-query'
import {apiClient} from './client'
import {queryKeys} from './queryKeys'
import type {Laadpaal} from './types'

// Query Hooks

export function useLaadpalen<TData = Laadpaal[]>(
  params?: Record<string, string | number | boolean>,
  options?: { enabled?: boolean }
) {
  return useQuery<Laadpaal[], unknown, TData>({
    queryKey: queryKeys.laadpalen.list(params),
    queryFn: async () => {
      return await apiClient.post<Laadpaal[]>('/v1/laadpalen',
        params
      )
    },
    enabled: options?.enabled,
  })
}
