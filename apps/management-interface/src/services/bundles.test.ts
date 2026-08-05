import { describe, it, expect } from 'vitest';
import { extractPdpUrls, BundleConfig } from './bundles';

const configWithTargets = (id: string, uris: string[]): BundleConfig => ({
  id,
  title: id,
  tags: [],
  targets: uris.map((uri) => ({ uri })),
});

describe('extractPdpUrls', () => {
  it('returns [] when there are no bundle configurations', () => {
    expect(extractPdpUrls(undefined)).toEqual([]);
    expect(extractPdpUrls([])).toEqual([]);
  });

  it('returns [] when a configuration has no targets', () => {
    expect(extractPdpUrls([configWithTargets('pdp1', [])])).toEqual([]);
  });

  it('returns the single target URI for one config with one target', () => {
    expect(extractPdpUrls([configWithTargets('pdp1', ['http://vlierdam-pdp1:9443'])])).toEqual([
      'http://vlierdam-pdp1:9443',
    ]);
  });

  it('collects targets from multiple configs', () => {
    const configs = [
      configWithTargets('pdp1', ['http://rdw-pdp1:9443']),
      configWithTargets('openftv-manager', ['http://rvig-manager:9443']),
    ];
    expect(extractPdpUrls(configs)).toEqual(['http://rdw-pdp1:9443', 'http://rvig-manager:9443']);
  });

  it('collects multiple targets from a single config', () => {
    const configs = [configWithTargets('pdp1', ['http://pdp-a:9443', 'http://pdp-b:9443'])];
    expect(extractPdpUrls(configs)).toEqual(['http://pdp-a:9443', 'http://pdp-b:9443']);
  });

  it('deduplicates a URI that appears in more than one config', () => {
    const configs = [
      configWithTargets('pdp1', ['http://shared-pdp:9443']),
      configWithTargets('pdp2', ['http://shared-pdp:9443']),
    ];
    expect(extractPdpUrls(configs)).toEqual(['http://shared-pdp:9443']);
  });
});
