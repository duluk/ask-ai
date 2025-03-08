#!/usr/bin/env node

/**
 * TUI entry point with diagnostic options
 */

// Use spawnSync to run the TUI entry point with the flag to show top-level await sources
const { spawnSync } = require('child_process');

const result = spawnSync('node', [
  '--experimental-print-required-tla',
  './dist/src/tui-entry.js'
], {
  stdio: 'inherit',
  env: process.env
});

// Propagate the exit code
process.exit(result.status || 0);