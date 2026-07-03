// Test CJS require in v28
const { createRequire } = require('node:module');
const req = createRequire('file:///Users/chen/a-share-assistant/frontend/package.json');

console.log('=== require("electron") ===');
try {
  const elec = req('electron');
  console.log('type:', typeof elec);
  console.log('value:', typeof elec === 'string' ? elec : Object.keys(elec || {}).join(','));
} catch (e) {
  console.log('error:', e.message.substring(0, 150));
}

console.log('\n=== require("electron/main") ===');
try {
  const main = req('electron/main');
  console.log('type:', typeof main);
  if (typeof main === 'object' && main) {
    console.log('keys:', Object.keys(main).join(','));
  } else {
    console.log('value:', main);
  }
} catch (e) {
  console.log('error:', e.message.substring(0, 150));
}

console.log('\n=== require("electron").app check ===');
try {
  const app = req('app');
  console.log('app type:', typeof app);
} catch (e) {
  console.log('app error:', e.message.substring(0, 100));
}

process.exit(0);
