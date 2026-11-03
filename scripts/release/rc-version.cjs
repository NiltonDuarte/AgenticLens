'use strict';

const {appendFileSync} = require('node:fs');
const {execFileSync} = require('node:child_process');
const {Manifest, GitHub, setLogger} = require('release-please');
const {Version} = require('release-please/build/src/version');

// Use plain logs instead of the library's decorative checkpoint logger.
setLogger({error: console.error, warn: console.warn, info: console.log, debug() {}, trace() {}});

function stableVersion(value) {
  if (typeof value !== 'string' || !/^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(value)) {
    throw new Error(`Expected a stable semantic version, received ${JSON.stringify(value)}`);
  }
  return Version.parse(value);
}

async function verifyCandidate({readHead, stableVersions}, expectedSha, targetVersion) {
  if (!/^[a-f0-9]{40}$/.test(expectedSha)) throw new Error('Expected the full candidate commit SHA');
  const target = stableVersion(targetVersion);
  if (await readHead() !== expectedSha) {
    return {status: 'skip', reason: 'Candidate commit has been superseded on main'};
  }
  for (const version of await stableVersions()) {
    if (target.compare(stableVersion(version)) <= 0) {
      throw new Error(`Candidate target ${targetVersion} must exceed published stable ${version}`);
    }
  }
  return {status: 'ready', version: targetVersion};
}

async function resolveCandidate(context, expectedSha) {
  if (!/^[a-f0-9]{40}$/.test(expectedSha)) throw new Error('Expected the full candidate commit SHA');
  if (await context.readHead() !== expectedSha) {
    return {status: 'skip', reason: 'Candidate commit has been superseded on main'};
  }
  const [config, versions] = await Promise.all([
    context.github.getFileJson('.release-please-config.json', 'main'),
    context.github.getFileJson('.release-please-manifest.json', 'main'),
  ]);
  if (!config || !config.packages || Object.keys(config.packages).length !== 1 ||
      !config.packages['.'] || typeof config.packages['.'] !== 'object') {
    throw new Error('RC resolution requires exactly one root release package');
  }
  if (!versions || typeof versions !== 'object') throw new Error('Invalid release manifest');
  if (!Object.hasOwn(versions, '.')) throw new Error('Root released version is missing from the release manifest');
  stableVersion(versions['.']);
  // This read-only API builds the same proposal as release-please-action v4.
  // No pull request is created and no existing PR title/body is consulted.
  const manifest = await Manifest.fromManifest(
    context.github, 'main', '.release-please-config.json', '.release-please-manifest.json',
  );
  if (Object.keys(manifest.repositoryConfig).length !== 1 || !manifest.repositoryConfig['.']) {
    throw new Error('RC resolution requires exactly one root release package');
  }
  if (manifest.repositoryConfig['.'].includeComponentInTag !== false) {
    throw new Error('RC resolution requires unprefixed root release tags');
  }
  const released = manifest.releasedVersions['.'];
  if (!released) throw new Error('Root released version is missing from the release manifest');
  stableVersion(released.toString());
  const candidates = await manifest.buildPullRequests();
  if (await context.readHead() !== expectedSha) {
    return {status: 'skip', reason: 'Main advanced while calculating the candidate version'};
  }
  if (candidates.length === 0) {
    return {status: 'skip', reason: 'Release Please found no release-worthy changes'};
  }
  if (candidates.length !== 1 || !candidates[0].version) {
    throw new Error('Release Please returned an ambiguous or missing root candidate version');
  }
  const targetVersion = candidates[0].version.toString();
  if (stableVersion(targetVersion).compare(released) <= 0) {
    throw new Error('Candidate target must exceed the last released manifest version');
  }
  return verifyCandidate(context, expectedSha, targetVersion);
}

function candidateMetadata(version, runNumber, shortSha, buildDate) {
  stableVersion(version);
  if (!/^[1-9]\d*$/.test(String(runNumber))) throw new Error('A positive workflow run number is required');
  if (!/^[a-f0-9]{7,40}$/.test(shortSha)) throw new Error('A valid short commit SHA is required');
  if (!/^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ$/.test(buildDate)) throw new Error('A UTC build date is required');
  return {
    VERSION: `v${version}-rc.${runNumber}+${shortSha}`,
    GIT_COMMIT: shortSha,
    BUILD_DATE: buildDate,
    RC_TARGET: version,
  };
}

