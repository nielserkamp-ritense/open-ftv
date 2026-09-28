import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'
import {apiClient} from './client'
import {queryKeys} from './queryKeys'
import type {ApiListResponse, BrpPerson, CreateLaadpaalRequest, Laadpaal, UpdateLaadpaalRequest,} from './types'

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

// Mutation Hooks

export function useCreateLaadpaal() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: CreateLaadpaalRequest }) => {
      return apiClient.post<Laadpaal>(`/v1/laadpaal/${id}`, data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.laadpalen.all })
    },
  })
}

export function useUpdateLaadpaal() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateLaadpaalRequest }) => {
      return apiClient.put<Laadpaal>(`/v1/laadpaal/${id}`, data)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.laadpalen.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.laadpalen.detail(variables.id) })
    },
  })
}

export function useDeleteLaadpaal() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      return apiClient.delete<void>(`/v1/laadpaal/${id}`)
    },
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.laadpalen.all })
      queryClient.removeQueries({ queryKey: queryKeys.laadpalen.detail(id) })
    },
  })
}
