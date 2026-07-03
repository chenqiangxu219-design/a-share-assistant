const path = require('path');
const fs = require('fs');
const vm = require('vm');
const Module = require('module');
const { spawnSync } = require('child_process');

// ===== Approach 1: Try individual electron: API imports =====
console.log('=== Individual API imports ===');
const individualImportScript = `
(async () => {
  const apis = ['electron/app', 'electron/browser-window', 'electron/screen', 'electron/ipc-main', 'electron/ipc-main-init', 'electron/cli-app-args'];
  for (const api of apis) {
    try {
      const mod = await import(api);
      const keys = Object.keys(mod).slice(0, 5).join(',');
      console.log('[import] ' + api + ':', typeof mod.default, keys);
    } catch (e) {
      console.log('[import] ' + api + ':', e.code || e.name, e.message.substring(0, 80));
    }
  }
  
  // Also try the common patterns
  const patterns = ['electron/main', 'electron/common', 'electron/renderer'];
  for (const p of patterns) {
    try {
      const mod = await import(p);
      const keys = Object.keys(mod).slice(0, 8).join(',');
      console.log('[import] ' + p + ':', keys);
    } catch (e) {
      console.log('[import] ' + p + ':', e.code || e.name, e.message.substring(0, 80));
    }
  }
  process.exit(0);
})();
`;

const innerScript = path.join(__dirname, 'vm_inner.mjs');
fs.writeFileSync(innerScript, individualImportScript);

const r1 = spawnSync(ELECTRON, [innerScript], {
  env: { ...process.env },
  encoding: 'utf8',
  maxBuffer: 500000,
});
console.log(r1.stdout);
if (r1.stderr) console.log('stderr:', r1.stderr.substring(0, 300));

// ===== Approach 2: Read main.js from ASAR using process.noAsar =====
console.log('\n=== Reading ASAR with noAsar ===');
const noAsarScript = `
const path = require('path');
const fs = require('fs');

// Temporarily disable ASAR interception
const prevNoAsar = process.noAsar;
process.noAsar = true;

const asarPath = path.join(__dirname, '..', 'node_modules', 'electron', 'dist', 'Electron.app', 'Contents', 'Resources', 'default_app.asar');

try {
  const content = fs.readFileSync(asarPath, 'utf-8');
  console.log('[noAsar] read', content.length, 'bytes');
  
  // Parse ASAR header to find main.js
  const entries = [];
  const regex = /"path":"([^"]*)","offset":(\\d+),"size":(\\d+)/g;
  let m;
  while ((m = regex.exec(content)) !== null) {
    entries.push({ path: m[1], offset: parseInt(m[2]), size: parseInt(m[3]) });
  }
  console.log('[noAsar] entries:', entries.length);
  
  const mainEntry = entries.find(e => e.path === 'main.js');
  if (mainEntry) {
    const mainJs = content.substring(mainEntry.offset, mainEntry.offset + mainEntry.size);
    console.log('[noAsar] main.js:', mainJs.length, 'chars');
    console.log('\\n[main.js content]:\\n', mainJs);
  }
} finally {
  process.noAsar = prevNoAsar;
}
process.exit(0);
`;

const noAsarPath = path.join(__dirname, 'no_asar_test.js');
fs.writeFileSync(noAsarPath, noAsarScript);

const r2 = spawnSync(ELECTRON, [noAsarPath], {
  env: { ...process.env },
  encoding: 'utf8',
  maxBuffer: 500000,
});
console.log(r2.stdout);
if (r2.stderr) console.log('stderr:', r2.stderr.substring(0, 300));

