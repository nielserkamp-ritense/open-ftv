// Who performed an action, as the API now reports it. The manager stores an id on every object
// and fills in the name from its local principal record; the name is absent for a principal it
// has never seen, and then the raw id is all there is to show. See docs/adr/0004.
export interface PrincipalRef {
  id: string;
  name?: string;
  kind?: 'user' | 'system' | 'legacy';
}

// principalName renders a principal, falling back to the raw id when no name is known.
// Written without `||` so an empty name falls back rather than rendering as blank: `??` would
// not, since '' is not nullish.
export const principalName = (p?: PrincipalRef): string => {
  const name = p?.name?.trim();
  if (name !== undefined && name !== '') {
    return name;
  }

  return p?.id?.trim() ?? '';
};

// isLinkable reports whether a principal identifies a real subject that can be linked to. A
// `legacy` principal is keyed by a display name and identifies nobody; `system` is the manager.
export const isLinkable = (p?: PrincipalRef): boolean => p?.kind === 'user';
