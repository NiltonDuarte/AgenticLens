#!/usr/bin/env node
// Repository setup workflow: validate a ticket, then let OpenSpec create its change.
import { readFileSync, existsSync, realpathSync, lstatSync } from 'node:fs';
import { resolve, join, sep } from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';
import { spawnSync } from 'node:child_process';

function fail(message) {
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
const toolRoot = fileURLToPath(new URL('../tools/openspec/', import.meta.url));
const require = createRequire(join(toolRoot, 'package.json'));
let parseDocument;
try { ({ parseDocument } = require('yaml')); }
catch { fail('Install the pinned toolchain first: npm ci --prefix tools/openspec'); }
const args = process.argv.slice(2);
if (args.length < 2 || args.length > 3 || args.some(value => !value.trim())) {
  fail('Usage: node scripts/openspec-change.mjs CANONICAL_TICKET "description" [--planning]');
}
const [ticket, description, mode] = args;
if (mode !== undefined && mode !== '--planning') fail('Only --planning is supported as the third argument.');
const project = realpathSync(process.cwd());
const configs = ['.agenticlens-config.yaml', '.agenticlens-config.yml'].map(name => join(project, name)).filter(existsSync);
if (configs.length !== 1) fail('Run from a project with exactly one canonical shared AgenticLens YAML file.');
const document = parseDocument(readFileSync(configs[0], 'utf8'), { uniqueKeys: true });
if (document.errors.length) fail(`Invalid shared YAML: ${document.errors.map(error => error.message).join('; ')}`);
let policy;
try { policy = document.toJS(); }
catch (error) { fail(`Invalid shared YAML structure: ${error.message}`); }
const prefixes = policy?.ticket?.prefixes;
if (!Array.isArray(prefixes) || !prefixes.length || prefixes.some(entry => typeof entry?.prefix !== 'string' || !entry.prefix.trim() || typeof entry?.description !== 'string' || !entry.description.trim())) fail('Configure explicit ticket prefixes and descriptions before creating a change.');
const valid = prefixes.some(entry => typeof entry?.prefix === 'string' && typeof entry?.description === 'string' && entry.description.trim() && ticket.startsWith(entry.prefix) && /^\d+$/.test(ticket.slice(entry.prefix.length)));
if (!valid) fail('Ticket must use a configured literal prefix in canonical spelling followed by digits.');
const slug = description.normalize('NFKD').replace(/[\u0300-\u036f]/g, '').toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
const name = `${ticket.toLowerCase()}-${slug}`;
if (!slug || !/^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$/.test(name)) fail('Ticket and description must produce a valid lowercase OpenSpec change name.');
const openConfigs = ['config.yaml', 'config.yml'].map(name => join(project, 'openspec', name)).filter(existsSync);
if (openConfigs.length !== 1) fail('Set up exactly one reviewed project OpenSpec configuration before creating a change.');
const openDocument = parseDocument(readFileSync(openConfigs[0], 'utf8'), { uniqueKeys: true });
if (openDocument.errors.length) fail('Invalid project OpenSpec configuration; resolve it before creation.');
const openPolicy = openDocument.toJS();
if (!openPolicy || typeof openPolicy !== 'object' || Array.isArray(openPolicy)) fail('OpenSpec configuration must be a mapping.');
if (Object.hasOwn(openPolicy, 'store')) fail('Cross-repository OpenSpec stores are outside this project-local creation workflow.');
for (const directory of [join(project, 'openspec'), join(project, 'openspec', 'changes')]) {
  if (existsSync(directory) && !realpathSync(directory).startsWith(project + sep)) fail('OpenSpec change creation must stay in the selected repository; cross-repository stores are outside this workflow.');
}
const schema = mode === '--planning' ? 'agenticlens-planning' : 'agenticlens';
if (!existsSync(join(project, 'openspec', 'schemas', schema, 'schema.yaml'))) fail(`Missing approved project schema: ${schema}`);
const destination = resolve(project, 'openspec', 'changes', name);
try {
  lstatSync(destination);
  fail(`Change already exists; refusing to overwrite ${destination}`);
} catch (error) {
  if (error.code !== 'ENOENT') fail(`Cannot inspect change destination: ${error.message}`);
}
const cli = join(toolRoot, 'node_modules', '@fission-ai', 'openspec', 'bin', 'openspec.js');
if (!existsSync(cli)) fail('Install the pinned toolchain first: npm ci --prefix tools/openspec');
const result = spawnSync(process.execPath, [cli, 'new', 'change', name, '--description', description, '--schema', schema, '--json'], {
  cwd: project, env: { ...process.env, OPENSPEC_TELEMETRY: '0' }, encoding: 'utf8', stdio: ['inherit', 'pipe', 'pipe'],
});
if (result.error) fail(`OpenSpec creation failed: ${result.error.message}`);
if (result.stdout) process.stdout.write(result.stdout);
if (result.stderr) process.stderr.write(result.stderr);
if (result.status !== 0) fail('OpenSpec creation failed; inspect its diagnostics and preserve existing content.');
process.stdout.write(`Canonical ticket: ${ticket}\nChange directory: ${destination}\nNo telemetry ledger or workspace registration was created.\n`);
