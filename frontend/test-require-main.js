const Module = require('module');
console.log('[test] Module._cache size:', Object.keys(Module._cache).length);

// Check if 'electron' is in the cache
const electronPath = require.resolve('electron');
console.log('[test] resolved electron:', electronPath);
console.log('[test] cached:', Module._cache[electronPath] ? 'YES' : 'NO');

const mod = require('electron');
console.log('[test] require type:', typeof mod);
console.log('[test] require value:', typeof mod === 'string' ? mod : Object.keys(mod || {}).join(','));

// After require, check cache again
console.log('[test] cached after require:', Module._cache[electronPath] ? 'YES' : 'NO');
if (Module._cache[electronPath]) {
  console.log('[test] cache exports type:', typeof Module._cache[electronPath].exports);
}
