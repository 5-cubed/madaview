import { spawn, execSync } from 'node:child_process';
import { mkdir, rm, writeFile } from 'node:fs/promises';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const LIB_DIR = path.dirname(fileURLToPath(import.meta.url));
export const REPO_ROOT = path.resolve(LIB_DIR, '..', '..');
export const BIN_PATH = path.join(REPO_ROOT, 'e2e', '.bin', 'madaview');

// Builds the madaview binary (frontend included, via `make build`) once and
// reuses it across scenario runs. Each scenario's `run` only resets its own
// result/ dir (see resetResultDir) — rebuilding the binary per scenario
// would be redundant, since scenarios don't modify source. A plain
// `go build` isn't enough here: scenarios that drive a real browser need
// the actual embedded frontend, not the go:embed placeholder.
export async function ensureBinary() {
  if (!existsSync(BIN_PATH)) {
    await runCmd('make', ['build'], { cwd: REPO_ROOT });
    await mkdir(path.dirname(BIN_PATH), { recursive: true });
    await runCmd('cp', [path.join(REPO_ROOT, 'madaview'), BIN_PATH], { cwd: REPO_ROOT });
  }
  return BIN_PATH;
}

function runCmd(cmd, args, opts) {
  return new Promise((resolve, reject) => {
    const child = spawn(cmd, args, { stdio: 'inherit', ...opts });
    child.on('error', reject);
    child.on('exit', (code) => (code === 0 ? resolve() : reject(new Error(`${cmd} ${args.join(' ')} exited ${code}`))));
  });
}

// Resets a scenario's result/ directory to a clean, empty state.
export async function resetResultDir(scenarioDir) {
  const resultDir = path.join(scenarioDir, 'result');
  await rm(resultDir, { recursive: true, force: true });
  await mkdir(resultDir, { recursive: true });
  return resultDir;
}

export function nextPort() {
  return 20000 + Math.floor(Math.random() * 20000);
}

// Starts a madaview server instance rooted at `root`, capturing its stdout
// logs, and waits until /api/status responds before returning. Pass
// root: null to omit --root entirely, so the binary falls through to its
// own priority: persisted config, then cwd default.
export async function startServer({ root, port, extraArgs = [], env = {} }) {
  const bin = await ensureBinary();
  const logs = [];
  const rootArgs = root === null ? [] : ['--root', root];
  const child = spawn(bin, [...rootArgs, '--port', String(port), '--verbose', ...extraArgs], {
    env: { ...process.env, ...env },
  });
  child.stdout.on('data', (d) => logs.push(d.toString()));
  child.stderr.on('data', (d) => logs.push(d.toString()));

  await waitForReady(port);

  return {
    port,
    getLogs: () => logs.join(''),
    stop: () =>
      new Promise((resolve) => {
        child.once('exit', resolve);
        child.kill();
      }),
  };
}

async function waitForReady(port, timeoutMs = 8000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await fetch(`http://localhost:${port}/api/status`);
      if (res.ok) return;
    } catch {
      // server not up yet
    }
    await new Promise((r) => setTimeout(r, 100));
  }
  throw new Error(`server did not become ready on port ${port} within ${timeoutMs}ms`);
}

export async function writeMetadata(resultDir, extra = {}) {
  const metadata = {
    timestamp: new Date().toISOString(),
    os: process.platform,
    arch: process.arch,
    commit: safeGitCommit(),
    ...extra,
  };
  await writeFile(path.join(resultDir, 'metadata.json'), JSON.stringify(metadata, null, 2));
  return metadata;
}

function safeGitCommit() {
  try {
    return execSync('git rev-parse HEAD', { cwd: REPO_ROOT }).toString().trim();
  } catch {
    return 'unknown';
  }
}

export async function writeJSON(resultDir, name, data) {
  await writeFile(path.join(resultDir, name), JSON.stringify(data, null, 2));
}

export async function writeText(resultDir, name, content) {
  await writeFile(path.join(resultDir, name), content);
}

export function readResultJSON(resultDir, name) {
  return JSON.parse(readFileSync(path.join(resultDir, name), 'utf8'));
}

// Report helper: verify scripts call this to emit a consistent good /
// unexpected / ambiguous verdict and exit with a matching status code.
export function report(verdict, message) {
  const prefix = { good: 'GOOD', unexpected: 'UNEXPECTED', ambiguous: 'AMBIGUOUS' }[verdict];
  if (!prefix) throw new Error(`unknown verdict: ${verdict}`);
  console.log(`[${prefix}] ${message}`);
  process.exit(verdict === 'unexpected' ? 1 : 0);
}
