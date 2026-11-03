'use strict';

const assert = require('node:assert/strict');
const {test} = require('node:test');
const {setLogger} = require('release-please');
const {resolveCandidate, verifyCandidate, candidateMetadata, githubContext} = require('./rc-version.cjs');

// Only the GitHub boundary is replaced. Version selection executes the pinned
// Manifest implementation and its conventional-commit parser and Go strategy.
setLogger({error() {}, warn() {}, info() {}, debug() {}, trace() {}});
const HEAD = 'a'.repeat(40);
const STABLE = 'b'.repeat(40);
const NEW_HEAD = 'c'.repeat(40);

function fixture({message = 'feat: add a capability', config, manifest, heads, published = ['2.0.0']} = {}) {
  const reads = [];
  const headSequence = heads ? [...heads] : [HEAD];
  const rootConfig = config || {
    packages: {'.': {
      'release-type': 'go',
      'package-name': 'agentic-lens',
      'include-component-in-tag': false,
      'extra-files': ['web/package.json'],
    }},
  };
  return {
    reads,
    readHead: async () => headSequence.length > 1 ? headSequence.shift() : headSequence[0],
    stableVersions: async () => published,
    github: {
      repository: {owner: 'fixture', repo: 'repository'},
      async getFileJson(name, branch) {
        reads.push([name, branch]);
        if (name === '.release-please-config.json') return rootConfig;
        if (name === '.release-please-manifest.json') return manifest || {'.': '2.0.0'};
        throw new Error(`Unexpected file read: ${name}`);
      },
      async *releaseIterator() {
        // An RC of the released version must not become the release baseline.
        yield {tagName: 'v2.0.0-rc.113+1234567', name: 'prior RC', sha: NEW_HEAD};
        yield {tagName: 'v2.0.0', name: 'stable', sha: STABLE};
      },
      async *mergeCommitIterator(branch) {
        assert.equal(branch, 'main');
        yield {sha: HEAD, message, files: ['backend/example.go']};
        yield {sha: STABLE, message: 'chore(main): release 2.0.0', files: []};
      },
    },
  };
}

test('resolver follows the real release engine for feature, patch, breaking, and explicit releases', async t => {
  const cases = [
    ['feature', 'feat: add a capability', '2.1.0'],
    ['patch', 'fix: correct a defect', '2.0.1'],
    ['breaking', 'feat!: remove the old API\n\nBREAKING CHANGE: clients must migrate', '3.0.0'],
    ['commit release-as', 'fix: prepare release\n\nRelease-As: 2.3.0', '2.3.0'],
  ];
  for (const [name, message, version] of cases) {
    await t.test(name, async () => {
      const context = fixture({message});
      assert.deepEqual(await resolveCandidate(context, HEAD), {status: 'ready', version});
      assert.deepEqual(new Set(context.reads.map(([name]) => name)), new Set([
        '.release-please-config.json', '.release-please-manifest.json',
      ]));
      assert.ok(context.reads.every(([, branch]) => branch === 'main'));
    });
  }
  await t.test('configured release-as', async () => {
    const context = fixture();
    const readFile = context.github.getFileJson;
    context.github.getFileJson = async (name, branch) => {
      const value = await readFile(name, branch);
      if (name === '.release-please-config.json') value.packages['.']['release-as'] = '4.0.0';
      return value;
    };
    assert.deepEqual(await resolveCandidate(context, HEAD), {status: 'ready', version: '4.0.0'});
  });
});

test('no release-worthy commits skip distinctly from GitHub errors', async () => {
  const result = await resolveCandidate(fixture({message: 'chore: maintain contributor guidance'}), HEAD);
  assert.equal(result.status, 'skip');
  assert.match(result.reason, /no release-worthy/);
  const failed = fixture();
  failed.github.getFileJson = async () => { throw new Error('fixture API unavailable'); };
  await assert.rejects(resolveCandidate(failed, HEAD), /fixture API unavailable/);
});

test('stale candidate commits skip before reading configuration or after branch calculation', async () => {
  const stale = fixture({heads: [NEW_HEAD]});
  assert.equal((await resolveCandidate(stale, HEAD)).status, 'skip');
  assert.equal(stale.reads.length, 0);
  for (const heads of [[HEAD, NEW_HEAD], [HEAD, HEAD, NEW_HEAD]]) {
    const result = await resolveCandidate(fixture({heads}), HEAD);
    assert.equal(result.status, 'skip');
    assert.match(result.reason, /advanced|superseded/);
  }
});

