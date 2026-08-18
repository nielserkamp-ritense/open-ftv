import { describe, expect, it } from 'vitest';
import { isLinkable, principalName } from './principal';

describe('principalName', () => {
  it('prefers the resolved name', () => {
    expect(principalName({ id: 'sub-123', name: 'Ton de Vries', kind: 'user' })).toBe('Ton de Vries');
  });

  // A principal the manager has never seen carries no name, so the id is all we can show.
  it('falls back to the raw id', () => {
    expect(principalName({ id: 'sub-123', kind: 'user' })).toBe('sub-123');
    expect(principalName({ id: 'sub-123', name: '   ' })).toBe('sub-123');
  });

  it('renders nothing when there is no principal at all', () => {
    expect(principalName(undefined)).toBe('');
    expect(principalName({ id: '' })).toBe('');
  });
});

describe('isLinkable', () => {
  it('only links to real subjects', () => {
    expect(isLinkable({ id: 'sub-123', kind: 'user' })).toBe(true);
    expect(isLinkable({ id: 'Ton de Vries', kind: 'legacy' })).toBe(false);
    expect(isLinkable({ id: '*SEED*', kind: 'system' })).toBe(false);
    expect(isLinkable(undefined)).toBe(false);
  });
});
