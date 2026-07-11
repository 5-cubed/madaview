import path from 'node:path';
import { spawnSync } from 'node:child_process';

// Shared `test` entry point: run then verify, in one command. Each
// scenario's `test` file is a one-liner that calls this with its own
// directory (import.meta.url).
export function runThenVerify(scenarioDir) {
  for (const script of ['run', 'verify']) {
    const res = spawnSync(process.execPath, [path.join(scenarioDir, script)], { stdio: 'inherit' });
    if (res.status !== 0) process.exit(res.status ?? 1);
  }
}
