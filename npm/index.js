#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

const llamaCliPath = path.join(__dirname, 'bin', 'llama-cli');

const args = process.argv.slice(2);
const promptFlagIndex = args.indexOf('--prompt');

if (args.length < 1 || (args[0].startsWith('-') && promptFlagIndex === -1)) {
  console.error('❌ Usage: explain-me <file_path> [--prompt "custom prompt"]');
  process.exit(1);
}

const filePath = args[0];
const customPrompt = promptFlagIndex !== -1 ? args[promptFlagIndex + 1] : '';

if (!fs.existsSync(filePath)) {
  console.error(`❌ File not found: ${filePath}`);
  process.exit(1);
}

const code = fs.readFileSync(filePath, 'utf-8');

const prompt = customPrompt
  ? `[INST] ${customPrompt}\n\n${code}\n\n[/INST]`
  : `[INST] Explain what this code does:\n\n${code}\n\n[/INST]`;

if (!process.env.MODEL_PATH) {
  console.error('❌ MODEL_PATH environment variable is not set');
  process.exit(1);
}

const result = spawnSync(llamaCliPath, [
  '-m', process.env.MODEL_PATH,
  '-p', prompt,
  '-ngl', '1',
  '-n', '512'
], {
  encoding: 'utf-8'
});

if (result.error) {
  console.error('❌ Failed to start llama-cli:', result.error);
  process.exit(1);
}

if (result.status !== 0) {
  console.error('❌ llama-cli exited with error:');
  console.error(result.stderr);
  process.exit(result.status);
}

const output = result.stdout;
const match = output.match(/\[\/INST\](.*?)> EOF by user/s);

if (match && match[1]) {
  let respuesta = match[1].trim();
  respuesta = respuesta.replace(/^\[\/INST\]/, '').trim();
  respuesta = respuesta.replace(/> EOF by user$/, '').trim();
  console.log(respuesta);
} else {
  console.log(output);
}
