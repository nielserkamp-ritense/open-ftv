import { request } from '../api/client/base';
import { components } from '../oas/policies';
import { useQuery } from '@tanstack/react-query';

// Types for the tags service
export type Tag = components['schemas']['Tag'];
export type TagsResponse = components['schemas']['Tags'];

const TAGS_QUERY_KEYS = {
  all: ['tags'] as const,
  lists: () => [...TAGS_QUERY_KEYS.all, 'list'] as const,
  list: (filters: Record<string, unknown>) => [...TAGS_QUERY_KEYS.lists(), filters] as const,
};

/**
 * Service for interacting with the tags API endpoints
 */
export const tagsService = {
  /**
   * Get all tags
   * @returns Promise with all tags
   */
  getAllTags: async (): Promise<TagsResponse> => {
    return request<TagsResponse>({
      method: 'GET',
      url: '/v1/tags',
    });
  },
};

/**
 * Hook to fetch all tags
 */
export const useTags = () => {
  return useQuery({
    queryKey: TAGS_QUERY_KEYS.lists(),
    queryFn: tagsService.getAllTags,
  });
};
