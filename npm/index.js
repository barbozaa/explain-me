#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');

const goBinaryPath = path.join(__dirname, './bin/explain-me');

const args = process.argv.slice(2);

const child = spawn(goBinaryPath, args, { stdio: 'inherit' });

child.on('error', (err) => {
  console.error('Failed to start subprocess:', err);
  process.exit(1);
});

child.on('exit', (code) => {
  process.exit(code);
});
