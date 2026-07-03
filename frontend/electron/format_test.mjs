// This file is pure ESM - no CJS interop needed
// We'll run it directly with electron which has the custom loader

// Check if we can import electron/main and get actual bindings
try {
  const electron = await import('electron/main');
  console.log('[format] import succeeded');
  console.log('[format] type:', typeof electron);
  console.log('[format] keys:', Object.keys(electron).slice(0, 10).join(','));
  console.log('[format] app:', typeof electron.app);
  console.log('[format] BrowserWindow:', typeof electron.BrowserWindow);
  
  // The key question: is it a namespace export or CJS wrapped?
  console.log('[format] __esModule:', electron.__esModule);
  console.log('[format] default:', typeof electron.default);
  if (typeof electron.default === 'object' && electron.default) {
    console.log('[format] default.keys:', Object.keys(electron.default).slice(0, 10).join(','));
  }
} catch (e) {
  console.log('[format] Error:', e.name, e.message);
  console.log('[format] stack:', e.stack.split('\n').slice(0, 3).join('\n'));
}

// Also try the physical ASAR path
try {
  const path = await import('node:path');
  const url = await import('node:url');
  const fs = await import('node:fs');
  
  // Find electron binary
  const electronDist = path.default.join(url.default.fileURLToPath(import.meta.url), '../../node_modules/electron/dist');
  const asarPath = path.default.join(electronDist, 'Electron.app', 'Contents', 'Resources', 'default_app.asar');
  console.log('\n[format] asarPath:', asarPath, 'exists:', fs.default.existsSync(asarPath));
  
  // Try importing from the ASAR path
  const asarMain = await import('file://' + asarPath + '/main.js');
  console.log('[format] ASAR import type:', typeof asarMain);
  console.log('[format] ASAR keys:', Object.keys(asarMain).slice(0, 10).join(','));
  console.log('[format] ASAR app:', typeof asarMain.app);
} catch (e) {
  console.log('[format] ASAR error:', e.name, e.message.substring(0, 200));
}

process.exit(0);
