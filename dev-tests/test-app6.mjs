// Try default import only
import electron from 'electron/main';
console.log('default:', typeof electron, electron !== null && Object.getOwnPropertyNames(electron).join(','));

// Try namespace
import * as e from 'electron/main';
console.log('NS default:', typeof e.default, e.default && Object.getOwnPropertyNames(e.default).join(','));
console.log('NS module.exports:', typeof e.module_exports, e['module.exports'] && Object.getOwnPropertyNames(e['module.exports']).join(','));

// Check if it's a getter-based lazy object
const ns = e;
for (const key of ['app', 'ipcMain', 'BrowserWindow', 'screen', 'protocol', 'session', 'dialog']) {
  const desc = Object.getOwnPropertyDescriptor(ns, key);
  console.log(key, ':', desc ? ('get' in desc ? 'getter' : 'value:' + typeof desc.value) : 'undefined');
}
process.exit(0);
