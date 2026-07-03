// Check process resources
console.log('resourcesPath:', process.resourcesPath);

// Try to find the electron module in the asar
import fs from 'fs';
import path from 'path';

// Check if there's an electron.asar in resources
const resourcesDir = process.resourcesPath;
const files = fs.readdirSync(resourcesDir);
console.log('resources files:', files.filter(f => f.includes('electron') || f.includes('module')).join(', '));

// Try to import from the resources path
const electronAsar = path.join(resourcesDir, 'default_app.asar');
console.log('default_app.asar exists:', fs.existsSync(electronAsar));

// The default_app.asar has the electron module resolution. Let's check if there's a way to access it.
// Electron might use a custom ESM loader
console.log('process.features:', Object.getOwnPropertyNames(process).filter(k => k.includes('esm') || k.includes('hook') || k.includes('link')).join(','));

// Check if we can use _linkedBinding
try {
  const bindings = process._linkedBinding ? process._linkedBinding('electron/main') : null;
  console.log('linkedBinding electron/main:', bindings ? Object.getOwnPropertyNames(bindings).slice(0, 10).join(',') : 'N/A');
} catch(e) {
  console.log('linkedBinding error:', e.message);
}

process.exit(0);
