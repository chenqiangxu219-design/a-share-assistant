console.log('_linkedBinding type:', typeof process._linkedBinding);
try {
  const b = process._linkedBinding('node');
  console.log('node binding:', typeof b, Object.getOwnPropertyNames(b).slice(0, 5).join(','));
} catch(e) {
  console.log('node binding error:', e.message);
}
try {
  const b = process._linkedBinding('options');
  console.log('options binding:', typeof b, b && Object.getOwnPropertyNames(b).slice(0, 5).join(','));
} catch(e) {
  console.log('options binding error:', e.message);
}
// Try the process binding which should always exist
try {
  const b = process._linkedBinding('internal/v8');
  console.log('internal/v8:', typeof b);
} catch(e) {
  console.log('internal/v8 error:', e.message);
}
process.exit(0);
