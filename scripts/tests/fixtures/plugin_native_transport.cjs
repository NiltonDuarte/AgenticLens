// Exercise production plugin command construction on Linux without claiming
// execution of Windows binaries or a real OpenCode lifecycle exchange.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const { EventEmitter } = require('node:events');
const ts = require(process.argv[3]);
const template = fs.readFileSync(process.argv[2], 'utf8').split('const base =', 1)[0]
  .split('\n').filter(line => !line.includes('@opencode-ai/plugin')).join('\n');

async function exercise(platform, operation, timeout = false) {
  const windows = platform === 'win32';
  const manager = windows ? "C:\\Program Files\\Agent's Lens\\agentic-lens.exe" : "/native agent's/agentic-lens";
  const project = windows ? "C:\\Projects\\Owner's repo" : "/project owner's/repo";
  const state = windows ? "C:\\State\\owner's data" : "/state owner's/data";
  const store = windows ? "C:\\Source\\owner's home" : "/source owner's/home";
  let source = template;
  for (const [marker, value] of Object.entries({ __WRAPPER__: manager, __PROJECT__: project, __STATE__: state, __STORE__: store })) {
    source = source.replace(marker, JSON.stringify(value));
  }
  const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS } }).outputText;
  const records = [];
  const timers = new Map();
  let sequence = 0;
  const timer = (callback, milliseconds) => { const id = ++sequence; timers.set(id, { callback, milliseconds }); return id; };
  const streams = () => { const stream = new EventEmitter(); stream.destroy = () => { stream.destroyed = true; }; return stream; };
  const processes = {
    spawn(executable, args, options) {
      const record = { executable, args: Array.from(args), options, payload: undefined };
      records.push(record);
      const child = new EventEmitter();
      child.pid = 4321;
      child.exitCode = null;
      child.signalCode = null;
      child.stdout = streams(); child.stderr = streams(); child.stdin = streams();
      record.child = child;
      child.kill = () => { child.killed = true; };
      if (executable.endsWith('taskkill.exe')) {
        queueMicrotask(() => { child.exitCode = 0; child.emit('exit', 0); });
      }
      child.stdin.end = payload => {
        record.payload = payload;
        if (!timeout) queueMicrotask(() => {
          child.stdout.emit('data', Buffer.from('native response ü'));
          child.exitCode = 0; child.emit('close', 0);
        });
      };
      return child;
    },
    execFile(executable, args, options, callback) {
      const record = { executable, args: Array.from(args), options, payload: undefined };
      records.push(record);
      queueMicrotask(() => callback(null, 'native response ü'));
      return { stdin: { end(payload) { record.payload = payload; } } };
    }
  };
  const context = vm.createContext({
    require(name) { assert.equal(name, 'node:child_process'); return processes; },
    exports: {}, process: { platform, env: { SystemRoot: 'C:\\Windows' } },
    setTimeout: timer, clearTimeout: id => timers.delete(id), Buffer, console
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
  if (windows) {
    assert.equal(records[0].options.windowsHide, true);
    const deadline = Array.from(timers.values()).find(item => item.milliseconds === 5000);
    assert.ok(deadline, 'native Windows handoff must retain its bounded deadline');
    if (timeout) {
      deadline.callback();
      await assert.rejects(promise, /timed out/);
      assert.equal(records.length, 2);
      assert.equal(records[1].executable, 'C:\\Windows\\System32\\taskkill.exe');
      assert.deepEqual(records[1].args, ['/PID', '4321', '/T', '/F']);
      for (const stream of ['stdin', 'stdout', 'stderr']) assert.equal(records[0].child[stream].destroyed, true);
    } else assert.equal(await promise, 'native response ü');
  } else {
    assert.equal(records[0].options.timeout, 1500);
    assert.equal(records[0].options.maxBuffer, 65536);
    assert.equal(await promise, 'native response ü');
  }
}

(async () => {
  for (const platform of ['win32', 'linux']) {
    for (const operation of ['receive', 'request', 'bind']) await exercise(platform, operation);
  }
  await exercise('win32', 'receive', true);
  console.log('Native plugin argv/payload controls and bounded Windows cleanup passed; Linux VM construction only.');
})().catch(error => { console.error(error); process.exitCode = 1; });
