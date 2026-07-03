// Test: Can we create bridge files that work?
// Strategy: create node_modules/electron/main.js that exports the APIs
// We need to figure out how to get the actual bindings

const Module = require('node:module');
const path = require('node:path');
const fs = require('node:fs');
const vm = require('node:vm');
const { spawnSync } = require('node:child_process');

const electronDist = path.join(__dirname, '..', 'node_modules', 'electron', 'dist');
const asarPath = path.join(electronDist, 'Electron.app', 'Contents', 'Resources', 'default_app.asar');
const electronBinary = path.join(electronDist, 'Electron.app', 'Contents', 'MacOS', 'Electron');

// ===== Step 1: Extract main.js from ASAR =====
console.log('=== Extracting default_app.asar/main.js ===');
process.noAsar = true;
const asarContent = fs.readFileSync(asarPath, 'utf-8');
process.noAsar = false;

// Parse ASAR header
const filesMatch = asarContent.match(/\{"files"\:(\{.*?\})\}/s);
if (!filesMatch) {
  console.log('[ERROR] Could not parse ASAR');
  process.exit(1);
}

const filesJson = '{"files":' + filesMatch[1];
const files = JSON.parse(filesJson).files;
console.log('[asar] entries:', Object.keys(files).join(', '));

const mainEntry = files['main.js'];
const mainJs = asarContent.substring(mainEntry.offset, mainEntry.offset + mainEntry.size);
console.log('[asar] main.js length:', mainJs.length);
console.log('[asar] main.js first 500 chars:');
console.log(mainJs.substring(0, 500));

// ===== Step 2: Analyze the webpack structure =====
console.log('\n=== Analyzing webpack structure ===');
// Find the webpack runtime function
const webpackMatch = mainJs.match(/\(function\(e\)\s*\{[^}]*var t\s*=\s*\{\}/s);
if (webpackMatch) {
  console.log('[webpack] found runtime:', webpackMatch[0].substring(0, 200));
}

// Find module definitions - look for t[x] = {exports: {}}
const moduleDefs = mainJs.match(/t\[\d+\]\s*=\s*\{exports:\s*\{\}\}/g);
if (moduleDefs) {
  console.log('[webpack] module definitions found:', moduleDefs.length);
  console.log('[webpack] first few:', moduleDefs.slice(0, 5).join(', '));
}

// Find the entry point - look for the last e[x] call which is the entry
const entryMatch = mainJs.match(/e\[(\d+)\]\s*\(\)\s*\)/);
if (entryMatch) {
  console.log('[webpack] entry module:', entryMatch[1]);
}

// ===== Step 3: Test approach - patch the electron package =====
console.log('\n=== Testing bridge file approach ===');

// The idea: create node_modules/electron/main.js as a CJS file
// that exports what we need. The challenge: getting the real bindings.

// Approach A: Use require() from the ASAR context
console.log('[A] createRequire from ASAR:');
const qr = Module.createRequire(asarPath + '/main.js');
for (const name of ['fs', 'path', 'url', 'events', 'timers']) {
  try {
    const m = qr(name);
    console.log('[A]  ' + name + ': OK, keys:', Object.keys(m).slice(0, 3).join(','));
  } catch (e) {
    console.log('[A]  ' + name + ': ' + e.code);
  }
}

// Approach B: Read the js2c source from the binary
console.log('\n[B] Try to get js2c source via ESM getFormat/getSource:');
// We need to run this inside Electron's ESM context
const esmTestScript = `
(async () => {
  const { getFormat, getSource } = require('internal_module_hooks') || {};
  console.log('[esm] internal_module_hooks available:', !!getFormat);

  // Try the module loader API
  const { register } = await import('node:module');
  console.log('[esm] register available:', typeof register);

  // Check what the ESM loader sees
  const currentLoader = await import('electron:main');
  console.log('[esm] electron:main type:', typeof currentLoader);
})().catch(e => console.log('[esm] error:', e.message.substring(0, 100)));
`;

// ===== Step 4: The real test - can we eval default_app.asar/main.js? =====
console.log('\n=== Evaluating default_app.asar/main.js in a context ===');

// The key insight: main.js starts with:
// import * as electron from 'electron/main';
// If we can make THAT import work, everything else follows.

// Strategy: use vm.SourceTextModule with custom import resolver
// that returns the REAL electron bindings for 'electron/main'

