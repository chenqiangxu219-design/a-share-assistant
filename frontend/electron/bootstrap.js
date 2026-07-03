// Bootstrap: use import('electron/main') from CJS and cache the result
const Module = require('module');
const path = require('path');

const npmIndexPath = Module._resolveFilename('electron', module);

// Proxy that spin-waits for real bindings
let realBindings = null;
const proxy = new Proxy({}, {
  get(target, prop) {
    if (realBindings) return realBindings[prop];
    const deadline = Date.now() + 5000;
    while (!realBindings && Date.now() < deadline) {}
    return realBindings ? realBindings[prop] : undefined;
  },
  getOwnPropertyDescriptor(target, prop) {
    if (realBindings && prop in realBindings) {
      return { configurable: true, enumerable: true, value: realBindings[prop], writable: true };
    }
    return undefined;
  },
  ownKeys(target) { return realBindings ? Object.keys(realBindings) : []; },
  has(target, prop) { return realBindings ? prop in realBindings : false; },
});
Module._cache[npmIndexPath] = { exports: proxy };

// Async import of electron/main
// We use setImmediate so the main script can start loading (and hit our proxy)
// while the import completes in the background
setImmediate(() => {
  // This goes through Node's ESM loader which has Electron's custom hooks
  Promise.resolve()
    .then(() => import('electron/main'))
    .then((mod) => {
      console.log('[bootstrap] electron/main import succeeded');
      const keys = Object.keys(mod);
      console.log('[bootstrap] keys:', keys.join(','));
      console.log('[bootstrap] app:', typeof mod.app);
      console.log('[bootstrap] BrowserWindow:', typeof mod.BrowserWindow);

      if (keys.length === 0) {
        // Empty object — check if it's using getters or Symbol properties
        console.log('[bootstrap] desc of app:', Object.getOwnPropertyDescriptor(mod, 'app'));
        console.log('[bootstrap] getOwnPropertyNames:', Object.getOwnPropertyNames(mod).join(','));
        console.log('[bootstrap] symbols:', Object.getOwnPropertySymbols(mod).map(s => s.toString()).join(','));
      }

      if (mod.app) {
        realBindings = mod;
        console.log('[bootstrap] SUCCESS!');
      }
    })
    .catch((e) => {
      console.log('[bootstrap] electron/main import error:', e.message);
    });

  // Also try bare 'electron' import
  Promise.resolve()
    .then(() => import('electron'))
    .then((mod) => {
      console.log('[bootstrap] bare electron import:');
      const keys = Object.keys(mod);
      console.log('[bootstrap] keys:', keys.join(','));
      if (mod.default && typeof mod.default === 'object') {
        console.log('[bootstrap] default keys:', Object.keys(mod.default).join(','));
      }
      if (mod.app && !realBindings) {
        realBindings = mod;
        console.log('[bootstrap] Using bare electron!');
      }
    })
    .catch((e) => {
      console.log('[bootstrap] bare electron import error:', e.message);
    });
})
.then(() => {
  // Wait for imports to complete
  return new Promise(resolve => setTimeout(resolve, 3000));
})
.then(() => {
  console.log('[bootstrap] Final: realBindings =', realBindings ? Object.keys(realBindings).join(',') : 'NONE');
});

// Keep process alive
setTimeout(() => process.exit(0), 5000);
