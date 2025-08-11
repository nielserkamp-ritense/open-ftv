import { pipRequest as request } from '../api/client/pip.ts';
import { components } from '../oas/attributes';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

// Types for the attributes service
export type Attribute = components['schemas']['Attribute'];
export type AttributesResponse = components['schemas']['Attributes'];
export type AttributeResponse = components['schemas']['Attribute'];

const ATTRIBUTES_QUERY_KEYS = {
  all: ['attributes'] as const,
  lists: () => [...ATTRIBUTES_QUERY_KEYS.all, 'list'] as const,
  list: (filters: Record<string, unknown>) => [...ATTRIBUTES_QUERY_KEYS.lists(), filters] as const,
  details: () => [...ATTRIBUTES_QUERY_KEYS.all, 'detail'] as const,
  detail: (key: string) => [...ATTRIBUTES_QUERY_KEYS.details(), key] as const,
};

/**
 * Service for interacting with the attributes API endpoints
 */
export const attributesService = {
  /**
   * Get all attributes
   * @returns Promise with all attributes
   */
  getAllAttributes: async (): Promise<AttributesResponse> => {
    return request<AttributesResponse>({
      method: 'GET',
      url: '/v1/attributes',
    });
  },

  /**
   * Get a specific attribute
   * @param key - Unique key of the attribute
   * @returns Promise with the requested attribute
   */
  getAttribute: async (key: string): Promise<AttributeResponse> => {
    return request<AttributeResponse>({
      method: 'GET',
      url: `/v1/attribute/${encodeURIComponent(key)}`,
    });
  },

  /**
   * Add a new attribute
   * @param key - Unique key of the attribute
   * @param attribute - Attribute data to add
   * @param force - Force upsert if attribute already exists
   * @returns Promise with the added attribute
   */
  addAttribute: async (
    key: string,
    attribute: Attribute,
    force?: boolean
  ): Promise<AttributeResponse> => {
    return request<AttributeResponse>({
      method: 'POST',
      url: `/v1/attribute/${encodeURIComponent(key)}`,
      params: force ? { force } : undefined,
      data: attribute,
    });
  },

  /**
   * Replace an existing attribute
   * @param key - Unique key of the attribute
   * @param attribute - New attribute data
   * @param force - Force upsert if attribute doesn't exist
   * @returns Promise with the updated attribute
   */
  replaceAttribute: async (
    key: string,
    attribute: Attribute,
    force?: boolean
  ): Promise<AttributeResponse> => {
    return request<AttributeResponse>({
      method: 'PUT',
      url: `/v1/attribute/${encodeURIComponent(key)}`,
      params: force ? { force } : undefined,
      data: attribute,
    });
  },

  /**
   * Delete an attribute
   * @param key - Unique key of the attribute
   * @param force - Ignore missing data during delete
   * @returns Promise with the deleted attribute
   */
  deleteAttribute: async (
    key: string,
    force?: boolean
  ): Promise<AttributeResponse> => {
    return request<AttributeResponse>({
      method: 'DELETE',
      url: `/v1/attribute/${encodeURIComponent(key)}`,
      params: force ? { force } : undefined,
    });
  },
};

/**
 * Hook to fetch all attributes
 */
export const useAttributes = () => {
  return useQuery({
    queryKey: ATTRIBUTES_QUERY_KEYS.lists(),
    queryFn: attributesService.getAllAttributes,
  });
};

/**
 * Hook to fetch a specific attribute
 */
export const useAttribute = (key: string) => {
  return useQuery({
    queryKey: ATTRIBUTES_QUERY_KEYS.detail(key),
    queryFn: () => attributesService.getAttribute(key),
    enabled: !!key,
  });
};

/**
 * Hook to add a new attribute
 */
export const useAddAttribute = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ key, attribute, force }: {
      key: string;
      attribute: Attribute;
      force?: boolean;
    }) => attributesService.addAttribute(key, attribute, force),
    onSuccess: async (_data, variables) => {
      await queryClient.invalidateQueries({ queryKey: ATTRIBUTES_QUERY_KEYS.lists() });
      await queryClient.invalidateQueries({ queryKey: ATTRIBUTES_QUERY_KEYS.detail(variables.key) });
    },
  });
};

/**
 * Hook to replace an existing attribute
 */
export const useReplaceAttribute = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ key, attribute, force }: {
      key: string;
      attribute: Attribute;
      force?: boolean;
    }) => attributesService.replaceAttribute(key, attribute, force),
    onSuccess: async (_data, variables) => {
      await queryClient.invalidateQueries({ queryKey: ATTRIBUTES_QUERY_KEYS.lists() });
      await queryClient.invalidateQueries({ queryKey: ATTRIBUTES_QUERY_KEYS.detail(variables.key) });
    },
  });
};

/**
 * Hook to delete an attribute
 */
export const useDeleteAttribute = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ key, force }: {
      key: string;
      force?: boolean;
    }) => attributesService.deleteAttribute(key, force),
    onSuccess: async (_data, variables) => {
      await queryClient.invalidateQueries({ queryKey: ATTRIBUTES_QUERY_KEYS.lists() });
      await queryClient.invalidateQueries({ queryKey: ATTRIBUTES_QUERY_KEYS.detail(variables.key) });
    },
  });
};
