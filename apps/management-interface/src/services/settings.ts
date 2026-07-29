import { request } from '../api/client/base';
import { components } from '../oas/settings';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

// Types for the settings service
export type Settings = components['schemas']['Settings'];

export const DEFAULT_HEADER_TITLE = 'OpenFTV beheeromgeving';
export const DEFAULT_HEADER_COLOR = '#F7E8E8';
export const DEFAULT_TITLE_COLOR = '#000000';

const SETTINGS_QUERY_KEYS = {
  all: ['settings'] as const,
};

/**
 * Service for interacting with the settings API endpoints
 */
export const settingsService = {
  /**
   * Get the manager ui settings
   */
  getSettings: async (): Promise<Settings> => {
    return request<Settings>({
      method: 'GET',
      url: '/v1/settings',
    });
  },

  /**
   * Replace the manager ui settings
   */
  updateSettings: async (settings: Settings): Promise<Settings> => {
    return request<Settings>({
      method: 'PUT',
      url: '/v1/settings',
      data: settings,
    });
  },
};

/**
 * Hook to fetch the manager ui settings
 */
export const useSettings = () => {
  return useQuery({
    queryKey: SETTINGS_QUERY_KEYS.all,
    queryFn: settingsService.getSettings,
  });
};

/**
 * Hook to replace the manager ui settings
 */
export const useUpdateSettings = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: settingsService.updateSettings,
    onSuccess: (data) => {
      queryClient.setQueryData(SETTINGS_QUERY_KEYS.all, data);
    },
  });
};
