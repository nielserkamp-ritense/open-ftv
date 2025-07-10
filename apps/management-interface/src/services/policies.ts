import { request } from '../api/client/base';
import { components } from '../oas/policies';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

// Types for the policies service
type Policy = components['schemas']['Policy'];
type PoliciesResponse = components['schemas']['PoliciesResponse'];
type PolicyResponse = components['schemas']['PolicyResponse'];

const POLICIES_QUERY_KEYS = {
  all: ['policies'] as const,
  lists: () => [...POLICIES_QUERY_KEYS.all, 'list'] as const,
  list: (filters: Record<string, unknown>) => [...POLICIES_QUERY_KEYS.lists(), filters] as const,
  details: () => [...POLICIES_QUERY_KEYS.all, 'detail'] as const,
  detail: (language: string, id: string) => [...POLICIES_QUERY_KEYS.details(), language, id] as const,
};


/**
 * Service for interacting with the policies API endpoints
 */
export const policiesService = {
  /**
   * Get all policies
   * @returns Promise with all policies
   */
  getAllPolicies: async (): Promise<PoliciesResponse> => {
    return request<PoliciesResponse>({
      method: 'GET',
      url: '/v1/policies',
    });
  },

  /**
   * Get a specific policy
   * @param language - Language of the policy
   * @param id - Unique identifier of the policy
   * @returns Promise with the requested policy
   */
  getPolicy: async (language: string, id: string): Promise<PolicyResponse> => {
    return request<PolicyResponse>({
      method: 'GET',
      url: `/v1/policy/${language}/${id}`,
    });
  },

  /**
   * Add a new policy
   * @param language - Language of the policy
   * @param id - Unique identifier of the policy
   * @param policy - Policy data to add
   * @param force - Force upsert if policy already exists
   * @returns Promise with the added policy
   */
  addPolicy: async (
    language: string,
    id: string,
    policy: Policy,
    force?: boolean
  ): Promise<PolicyResponse> => {
    return request<PolicyResponse>({
      method: 'POST',
      url: `/v1/policy/${language}/${id}`,
      params: force ? { force } : undefined,
      data: policy,
    });
  },

  /**
   * Replace an existing policy
   * @param language - Language of the policy
   * @param id - Unique identifier of the policy
   * @param policy - New policy data
   * @param force - Force upsert if policy doesn't exist
   * @returns Promise with the updated policy
   */
  replacePolicy: async (
    language: string,
    id: string,
    policy: Policy,
    force?: boolean
  ): Promise<PolicyResponse> => {
    return request<PolicyResponse>({
      method: 'PUT',
      url: `/v1/policy/${language}/${id}`,
      params: force ? { force } : undefined,
      data: policy,
    });
  },

  /**
   * Delete a policy
   * @param language - Language of the policy
   * @param id - Unique identifier of the policy
   * @param force - Ignore missing data during delete
   * @returns Promise with the deleted policy
   */
  deletePolicy: async (
    language: string,
    id: string,
    force?: boolean
  ): Promise<PolicyResponse> => {
    return request<PolicyResponse>({
      method: 'DELETE',
      url: `/v1/policy/${language}/${id}`,
      params: force ? { force } : undefined,
    });
  },
};

/**
 * Hook to fetch all policies
 */
export const usePolicies = () => {
  return useQuery({
    queryKey: POLICIES_QUERY_KEYS.lists(),
    queryFn: policiesService.getAllPolicies,
  });
};

/**
 * Hook to fetch a specific policy
 */
export const usePolicy = (language: string, id: string) => {
  return useQuery({
    queryKey: POLICIES_QUERY_KEYS.detail(language, id),
    queryFn: () => policiesService.getPolicy(language, id),
    enabled: !!language && !!id, // Only run the query if both parameters are provided
  });
};

/**
 * Hook to add a new policy
 */
export const useAddPolicy = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ language, id, policy, force }: {
      language: string;
      id: string;
      policy: Policy;
      force?: boolean
    }) => policiesService.addPolicy(language, id, policy, force),
    onSuccess: async () => {
      // Invalidate the policies list query to refetch the updated data
      await queryClient.invalidateQueries({ queryKey: POLICIES_QUERY_KEYS.lists() });
    },
  });
};

/**
 * Hook to replace an existing policy
 */
export const useReplacePolicy = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ language, id, policy, force }: {
      language: string;
      id: string;
      policy: Policy;
      force?: boolean
    }) => policiesService.replacePolicy(language, id, policy, force),
    onSuccess: async (_data, variables) => {
      // Invalidate both the list and the specific policy query
      await queryClient.invalidateQueries({ queryKey: POLICIES_QUERY_KEYS.lists() });
      await queryClient.invalidateQueries({
        queryKey: POLICIES_QUERY_KEYS.detail(variables.language, variables.id)
      });
    },
  });
};

/**
 * Hook to delete a policy
 */
export const useDeletePolicy = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ language, id, force }: {
      language: string;
      id: string;
      force?: boolean
    }) => policiesService.deletePolicy(language, id, force),
    onSuccess: async (_data, variables) => {
      // Invalidate both the list and the specific policy query
      await queryClient.invalidateQueries({ queryKey: POLICIES_QUERY_KEYS.lists() });
      await queryClient.invalidateQueries({
        queryKey: POLICIES_QUERY_KEYS.detail(variables.language, variables.id)
      });
    },
  });
};
