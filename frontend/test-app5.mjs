// This is how Electron v42 expects you to import
import { app, BrowserWindow, screen, ipcMain } from 'electron/main';

console.log('app:', typeof app);
console.log('BrowserWindow:', typeof BrowserWindow);
console.log('screen:', typeof screen);
console.log('ipcMain:', typeof ipcMain);

if (typeof app === 'object' && app !== null) {
  console.log('app keys:', Object.getOwnPropertyNames(app).slice(0, 10).join(','));
}
process.exit(0);
