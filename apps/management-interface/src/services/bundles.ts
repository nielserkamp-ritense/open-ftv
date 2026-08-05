import { request } from '../api/client/base';
import { components } from '../oas/bundles';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

// Types for the bundles service
export type Status = components['schemas']['Status'];
export type StatusesResponse = components['schemas']['Statuses'];
export type CompressType = components['schemas']['CompressType'];
export type CompressTypesResponse = components['schemas']['CompressTypes'];
export type BundleConfig = components['schemas']['BundleConfig'];
export type BundleConfigsResponse = components['schemas']['BundleConfigs'];
export type Deployment = components['schemas']['Deployment'];
export type DeploymentsResponse = components['schemas']['Deployments'];
export type BundleActivated = components['schemas']['BundleActivated'];
export type NewDeploymentBody = components['schemas']['NewDeploymentBody'];

export type DeploymentID = components['parameters']['DeploymentID'];
export type BundleID = components['parameters']['BundleID'];

/**
 * Extracts the unique set of PDP target URIs across all bundle configurations.
 */
export function extractPdpUrls(configs: BundleConfigsResponse | undefined): string[] {
  return Array.from(new Set((configs ?? []).flatMap((config) => config.targets.map((target) => target.uri))));
}

const BUNDLES_QUERY_KEYS = {
  all: ['bundles'] as const,
  statuses: () => [...BUNDLES_QUERY_KEYS.all, 'statuses'] as const,
  compressionTypes: () => [...BUNDLES_QUERY_KEYS.all, 'compression-types'] as const,
  configurations: () => [...BUNDLES_QUERY_KEYS.all, 'configurations'] as const,
  deployments: () => [...BUNDLES_QUERY_KEYS.all, 'deployments'] as const,
  deployment: (id: DeploymentID | string) => [...BUNDLES_QUERY_KEYS.all, 'deployment', id] as const,
  lastDeployment: () => [...BUNDLES_QUERY_KEYS.all, 'deployment', 'last'] as const,
  bundle: (id: BundleID | string) => [...BUNDLES_QUERY_KEYS.all, 'bundle', id] as const,
};

/**
 * Service for interacting with the bundles API endpoints
 */
export const bundlesService = {
  /**
   * Retrieve deployment status codes and names
   */
  getStatuses: async (): Promise<StatusesResponse> => {
    return request<StatusesResponse>({
      method: 'GET',
      url: '/v1/statuses',
    });
  },

  /**
   * Retrieve bundle compression types
   */
  getCompressionTypes: async (): Promise<CompressTypesResponse> => {
    return request<CompressTypesResponse>({
      method: 'GET',
      url: '/v1/compression-types',
    });
  },

  /**
   * Retrieve bundle configurations
   */
  getBundleConfigurations: async (): Promise<BundleConfigsResponse> => {
    return request<BundleConfigsResponse>({
      method: 'GET',
      url: '/v1/bundle-configurations',
    });
  },

  /**
   * Retrieve all deployments
   */
  getDeployments: async (): Promise<DeploymentsResponse> => {
    return request<DeploymentsResponse>({
      method: 'GET',
      url: '/v1/deployments',
    });
  },

  /**
   * Retrieve a specific deployment by version key
   */
  getDeployment: async (key: DeploymentID): Promise<Deployment> => {
    return request<Deployment>({
      method: 'GET',
      url: `/v1/deployment/${key}`,
    });
  },

  /**
   * Retrieve the last deployment
   */
  getLastDeployment: async (): Promise<Deployment> => {
    return request<Deployment>({
      method: 'GET',
      url: `/v1/deployment/last`,
    });
  },

  /**
   * Start a new deployment
   */
  startDeployment: async (body: NewDeploymentBody): Promise<Deployment> => {
    return request<Deployment>({
      method: 'POST',
      url: '/v1/deployment',
      data: body,
    });
  },

  /**
   * Retrieve the latest deployment bundle by id
   * Returns octet-stream as string (transported as text)
   */
  getBundle: async (id: BundleID): Promise<string> => {
    return request<string>({
      method: 'GET',
      url: `/v1/bundle/${id}`,
      // Expect binary as string; adjust if backend returns blob
      responseType: 'arraybuffer' as unknown as undefined,
    });
  },

  /**
   * Push a new deployment bundle to activate
   * Accepts binary content as string (octet-stream)
   */
  pushBundle: async (bundle: string): Promise<BundleActivated> => {
    return request<BundleActivated>({
      method: 'POST',
      url: `/v1/bundle`,
      data: bundle,
      headers: {
        'Content-Type': 'application/octet-stream',
      },
    });
  },
};

// React Query hooks
export const useStatuses = () =>
  useQuery({
    queryKey: BUNDLES_QUERY_KEYS.statuses(),
    queryFn: bundlesService.getStatuses,
  });

export const useCompressionTypes = () =>
  useQuery({
    queryKey: BUNDLES_QUERY_KEYS.compressionTypes(),
    queryFn: bundlesService.getCompressionTypes,
  });

export const useBundleConfigurations = () =>
  useQuery({
    queryKey: BUNDLES_QUERY_KEYS.configurations(),
    queryFn: bundlesService.getBundleConfigurations,
  });

export const useDeployments = () =>
  useQuery({
    queryKey: BUNDLES_QUERY_KEYS.deployments(),
    queryFn: bundlesService.getDeployments,
  });

export const useDeployment = (key?: DeploymentID) =>
  useQuery({
    queryKey: BUNDLES_QUERY_KEYS.deployment(key ?? 'undefined'),
    queryFn: () => (key == null ? Promise.resolve(undefined as unknown as Deployment) : bundlesService.getDeployment(key)),
    enabled: key != null,
  });

export const useLastDeployment = () =>
  useQuery({
    queryKey: BUNDLES_QUERY_KEYS.lastDeployment(),
    queryFn: bundlesService.getLastDeployment,
  });

export const useStartDeployment = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: NewDeploymentBody) => bundlesService.startDeployment(body),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: BUNDLES_QUERY_KEYS.deployments() });
      await queryClient.invalidateQueries({ queryKey: BUNDLES_QUERY_KEYS.lastDeployment() });
    },
  });
};

export const useBundle = (id?: BundleID) =>
  useQuery({
    queryKey: BUNDLES_QUERY_KEYS.bundle(id ?? 'undefined'),
    queryFn: () => (id == null ? Promise.resolve(undefined as unknown as string) : bundlesService.getBundle(id)),
    enabled: id != null,
  });

export const usePushBundle = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (bundle: string) => bundlesService.pushBundle(bundle),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: BUNDLES_QUERY_KEYS.deployments() });
      await queryClient.invalidateQueries({ queryKey: BUNDLES_QUERY_KEYS.lastDeployment() });
    },
  });
};
