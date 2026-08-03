import { describe, it, expect } from 'vitest';
import { displayName, initials, UNKNOWN_USER } from './profile';

describe('displayName', () => {
  it('prefers name over preferred_username and email', () => {
    expect(displayName({ name: 'Admin User', preferred_username: 'admin-user', email: 'a@b.local' })).toBe('Admin User');
  });
  it('falls back to preferred_username, then email', () => {
    expect(displayName({ preferred_username: 'admin-user', email: 'a@b.local' })).toBe('admin-user');
    expect(displayName({ email: 'a@b.local' })).toBe('a@b.local');
  });
  it('ignores blank claims', () => {
    expect(displayName({ name: '   ', preferred_username: 'admin-user' })).toBe('admin-user');
  });
  it('returns a placeholder for an absent or empty profile', () => {
    expect(displayName(undefined)).toBe(UNKNOWN_USER);
    expect(displayName({})).toBe(UNKNOWN_USER);
  });
});

describe('initials', () => {
  it('takes the first and last word of a full name', () => {
    expect(initials('Admin User')).toBe('AU');
    expect(initials('Jan Willem van den Berg')).toBe('JB');
  });
  it('takes the first two letters of a single word', () => {
    expect(initials('admin-user')).toBe('AD');
    expect(initials('x')).toBe('X');
  });
  it('handles diacritics', () => {
    expect(initials('Émile Étienne')).toBe('ÉÉ');
  });
  it('returns a placeholder for an empty name', () => {
    expect(initials('')).toBe('?');
    expect(initials('   ')).toBe('?');
  });
});
