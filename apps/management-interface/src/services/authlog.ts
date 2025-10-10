import { pipRequest as request } from '../api/client/pip.ts';
import { components, operations } from '../oas/authlog';
import { useQuery } from '@tanstack/react-query';

// Types for the authlog service
export type AuthlogEntry = components['schemas']['AuthlogEntry'];
export type AuthlogEntriesResponse = components['schemas']['AuthlogEntries'];
export type AuthlogQueryParams = operations['get-adl-entries']['parameters']['query'];

const AUTHLOG_QUERY_KEYS = {
  all: ['authlog'] as const,
  lists: () => [...AUTHLOG_QUERY_KEYS.all, 'list'] as const,
  list: (filters: Record<string, unknown>) => [...AUTHLOG_QUERY_KEYS.lists(), filters] as const,
  details: () => [...AUTHLOG_QUERY_KEYS.all, 'detail'] as const,
  detail: (id: number | string) => [...AUTHLOG_QUERY_KEYS.details(), id] as const,
};

/**
 * Service for interacting with the Authorization Decision Log (authlog) API endpoints
 */
export const authlogService = {
  /**
   * Query Authorization Decision Log entries
   * @param params - Optional query parameters to filter results
   * @returns Promise with a list of authlog entries
   */
  getAuthlogEntries: async (
    params?: AuthlogQueryParams
  ): Promise<AuthlogEntriesResponse> => {
    return request<AuthlogEntriesResponse>({
      method: 'GET',
      url: '/v1/adl/entries',
      params,
    });
  },
};

/**
 * Hook to fetch Authorization Decision Log entries (optionally filtered)
 */
export const useAuthlogEntries = (filters?: AuthlogQueryParams) => {
  return useQuery({
    queryKey: AUTHLOG_QUERY_KEYS.list(filters ?? {}),
    queryFn: () => authlogService.getAuthlogEntries(filters),
  });
};

/**
 * Hook to fetch a single Authorization Decision Log entry by id
 * Note: This uses the same endpoint under the hood with the `id` filter
 */
export const useAuthlogEntry = (id?: number) => {
  return useQuery({
    queryKey: AUTHLOG_QUERY_KEYS.detail(id ?? 'undefined'),
    queryFn: async () => {
      if (id == null) return undefined as unknown as AuthlogEntry | undefined;
      const entries = await authlogService.getAuthlogEntries({ id });
      return entries?.[0];
    },
    enabled: id != null,
  });
};
