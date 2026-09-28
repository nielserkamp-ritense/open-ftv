import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'
import { queryKeys } from './queryKeys'
import type {
  Aanvraag,
  ApiListResponse,
  CreateAanvraagRequest,
  UpdateAanvraagRequest,
  PatchAanvraagRequest,
} from './types'

// Query Hooks

export function useAanvragen(params?: Record<string, string | number | boolean>) {
  return useQuery({
    queryKey: queryKeys.aanvragen.list(params),
    queryFn: async () => {
      const response = await apiClient.get<Aanvraag[] | ApiListResponse<Aanvraag>>('/v1/aanvragen', {
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

// Mutation Hooks

export function useCreateAanvraag() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: CreateAanvraagRequest }) => {
      return apiClient.post<Aanvraag>(`/v1/aanvraag/${id}`, data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.aanvragen.all })
    },
  })
}

export function useUpdateAanvraag() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateAanvraagRequest }) => {
      return apiClient.put<Aanvraag>(`/v1/aanvraag/${id}`, { id: id, ...data})
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.aanvragen.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.aanvragen.detail(variables.id) })
    },
  })
}

export function usePatchAanvraag() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: PatchAanvraagRequest }) => {
      return apiClient.patch<Aanvraag>(`/v1/aanvraag/${id}`, { id: id, ...data})
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.aanvragen.all })
      queryClient.invalidateQueries({ queryKey: queryKeys.aanvragen.detail(variables.id) })
    },
  })
}

export function useDeleteAanvraag() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      return apiClient.delete<void>(`/v1/aanvraag/${id}`)
    },
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.aanvragen.all })
      queryClient.removeQueries({ queryKey: queryKeys.aanvragen.detail(id) })
    },
  })
}
