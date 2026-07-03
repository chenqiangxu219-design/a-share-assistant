import { Module } from 'module';
const cache = Module._cache;
const keys = Object.keys(cache);
console.log('Total cached modules:', keys.length);
// Show all cached module paths
keys.forEach(k => {
  if (k.includes('electron') || k.includes('default_app') || k.includes('Resources')) {
    console.log(k);
  }
});
// Also show the first few
keys.slice(0, 10).forEach(k => console.log(k));
process.exit(0);
