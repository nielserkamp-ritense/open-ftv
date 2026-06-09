import { describe, it, expect } from 'vitest';
import { rolesFromToken, canWrite, canPublish, isReadOnly } from './capabilities';

// Build a fake JWT: header.payload.signature with a base64url-encoded payload.
const jwt = (payload: object): string =>
  'x.' + Buffer.from(JSON.stringify(payload)).toString('base64url') + '.y';

describe('rolesFromToken', () => {
  it('returns [] for undefined / malformed tokens', () => {
    expect(rolesFromToken(undefined)).toEqual([]);
    expect(rolesFromToken('not-a-jwt')).toEqual([]);
  });
  it('reads the top-level roles claim', () => {
    expect(rolesFromToken(jwt({ roles: ['author'] }))).toEqual(['author']);
  });
  it('returns [] when there is no roles claim', () => {
    expect(rolesFromToken(jwt({ sub: 'x' }))).toEqual([]);
  });
});

describe('capabilities', () => {
  it('admin can write and publish', () => {
    expect(canWrite(['admin'])).toBe(true);
    expect(canPublish(['admin'])).toBe(true);
    expect(isReadOnly(['admin'])).toBe(false);
  });
  it('author can write but not publish', () => {
    expect(canWrite(['author'])).toBe(true);
    expect(canPublish(['author'])).toBe(false);
  });
  it('auditor is read-only', () => {
    expect(canWrite(['auditor'])).toBe(false);
    expect(canPublish(['auditor'])).toBe(false);
    expect(isReadOnly(['auditor'])).toBe(true);
  });
});
