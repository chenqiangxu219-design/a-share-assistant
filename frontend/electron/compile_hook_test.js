// Test: intercept electron: module compilation via Module._compile hook
// The idea: when the ESM loader evaluates electron:main's js2c webpack source,
// it goes through CJS instantiation. If we hook _compile, we can transform the source
// to fix the cjsPreparseModuleExports crash.

const Module = require('node:module');
const vm = require('node:vm');
const path = require('node:path');
const fs = require('node:fs');

// ===== Step 1: Understand the crash =====
// cjsPreparseModuleExports does: Object.keys(module.exports)
// But webpack source uses: t[id] = {exports: {}}
// The actual exports are in t[ENTRY_ID].exports, not module.exports

// ===== Step 2: Hook _compile to transform webpack source =====
const originalCompile = Module._compile;

Module._compile = function(content, filename) {
  // Check if this is a js2c webpack module
  if (content.includes('function(e) {') && content.includes('t[') && content.includes('{exports: {}')) {
    console.log('[compile] intercepted webpack module:', filename);
    console.log('[compile] length:', content.length);

    // Find the entry point and transform
    // Webpack pattern: (function(e) { ... t[ENTRY](); return t[ENTRY].exports; })(modules)
    // We need to make module.exports = t[ENTRY].exports

    // Find "return t[X].exports" at the end
    const returnMatch = content.match(/return t\[(\d+)\]\.exports\s*;?\s*\)/);
    if (returnMatch) {
      const entryId = returnMatch[1];
      console.log('[compile] entry module:', entryId);

      // Transform: add "var module = {exports: {}}; var exports = {};" at the start
      // And at the end: "module.exports = t[ENTRY].exports; exports = module.exports;"
      const transformed = content + '\nmodule.exports = (typeof module !== "undefined" ? module.exports : {});';
      console.log('[compile] applying transform...');
      return originalCompile.call(this, transformed, filename);
    }
  }
  return originalCompile.call(this, content, filename);
};

// ===== Step 3: Try require('electron/main') =====
console.log('[test] trying require approach...');

// But wait - require('electron/main') goes through _load -> _resolveFilename -> MODULE_NOT_FOUND
// The electron:main is an ESM-only protocol. Let's try _load directly.

// First, let's see if we can get the electron:main source somehow
console.log('\n[test] trying to trace _load for electron modules...');

const traceLoad = Module._load;
Module._load = function(request, parent, isMain) {
  console.log('[load]', JSON.stringify(request), 'from', parent?.id);
  try {
    return traceLoad.apply(this, arguments);
  } catch (e) {
    console.log('[load] ERROR:', e.code, e.message.substring(0, 80));
    throw e;
  }
};

// Try different specifiers
for (const spec of ['electron', 'electron/main', 'app', 'screen']) {
  console.log('\n--- require("' + spec + '") ---');
  try {
    const m = require(spec);
    console.log('[result] type:', typeof m, 'keys:', Object.keys(m).slice(0, 5).join(','));
  } catch (e) {
    // expected for some
  }
}

// ===== Step 4: Try the ESM import from CJS =====
console.log('\n\n=== ESM import from CJS ===');
const runESM = async () => {
  try {
    console.log('[esm] importing electron/main...');
    const mod = await import('electron/main');
    console.log('[esm] SUCCESS! type:', typeof mod, 'keys:', Object.keys(mod).slice(0, 10).join(','));
    return mod;
  } catch (e) {
    console.log('[esm] CRASH:', e.code || e.name, '-', e.message.substring(0, 200));
  }
};

runESM().then(() => {
  setTimeout(() => process.exit(0), 500);
});
