// This simulates our compiled main.js
const electron_1 = require("electron");
console.log("require('electron') type:", typeof electron_1);
console.log("require('electron') value:", typeof electron_1 === 'string' ? electron_1 : Object.keys(electron_1 || {}).join(','));

// Try electron/main
const em = require("electron/main");
console.log("require('electron/main') type:", typeof em);
if (em && typeof em !== 'string') {
  console.log("electron/main keys:", Object.getOwnPropertyNames(em).join(','));
}
