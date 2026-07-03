// This runs before the default app loads. We can import electron/main here
// and cache it for CJS require() to find.
import electron from 'electron/main';
import { Module } from 'node:module';

// Cache it so require('electron') finds it
Module._cache['electron'] = { exports: electron };
Module._cache[require.resolve('electron')] = { exports: electron };