async function githubContext(repository, token, apiUrl) {
  if (!/^[\w.-]+\/[\w.-]+$/.test(repository || '')) throw new Error('GITHUB_REPOSITORY is required');
  if (!token) throw new Error('GH_TOKEN is required for read-only version resolution');
  const [owner, repo] = repository.split('/');
  async function get(path) {
    const response = await fetch(`${apiUrl}/repos/${repository}/${path}`, {
      headers: {Authorization: `Bearer ${token}`, Accept: 'application/vnd.github+json'},
    });
    if (!response.ok) throw new Error(`GitHub read ${path} failed: HTTP ${response.status}`);
    return response.json();
  }
  return {
    github: await GitHub.create({owner, repo, defaultBranch: 'main', token, apiUrl}),
    readHead: async () => {
      const ref = await get('git/ref/heads/main');
      if (!ref.object || !/^[a-f0-9]{40}$/.test(ref.object.sha)) throw new Error('Invalid main ref response');
      return ref.object.sha;
    },
    stableVersions: async () => {
      const versions = [];
      for (let page = 1; ; page++) {
        const releases = await get(`releases?per_page=100&page=${page}`);
        if (!Array.isArray(releases)) throw new Error('Invalid GitHub releases response');
        for (const release of releases) {
          if (!release || typeof release.tag_name !== 'string' ||
              typeof release.draft !== 'boolean' || typeof release.prerelease !== 'boolean') {
            throw new Error('Invalid GitHub release record');
          }
          if (release.draft || release.prerelease) continue;
          // The configured root package has unprefixed v tags. Other releases
          // and legacy component-prefixed tags are not its stable namespace.
          if (/^v\d+\.\d+\.\d+$/.test(release.tag_name)) versions.push(release.tag_name.slice(1));
        }
        if (releases.length < 100) return versions;
      }
    },
  };
}

async function main(mode, env = process.env) {
  if (!['resolve', 'verify'].includes(mode)) throw new Error('Usage: rc-version.cjs resolve|verify');
  if (!env.GITHUB_OUTPUT || !env.GITHUB_ENV) throw new Error('GitHub output and environment files are required');
  const expectedSha = env.EXPECTED_SHA;
  const checkedOut = execFileSync('git', ['rev-parse', 'HEAD'], {encoding: 'utf8'}).trim();
  if (checkedOut !== expectedSha) throw new Error('Checkout does not match the workflow candidate SHA');
  const context = await githubContext(env.GITHUB_REPOSITORY, env.GH_TOKEN, env.GITHUB_API_URL || 'https://api.github.com');
  const result = mode === 'resolve'
    ? await resolveCandidate(context, expectedSha)
    : await verifyCandidate(context, expectedSha, env.RC_TARGET);
  if (result.status === 'skip') {
    console.log(`Skipping release candidate: ${result.reason}`);
    appendFileSync(env.GITHUB_OUTPUT, 'should_publish=false\n');
    return;
  }
  if (mode === 'resolve') {
    const shortSha = execFileSync('git', ['rev-parse', '--short', 'HEAD'], {encoding: 'utf8'}).trim();
    const date = new Date().toISOString().replace(/\.\d{3}Z$/, 'Z');
    const metadata = candidateMetadata(result.version, env.GITHUB_RUN_NUMBER, shortSha, date);
    appendFileSync(env.GITHUB_ENV, Object.entries(metadata).map(([key, value]) => `${key}=${value}\n`).join(''));
    console.log(`Release candidate version: ${metadata.VERSION}`);
  }
  appendFileSync(env.GITHUB_OUTPUT, 'should_publish=true\n');
}

module.exports = {resolveCandidate, verifyCandidate, candidateMetadata, githubContext};
if (require.main === module) {
  main(process.argv[2]).catch(error => {
    console.error(`Release candidate resolution failed: ${error.message}`);
    process.exitCode = 1;
  });
}
