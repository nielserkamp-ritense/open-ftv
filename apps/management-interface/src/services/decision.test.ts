import { describe, it, expect } from 'vitest';
import type { AuthlogEntry } from './authlog';
import { entryDecision, entryReason, evaluationFailed } from './decision';

const entry = (fields: { status?: string; request?: object; response?: object | null }) =>
  fields as unknown as AuthlogEntry;

const batch = (decisions: object[], semantic?: string) =>
  entry({
    status: 'Ok',
    request: semantic ? { options: { evaluations_semantic: semantic } } : {},
    response: { evaluations: decisions },
  });

describe('entryDecision', () => {
  it('reads a single evaluation', () => {
    expect(entryDecision(entry({ status: 'Ok', response: { decision: true } }))).toBe('permitted');
    expect(entryDecision(entry({ status: 'Ok', response: { decision: false } }))).toBe('denied');
  });

  it('marks a failed evaluation as an error', () => {
    expect(entryDecision(entry({ status: 'Error', response: null }))).toBe('error');
  });

  it('permits a batch only when every evaluation is permitted', () => {
    expect(entryDecision(batch([{ decision: true }, { decision: true }]))).toBe('permitted');
    expect(entryDecision(batch([{ decision: true }, { decision: false }]))).toBe('denied');
    expect(entryDecision(batch([{ decision: false }], 'deny_on_first_deny'))).toBe('denied');
  });

  it('permits a permit_on_first_permit batch on one permit', () => {
    expect(entryDecision(batch([{ decision: false }, { decision: true }], 'permit_on_first_permit'))).toBe('permitted');
    expect(entryDecision(batch([{ decision: false }], 'permit_on_first_permit'))).toBe('denied');
  });

  it('marks a batch that is not permitted because an evaluation failed as an error', () => {
    const error = { context: { error: { status: 500, message: 'datastore unavailable' } } };
    expect(entryDecision(batch([{ decision: true }, { decision: false, ...error }]))).toBe('error');
    expect(entryDecision(batch([{ decision: false }, { decision: false, ...error }]))).toBe('error');
    expect(entryDecision(batch([{ decision: false, ...error }, { decision: true }], 'permit_on_first_permit'))).toBe('permitted');
  });

  it('has no decision for an entry without one', () => {
    expect(entryDecision(undefined)).toBeUndefined();
    expect(entryDecision(batch([]))).toBeUndefined();
    expect(entryDecision(entry({ status: 'Ok', response: { results: [] } }))).toBeUndefined();
  });
});

describe('entryReason', () => {
  it('reads the reason of a single evaluation', () => {
    expect(entryReason(entry({ status: 'Ok', response: { decision: false, context: { reason: 'buiten kantoortijd' } } }))).toBe(
      'buiten kantoortijd',
    );
  });

  it('reads the reason of the first denied evaluation in a batch', () => {
    const reason = entryReason(
      batch([
        { decision: true, context: { reason: 'toegestaan' } },
        { decision: false, context: { reason: 'buiten kantoortijd' } },
        { decision: false, context: { reason: 'onbekend' } },
      ]),
    );
    expect(reason).toBe('buiten kantoortijd');
  });

  it('reads the error of a batch evaluation that failed, before any denial', () => {
    const reason = entryReason(
      batch([
        { decision: false, context: { reason: 'buiten kantoortijd' } },
        { decision: false, context: { error: { status: 500, message: 'datastore unavailable' } } },
      ]),
    );
    expect(reason).toBe('datastore unavailable');
  });

  it('says a failed evaluation failed', () => {
    expect(entryReason(entry({ status: 'Error', response: null }))).toBe(evaluationFailed);
  });

  it('skips the placeholders of decisions logged before AuthZEN 1.0', () => {
    expect(entryReason(entry({ status: 'Ok', response: { decision: false, context: { reason_user: { en: 'not authorized' } } } }))).toBeUndefined();
    expect(entryReason(entry({ status: 'Ok', response: { decision: false, context: { reason_user: { en: 'te laat' } } } }))).toBe('te laat');
  });
});