// But we need the real bindings! How?
// The bindings come from the js2c compiled modules.
// When Electron's ESM loader handles 'electron/main', it:
// 1. Resolves to electron:main protocol
// 2. Gets the js2c source (webpack format)
// 3. Evaluates it and gets the exports

// Can we replicate steps 2-3?

// Step 2: Get the js2c source
// We know it's embedded in the binary. Can we extract it?
// Method: the ESM loader can get it via the electron: protocol

const extractScript = `
const path = require('path');
const fs = require('fs');
const vm = require('vm');

// Use the ESM loader to get the source of electron:main
// We do this by creating a custom loader that intercepts the source

// Actually, simpler approach: the js2c modules are compiled from
// the webpack output of electron's build. They're stored as C++ strings.
// We can access them via the node_init binding.

// Check: what does the ESM loader's getSource return for electron:main?
// We need to use the experimental module loader API

process.noAsar = true;
const asarPath = path.join(process.resourcesPath, 'default_app.asar');
const asarContent = fs.readFileSync(asarPath, 'utf-8');
process.noAsar = false;

// Get main.js from ASAR
const filesMatch = asarContent.match(/\\{"files"\\:\\{(.*?)\\}\\}/s);
const files = JSON.parse('\\{"files":' + filesMatch[1] + '}').files;
const mainEntry = files['main.js'];
const mainJs = asarContent.substring(mainEntry.offset, mainEntry.offset + mainEntry.size);

// Now: main.js has "import * as electron from 'electron/main'"
// We want to eval it with a custom context where that import resolves to real bindings

// The trick: use vm.SourceTextModule
// 1. Create a context with all the globals Electron provides
// 2. Create a SourceTextModule for main.js
// 3. Handle the 'electron/main' import by returning a synthetic module
// 4. The synthetic module needs to export: app, BrowserWindow, screen, ipcMain, etc.

// But we need the REAL bindings, not synthetic ones!
// The real bindings come from evaluating the js2c webpack modules.

// Can we get the js2c source? Let's try the electron: protocol via ESM import
// and capture the result before cjsPreparseModuleExports crashes.

// Idea: the crash happens in cjsPreparseModuleExports which is called AFTER
// the source is fetched. So we might be able to intercept the source.

// Method: use register() to install a custom loader
const { register } = require('node:module');

// Create a loader script
const loaderCode = \`
let capturedSource = null;

async function load(url, context, nextLoad) {
  const result = await nextLoad(url, context);
  if (url.includes('electron:') && result.source) {
    capturedSource = result.source;
    console.log('[loader] captured:', url, '- length:', result.source.length);
    console.log('[loader] first 300:', result.source.substring(0, 300));
  }
  return result;
}

// Export for Node's loader API
module.exports = { load };

// Make capturedSource accessible
process.on('message', (msg) => {
  if (msg === 'get_source' && capturedSource) {
    process.send(capturedSource.substring(0, 5000));
    process.exit(0);
  }
});
\`;

const loaderPath = path.join(__dirname, 'electron', 'source_capture_loader.js');
fs.writeFileSync(loaderPath, loaderCode);

// Run a child process with the loader
const electronPath = path.join(process.resourcesPath, '..', 'MacOS', 'Electron');
const childScript = path.join(__dirname, 'electron', 'capture_child.js');
fs.writeFileSync(childScript, 'import("electron/main").catch(() => {}); setTimeout(() => process.exit(0), 1000);');

console.log('[extract] Running child with loader...');
const result = spawnSync(electronPath, [childScript], {
  env: { ...process.env, NODE_OPTIONS: '--loader file://' + loaderPath },
  encoding: 'utf8',
  maxBuffer: 1000000,
  stdio: 'pipe',
});

console.log('[extract] stdout:', result.stdout);
if (result.stderr) console.log('[extract] stderr:', result.stderr.substring(0, 500));

// Cleanup
fs.unlinkSync(loaderPath);
fs.unlinkSync(childScript);
process.exit(0);
`;

const extractPath = path.join(__dirname, 'electron', 'extract_js2c.js');
fs.writeFileSync(extractPath, extractScript);

const result = spawnSync(electronBinary, [extractPath], {
  encoding: 'utf8',
  maxBuffer: 1000000,
  cwd: path.join(__dirname, '..'),
});
console.log(result.stdout);
if (result.stderr) console.log('stderr:', result.stderr.substring(0, 500));

fs.unlinkSync(extractPath);
process.exit(0);
