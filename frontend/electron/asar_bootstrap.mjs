// Bootstrap approach: use default_app.asar/main.js context to get real electron bindings
// Then import our actual main process code

import path from 'node:path';
import fs from 'node:fs';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const resourcesPath = process.resourcesPath;
const asarPath = path.join(resourcesPath, 'default_app.asar', 'main.js');

console.log('[bootstrap] ASAR main.js path:', asarPath);
console.log('[bootstrap] exists:', fs.existsSync(asarPath));

// Read main.js source from ASAR
process.noAsar = true;
const rawContent = fs.readFileSync(asarPath, 'utf-8');
process.noAsar = false;

// Parse ASAR header to extract main.js
const filesMatch = rawContent.match(/\{"files"\:(\{.*?\})\}/s);
const files = JSON.parse('{"files":' + filesMatch[1]).files;
const entry = files['main.js'];
const mainJs = rawContent.substring(entry.offset, entry.offset + entry.size);

console.log('[bootstrap] main.js length:', mainJs.length);
console.log('[bootstrap] first line:', mainJs.split('\n')[0]);

// Now evaluate main.js in a vm.SourceTextModule context
// We need to intercept its import of 'electron/main' and return real bindings
// The trick: main.js does "import * as electron from 'electron/main'"
// If we can make that work, we get all the real bindings

import vm from 'node:vm';

const ssm = new vm.SourceTextModule(mainJs, {
  identifier: 'default_app.asar/main.js',
  async importModuleDynamically(specifier, moduleRef) {
    console.log('[bootstrap] intercepted import:', specifier);
    if (specifier === 'electron/main') {
      // This is the key! We need to return the REAL electron bindings.
      // But we don't have them yet... we're trying to get them!
      // The chicken-and-egg problem.
      console.log('[bootstrap] electron/main requested - we need the real bindings');
    }
    // For node: modules, use the real loader
    if (specifier.startsWith('node:')) {
      return import(specifier);
    }
    throw new Error('Unknown module: ' + specifier);
  },
});

try {
  await ssm.evaluate();
  console.log('[bootstrap] namespace:', Object.keys(ssm.namespace).join(','));
} catch (e) {
  console.log('[bootstrap] error:', e.message.substring(0, 200));
}

setTimeout(() => process.exit(0), 500);
