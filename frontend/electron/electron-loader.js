// Custom ESM loader to fix electron/main CJS interop crash
// Usage: NODE_OPTIONS='--experimental-loader ./electron/electron-loader.js' electron .

const path = require('path');
const fs = require('fs');

const JS2C_WEPACK_PATTERN = /function\(e\)\s*\{/;
const WEBPACK_MODULE_EXPORTS_PATTERN = /t\[\w+\]\s*=\s*\{exports:\s*\{\}\}/;

/**
 * Transform js2c webpack CJS source to be compatible with Node's ESM loader.
 *
 * The problem: js2c webpack modules use `e[id] = {exports: {}}` pattern.
 * Node's cjsPreparseModuleExports expects `module.exports` to be an object.
 *
 * The fix: wrap the module so that `exports` is pre-defined as an object,
 * and after evaluation, copy `exports` to `module.exports`.
 */
function transformJs2cSource(source) {
  // The webpack output looks like:
  // (function(e) {
  //   var t = {};
  //   function r(n) { ... }
  //   t[0] = {exports: {}};
  //   // ... module code that sets t[0].exports = something ...
  //   t[1] = {exports: {}};
  //   // ...
  //   e[ENTRY_ID]();  // execute entry
  //   return t[ENTRY_ID].exports;
  // })(webpackModules);
  //
  // cjsPreparseModuleExports does: Object.keys(module.exports)
  // But module.exports is undefined because the webpack code uses e[id].exports

  // Strategy: prepend a wrapper that sets module.exports = {} before webpack runs
  const wrapper = `
var exports = {};
var module = { exports: exports };
`;
  // Append: after webpack, ensure module.exports has the right value
  const append = `
// Ensure module.exports is set for CJS interop
if (typeof exports === 'object' && exports !== null && Object.keys(exports).length > 0) {
  module.exports = exports;
}
`;
  return wrapper + source + append;
}

async function resolve(specifier, context, nextResolve) {
  const result = await nextResolve(specifier, context);
  return result;
}

async function load(url, context, nextLoad) {
  const result = await nextLoad(url, context);

  // Detect js2c webpack modules
  if (result.format === 'commonjs' && result.source &&
      JS2C_WEPACK_PATTERN.test(result.source) &&
      WEBPACK_MODULE_EXPORTS_PATTERN.test(result.source)) {
    console.error('[electron-loader] transforming:', url);
    return {
      ...result,
      source: transformJs2cSource(result.source),
    };
  }

  return result;
}

module.exports = { resolve, load };
