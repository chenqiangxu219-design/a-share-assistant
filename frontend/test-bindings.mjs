const names = [
  'blink_bindings', 'bindings', 'electron', 'electron/browser', 'electron/common',
  'electron/main', 'electron/renderer', 'native_clib', 'node', 'options',
  'v8', 'vendor/common/bindings', 'vendor/electron', 'async_wrap',
  'config_source', 'connection_wrap', 'data_queue', 'encoding',
  'fs', 'http_parser', 'lib', 'module_wrap', 'pipe_wrap',
  'process', 'spawn_sync', 'stream_base', 'stream_wrap', 'string_bytes',
  'sys_wrap', 'tcp_wrap', 'timers', 'trace_events', 'tty_wrap',
  'udp_wrap', 'uv', 'variables_wrap', 'worker',
];

for (const name of names) {
  try {
    const b = process._linkedBinding(name);
    console.log(name, '-> OK, keys:', Object.getOwnPropertyNames(b).slice(0, 3).join(','));
  } catch(e) {
    // silently skip
  }
}
process.exit(0);
