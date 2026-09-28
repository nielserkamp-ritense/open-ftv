// Query Keys Factory for cache management

export const queryKeys = {
  aanvragen: {
    all: ['aanvragen'] as const,
    lists: () => [...queryKeys.aanvragen.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.aanvragen.lists(), filters] as const,
    details: () => [...queryKeys.aanvragen.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.aanvragen.details(), id] as const,
  },
  laadpalen: {
    all: ['laadpalen'] as const,
    lists: () => [...queryKeys.laadpalen.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.laadpalen.lists(), filters] as const,
    details: () => [...queryKeys.laadpalen.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.laadpalen.details(), id] as const,
  },
  brp: {
    all: ['brp'] as const,
    lists: () => [...queryKeys.brp.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.brp.lists(), filters] as const,
    details: () => [...queryKeys.brp.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.brp.details(), id] as const,
  },
  rdw: {
    all: ['rdw'] as const,
    lists: () => [...queryKeys.rdw.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.rdw.lists(), filters] as const,
    details: () => [...queryKeys.rdw.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.rdw.details(), id] as const,
  },
  gebruikers: {
    all: ['gebruikers'] as const,
    lists: () => [...queryKeys.gebruikers.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.gebruikers.lists(), filters] as const,
    details: () => [...queryKeys.gebruikers.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.gebruikers.details(), id] as const,
  },
}
