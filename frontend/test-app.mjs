import electron from 'electron/main';
console.log('import electron type:', typeof electron);
console.log('electron:', electron);

// Also try namespace import
import * as electronNS from 'electron/main';
console.log('NS keys:', Object.getOwnPropertyNames(electronNS).join(','));
process.exit(0);
