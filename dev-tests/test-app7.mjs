import { createRequire } from 'module';

// The trick: the default app already did `import * as electron from 'electron/main'`
// Before loading our code. So the ESM registry should have it cached.
// But ESM cache is accessed by file URL, not package name.

// Let's try to find the actual resolved path
import { resolve as _resolve } from 'import-meta-resolve';
try {
  const resolved = await _resolve('electron/main', import.meta.url);
  console.log('resolved electron/main:', resolved);
} catch(e) {
  console.log('resolve error:', e.message);
}

// Another approach: check if the default app's electron import is accessible
// via the Module namespace
import * as module from 'node:module';
console.log('Module exports:', Object.getOwnPropertyNames(module).filter(k => k.includes('cache') || k.includes('resolve') || k.includes('register')).join(','));

// Try Module._esmCache or similar
if (module._getESMCache) console.log('esmCache:', module._getESMCache());

process.exit(0);
