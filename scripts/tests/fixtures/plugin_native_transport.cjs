// Exercise production plugin command construction without claiming a real
// OpenCode lifecycle exchange.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const ts = require(process.argv[3]);
const template = fs.readFileSync(process.argv[2], 'utf8').split('const base =', 1)[0]
  .split('\n').filter(line => !line.includes('@opencode-ai/plugin')).join('\n');

async function exercise(platform, operation) {
  const manager = "/native agent's/agentic-lens";
  const project = "/project owner's/repo";
  const state = "/state owner's/data";
  const store = "/source owner's/home";
  let source = template;
  for (const [marker, value] of Object.entries({ __WRAPPER__: manager, __PROJECT__: project, __STATE__: state, __STORE__: store })) {
    source = source.replace(marker, JSON.stringify(value));
  }
  const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText;
  const records = [];
  const processes = {
    execFile(executable, args, options, callback) {
      const record = { executable, args: Array.from(args), options, payload: undefined };
      records.push(record);
      queueMicrotask(() => callback(null, 'native response ü'));
      return { stdin: { end(payload) { record.payload = payload; } } };
    }
  };
  const context = vm.createContext({
    require(name) { assert.equal(name, 'node:child_process'); return processes; },
    exports: {}, process: { platform }, Buffer, console
  });
  vm.runInContext(code, context);
  const args = [operation, '--source', 'opencode', '--store', store];
  if (operation === 'bind') args.push('--ticket-title', "User's literal $(value)", '--request', 'exact-request');
  const payload = operation === 'receive' ? '{"session_id":"root","text":"café"}\n' : undefined;
  context.fixtureArgs = args;
  context.fixturePayload = payload;
  const promise = vm.runInContext('call(fixtureArgs, fixturePayload)', context);
  assert.equal(records[0].executable, manager, 'plugin must invoke the native manager directly');
  assert.deepEqual(records[0].args, ['--al-offline', '--al-project', project, 'capture', ...args,
    '--state-dir', state, '--repo', project, '--wrapper', manager]);
  assert.equal(records[0].payload, payload);
  assert.equal(records[0].args.includes('-File'), false);
  assert.equal(records[0].args.includes('ExecutionPolicy'), false);
  assert.equal(records[0].options.timeout, 1500);
  assert.equal(records[0].options.maxBuffer, 65536);
  assert.equal(await promise, 'native response ü');
}

(async () => {
  for (const platform of ['linux', 'darwin']) {
    for (const operation of ['receive', 'request', 'bind']) await exercise(platform, operation);
  }
  console.log('Native plugin argv/payload controls and bounded transport passed; POSIX VM construction only.');
})().catch(error => { console.error(error); process.exitCode = 1; });
