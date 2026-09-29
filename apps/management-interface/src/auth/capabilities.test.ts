import { describe, it, expect } from 'vitest';
import { rolesFromToken, canWrite, canPublish, isReadOnly } from './capabilities';

// Build a fake JWT: header.payload.signature with a base64url-encoded payload.
const base64url = (s: string): string =>
  btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
const jwt = (payload: object): string =>
  'x.' + base64url(JSON.stringify(payload)) + '.y';

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
  it('author writes and publishes', () => {
    expect(canWrite(['author'])).toBe(true);
    expect(canPublish(['author'])).toBe(true);
    expect(isReadOnly(['author'])).toBe(false);
  });
  it('admin must not write or publish rules', () => {
    expect(canWrite(['admin'])).toBe(false);
    expect(canPublish(['admin'])).toBe(false);
    expect(isReadOnly(['admin'])).toBe(true);
  });
  it('auditor is read-only', () => {
    expect(canWrite(['auditor'])).toBe(false);
    expect(canPublish(['auditor'])).toBe(false);
    expect(isReadOnly(['auditor'])).toBe(true);
  });
});
