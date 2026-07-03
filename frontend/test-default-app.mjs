// Simulate what the default app does
import * as electron from 'electron/main';

// The namespace might use Symbol.toStringTag or similar tricks
console.log('toStringTag:', electron[Symbol.toStringTag]);
console.log('toString:', Object.prototype.toString.call(electron));

// Check if properties are accessible via bracket notation
for (const key of ['app', 'ipcMain', 'BrowserWindow', 'screen', 'protocol', 'session', 'dialog', 'contentTracing', 'globalShortcut', 'nativeTheme', 'powerSaveBlocker', 'systemPreferences', 'net', 'netLog', 'ipcRenderer', 'webContents']) {
  const val = electron[key];
  if (val !== undefined) {
    console.log(key, ':', typeof val);
  }
}

// Check valueOf
try {
  const v = electron.valueOf();
  console.log('valueOf:', v === electron ? 'same' : typeof v);
} catch(e) {
  console.log('valueOf error:', e.message);
}

process.exit(0);