// ===== Approach 3: vm.SourceTextModule =====
console.log('\n=== vm.SourceTextModule test ===');
const vmScript = `
const path = require('path');
const fs = require('fs');
const vm = require('vm');

// First read main.js
process.noAsar = true;
const asarPath = path.join(__dirname, '..', 'node_modules', 'electron', 'dist', 'Electron.app', 'Contents', 'Resources', 'default_app.asar');
const content = fs.readFileSync(asarPath, 'utf-8');
const entries = [];
const regex = /"path":"([^"]*)","offset":(\\d+),"size":(\\d+)/g;
let m;
while ((m = regex.exec(content)) !== null) {
  entries.push({ path: m[1], offset: parseInt(m[2]), size: parseInt(m[3]) });
}
const mainEntry = entries.find(e => e.path === 'main.js');
const mainJs = content.substring(mainEntry.offset, mainEntry.offset + mainEntry.size);
process.noAsar = false;

// Create a SourceTextModule with custom import handler
const context = vm.createContext({
  process,
  console,
  setTimeout,
  setInterval,
  setImmediate,
  clearTimeout,
  clearInterval,
  clearImmediate,
  __dirname: path.dirname(noAsarPath),
  __filename: noAsarPath,
});

// The customImportModule should handle 'electron/main' imports
// But we need to get the bindings first...
// What if we use a two-phase approach?

// Phase 1: Get bindings by importing electron/main through ESM
// Phase 2: Use those bindings to evaluate main.js

const bindingsPromise = import('electron/main');
// This will crash. But can we catch it?
// The crash is in cjsPreparseModuleExports which is synchronous within the async import

// Actually, let's try vm.SourceTextModule for the electron/main module
// We need the SOURCE of electron/main, which is in js2c

// Get js2c source by reading the node_init output
// The node_init module IS the js2c output. We can access it from Module._cache
const nodeInitKey = Object.keys(Module._cache).find(k => k.includes('node_init'));
console.log('[vm] node_init key:', nodeInitKey);

// The node_init module is compiled from js2c. Its source is embedded.
// Can we get it? It's not a file - it's compiled from C++ strings.
// But the ESM loader can GET the source for electron: protocol.

// Try: create a SourceTextModule for main.js and handle imports
const ssm = new vm.SourceTextModule(mainJs, {
  identifier: 'default_app.asar/main.js',
  importModuleDynamically(specifier, module) {
    console.log('[vm] import requested:', specifier);
    if (specifier === 'electron/main') {
      // We need to return a Promise<SourceTextModule> with the electron/main content
      // But we don't have the source! We need to extract it.
      // For now, return a module with synthetic bindings
      const syntheticSrc = \`
        const app = { name: 'synthetic', whenReady: () => Promise.resolve() };
        const BrowserWindow = function() {};
        const screen = { getPrimaryDisplay: () => ({ workAreaSize: { width: 1920, height: 1080 } }) };
        const ipcMain = { on: () => {}, emit: () => {} };
        export { app, BrowserWindow, screen, ipcMain };
      \`;
      return Promise.resolve(new vm.SourceTextModule(syntheticSrc, { identifier: specifier }));
    }
    // For node: modules, use require
    if (specifier.startsWith('node:')) {
      const mod = require(specifier);
      const src = 'const m = ' + JSON.stringify(mod) + '; export default m;';
      return Promise.resolve(new vm.SourceTextModule(src, { identifier: specifier }));
    }
    return Promise.reject(new Error('Unknown import: ' + specifier));
  },
});

ssm.evaluate().then(() => {
  console.log('[vm] main.js evaluated successfully!');
  console.log('[vm] exports:', Object.keys(ssm.namespace).join(','));
}).catch(e => {
  console.log('[vm] evaluation error:', e.message);
});
`;

const vmTestPath = path.join(__dirname, 'vm_test_inner.js');
fs.writeFileSync(vmTestPath, vmScript);

const r3 = spawnSync(ELECTRON, [vmTestPath], {
  env: { ...process.env },
  encoding: 'utf8',
  maxBuffer: 500000,
});
console.log(r3.stdout);
if (r3.stderr) console.log('stderr:', r3.stderr.substring(0, 300));

// Cleanup
fs.unlinkSync(innerScript);
fs.unlinkSync(noAsarPath);
fs.unlinkSync(vmTestPath);
process.exit(0);