test('publication recheck rejects targets already stable or overtaken by a stable release', async () => {
  for (const published of [['2.1.0'], ['1.9.0', '2.2.0']]) {
    await assert.rejects(verifyCandidate(fixture({published}), HEAD, '2.1.0'), /must exceed published stable/);
  }
  assert.equal((await verifyCandidate(fixture({heads: [NEW_HEAD]}), HEAD, '2.1.0')).status, 'skip');
  assert.deepEqual(await verifyCandidate(fixture(), HEAD, '2.1.0'), {status: 'ready', version: '2.1.0'});
  await assert.rejects(resolveCandidate(fixture({published: ['2.1.0']}), HEAD), /must exceed published stable/);
});

test('invalid root configuration and missing released version cannot silently produce metadata', async () => {
  for (const config of [{packages: {}}, {packages: {'backend': {'release-type': 'go'}}}]) {
    await assert.rejects(resolveCandidate(fixture({config}), HEAD), /root release package/);
  }
  await assert.rejects(resolveCandidate(fixture({manifest: {}}), HEAD), /Root released version is missing/);
  await assert.rejects(resolveCandidate(fixture({manifest: {'.': '2.0.0-rc.1'}}), HEAD), /stable semantic version/);
  await assert.rejects(resolveCandidate(fixture({manifest: {'.': 'prefix2.0.0suffix'}}), HEAD), /stable semantic version/);
  await assert.rejects(resolveCandidate(fixture(), 'not-a-sha'), /full candidate commit SHA/);
  const configuredOldTarget = fixture();
  const readFile = configuredOldTarget.github.getFileJson;
  configuredOldTarget.github.getFileJson = async (name, branch) => {
    const value = await readFile(name, branch);
    if (name === '.release-please-config.json') value.packages['.']['release-as'] = '1.9.0';
    return value;
  };
  await assert.rejects(resolveCandidate(configuredOldTarget, HEAD), /must exceed the last released/);
});

test('metadata preserves the candidate version, run, commit, and build date without unsafe output values', () => {
  assert.deepEqual(candidateMetadata('2.1.0', '114', 'abc1234', '2026-11-03T12:00:00Z'), {
    VERSION: 'v2.1.0-rc.114+abc1234', GIT_COMMIT: 'abc1234',
    BUILD_DATE: '2026-11-03T12:00:00Z', RC_TARGET: '2.1.0',
  });
  for (const version of ['v2.1.0', '2.1.0-rc.1', '2.1.0+abc', '2.1.0\nINJECTED=true']) {
    assert.throws(() => candidateMetadata(version, '114', 'abc1234', '2026-11-03T12:00:00Z'), /stable semantic version/);
  }
  for (const run of ['0', '-1', '114\nINJECTED=true']) {
    assert.throws(() => candidateMetadata('2.1.0', run, 'abc1234', '2026-11-03T12:00:00Z'), /run number/);
  }
});

test('GitHub adapter paginates stable releases and excludes draft, prerelease, and legacy namespaces', async t => {
  const calls = [];
  t.mock.method(globalThis, 'fetch', async url => {
    calls.push(url);
    if (url.endsWith('git/ref/heads/main')) return Response.json({object: {sha: HEAD}});
    const page = new URL(url).searchParams.get('page');
    if (page === '1') return Response.json(Array.from({length: 100}, (_, index) => ({
      tag_name: `v2.0.0-rc.${index}`, prerelease: true, draft: false,
    })));
    if (page === '2') return Response.json([
      {tag_name: 'v2.0.0', prerelease: false, draft: false},
      {tag_name: 'v3.0.0', prerelease: false, draft: true},
      {tag_name: 'agentic-lens-v1.0.0', prerelease: false, draft: false},
      {tag_name: 'latest', prerelease: false, draft: false},
    ]);
    assert.fail(`Unexpected GitHub request: ${url}`);
  });
  const context = await githubContext('fixture/repository', 'fixture-token', 'https://fixture.invalid');
  assert.equal(await context.readHead(), HEAD);
  assert.deepEqual(await context.stableVersions(), ['2.0.0']);
  assert.equal(calls.length, 3);
});

test('GitHub transport errors and malformed responses fail closed', async t => {
  const context = await githubContext('fixture/repository', 'fixture-token', 'https://fixture.invalid');
  const request = t.mock.method(globalThis, 'fetch', async () => new Response('', {status: 503}));
  await assert.rejects(context.stableVersions(), /HTTP 503/);
  await assert.rejects(context.readHead(), /HTTP 503/);
  request.mock.mockImplementation(async () => Response.json({unexpected: true}));
  await assert.rejects(context.stableVersions(), /Invalid GitHub releases response/);
  await assert.rejects(context.readHead(), /Invalid main ref response/);
});
