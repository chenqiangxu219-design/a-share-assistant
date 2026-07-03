const bindingNames = [
  'electron',
  'electron/browser',
  'electron/common',
  'electron/main',
  'electron/renderer',
  'vendor/common/bindings',
  'vendor/electron/bindings',
  'bindings',
];

for (const name of bindingNames) {
  try {
    const b = process._linkedBinding(name);
    const keys = b ? Object.getOwnPropertyNames(b).slice(0, 5).join(',') : '';
    console.log(name, '->', typeof b, keys);
  } catch(e) {
    console.log(name, '->', e.message.split('\n')[0]);
  }
}
process.exit(0);
