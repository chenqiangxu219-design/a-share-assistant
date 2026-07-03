import * as electron from 'electron/main'
import { Module } from 'node:module'
import { fileURLToPath } from 'node:url'
import * as path from 'node:path'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// Get the npm package index.js path
const indexPath = path.resolve(__dirname, '..', 'node_modules', 'electron', 'index.js')

// Cache the bindings so require('electron') from CJS code finds them
Module._cache[indexPath] = { exports: electron }

console.log('[bridge] Cached electron bindings:', Object.keys(electron).slice(0, 8).join(','))
console.log('[bridge] app type:', typeof electron.app)

export default electron
