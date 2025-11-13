const fs = require('fs');
const path = require('path');
const { spawn } = require('child_process');

function dirStrToCode(s) {
  s = String(s).toUpperCase().trim();
  switch (s) {
    case 'N': return 0;
    case 'NE': return 1;
    case 'E': return 2;
    case 'SE': return 3;
    case 'S': return 4;
    case 'SW': return 5;
    case 'W': return 6;
    case 'NW': return 7;
    default: throw new Error('Unknown direction ' + s);
  }
}

function parseSamples(inputPath) {
  const tokens = fs.readFileSync(inputPath, 'utf8').trim().split(/\s+/);
  let i = 0;
  const T = parseInt(tokens[i++], 10);
  const cases = [];
  for (let t = 0; t < T; t++) {
    const N = parseInt(tokens[i++], 10);
    const players = [];
    for (let k = 0; k < N; k++) {
      const x = BigInt(tokens[i++]);
      const y = BigInt(tokens[i++]);
      players.push({ x: Number(x), y: Number(y), alive: true });
    }
    const dirStr = tokens[i++];
    const s = parseInt(tokens[i++], 10) - 1; // convert to 0-based
    cases.push({ players, startDirection: dirStrToCode(dirStr), startIndex: s });
  }
  return cases;
}

function parseOutputs(outputPath) {
  const tokens = fs.readFileSync(outputPath, 'utf8').trim().split(/\s+/);
  let i = 0;
  const out = [];
  while (i < tokens.length) {
    const a = tokens[i++];
    if (a === undefined || a === '') break;
    const b = tokens[i++];
    if (b === undefined) break;
    out.push({ throws: parseInt(a, 10), lastPlayer: parseInt(b, 10) });
  }
  return out;
}

async function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }

async function run() {
  // start Go server
  const srv = spawn('go', ['run', './apps/dodgeball-go/server'], { stdio: ['ignore', 'pipe', 'pipe'], cwd: path.resolve(__dirname, '..') });
  let ready = false;
  srv.stdout.on('data', (d) => {
    const s = d.toString();
    if (s.includes('Dodgeball gRPC server running')) ready = true;
    process.stdout.write(s);
  });
  srv.stderr.on('data', (d) => process.stderr.write(d.toString()));

  // wait up to 2s until server ready
  for (let t = 0; t < 20 && !ready; t++) await sleep(100);
  if (!ready) {
    console.error('Server not ready, continuing anyway...');
  }

  // Load ts-node from the TS app's node_modules to avoid root dependency
  require(path.join(__dirname, '..', 'apps', 'dodgeball-ts', 'node_modules', 'ts-node')).register({ transpileOnly: true });
  const { runSimulationTS } = require('../apps/dodgeball-ts/src/client.ts');

  let allOk = true;
  const samplesDir = path.join(__dirname, 'samples');
  for (const name of ['sample2']) {
    const inPath = path.join(samplesDir, name + '.in');
    const outPath = path.join(samplesDir, name + '.out');
    const cases = parseSamples(inPath);
    const expected = parseOutputs(outPath);

    console.log(`\n=== ${name} ===`);
    for (let i = 0; i < cases.length; i++) {
      const res = await runSimulationTS(cases[i]);
      const exp = expected[i];
      const gotThrows = res.throws;
      const gotLast = res.lastPlayer + 1; // convert to 1-based for comparison
      const ok = gotThrows === exp.throws && gotLast === exp.lastPlayer;
      console.log(`#${i+1} -> expected ${exp.throws} ${exp.lastPlayer} | got ${gotThrows} ${gotLast} ${ok ? 'OK' : 'FAIL'}`);
      if (!ok) allOk = false;
    }
  }

  srv.kill('SIGTERM');
  if (!allOk) {
    process.exitCode = 1;
  }
}

run().catch((e) => { console.error(e); process.exitCode = 1; });
