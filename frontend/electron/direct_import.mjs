// Test: What happens when Electron runs our script as the MAIN module?
// (not ELECTRON_RUN_AS_NODE, but as the actual entry point)

console.log('[mjs] process.versions.electron:', process.versions?.electron);
console.log('[mjs] process.defaultApp:', process.defaultApp);
console.log('[mjs] process.noAsar:', process.noAsar);

// Try import('electron/main')
console.log('[mjs] importing electron/main...');
try {
  const mod = await import('electron/main');
  console.log('[mjs] type:', typeof mod);
  console.log('[mjs] keys:', Object.keys(mod).join(','));
  console.log('[mjs] app:', typeof mod.app);
  console.log('[mjs] BrowserWindow:', typeof mod.BrowserWindow);
  console.log('[mjs] screen:', typeof mod.screen);
  console.log('[mjs] ipcMain:', typeof mod.ipcMain);

  // Try getOwnPropertyDescriptor for app
  const desc = Object.getOwnPropertyDescriptor(mod, 'app');
  console.log('[mjs] app descriptor:', desc);
} catch (e) {
  console.log('[mjs] error:', e.name, e.message.substring(0, 120));
}

// Also try bare 'electron'
console.log('\n[mjs] importing bare electron...');
try {
  const mod = await import('electron');
  console.log('[mjs] type:', typeof mod);
  console.log('[mjs] keys:', Object.keys(mod).join(','));
  if (mod.default) {
    console.log('[mjs] default:', typeof mod.default, typeof mod.default === 'string' ? mod.default : Object.keys(mod.default).join(','));
  }
} catch (e) {
  console.log('[mjs] error:', e.name, e.message.substring(0, 120));
}

// Try the individual imports
console.log('\n[mjs] individual imports...');
for (const name of ['electron/app', 'electron/screen', 'electron/ipc-main']) {
  try {
    const mod = await import(name);
    console.log('[mjs] ' + name + ':', Object.keys(mod).slice(0, 5).join(','));
  } catch (e) {
    console.log('[mjs] ' + name + ':', e.message.substring(0, 80));
  }
}

setTimeout(() => process.exit(0), 500);
