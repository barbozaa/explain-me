#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const readline = require('readline');
const { spawnSync, spawn } = require('child_process');

const llamaCliPath = path.join(__dirname, 'bin', 'llama-cli');

const args = process.argv.slice(2);
let filePath = '';
let dirPath = '';
let customPrompt = '';
let summaryMode = false;
let bugCheckMode = false;
let chatMode = false;

for (let i = 0; i < args.length; i++) {
  const arg = args[i];
  switch (arg) {
    case '-f':
      filePath = args[i + 1];
      i++;
      break;
    case '-d':
      dirPath = args[i + 1];
      i++;
      break;
    case '--prompt':
      customPrompt = args[i + 1];
      i++;
      break;
    case '--summary':
      summaryMode = true;
      break;
    case '--bug-check':
      bugCheckMode = true;
      break;
    case '--chat-mode':
      chatMode = true;
      break;
    default:
      break;
  }
}

if (!process.env.MODEL_PATH) {
  console.error('❌ MODEL_PATH environment variable is not set');
  process.exit(1);
}

if (chatMode) {
  console.log('🔵 Entering chat mode. Type "exit" to quit.\n');
  const chat = spawn(llamaCliPath, [
    '-m', process.env.MODEL_PATH,
    '--interactive'
  ], { stdio: ['pipe', 'pipe', process.stderr] });

  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
  });

  chat.stdout.on('data', (data) => {
    process.stdout.write(data.toString());
  });

  rl.on('line', (line) => {
    if (line.trim().toLowerCase() === 'exit') {
      chat.kill();
      rl.close();
    } else {
      chat.stdin.write(`${line.trim()}\n`);
    }
  });

  return;
}

function buildPrompt(code, prompt, summary, bugCheck) {
  if (prompt) {
    return `[INST] ${prompt}\n\n${code}\n\n[/INST]`;
  } else if (bugCheck) {
    return `[INST] Analyze this code for bugs, vulnerabilities, or bad practices. Explain any issues found:\n\n${code}\n\n[/INST]`;
  } else if (summary) {
    return `[INST] Summarize this code in English, explaining its purpose and main functions:\n\n${code}\n\n[/INST]`;
  } else {
    return `[INST] Explain what this code does:\n\n${code}\n\n[/INST]`;
  }
}

function parseResponse(output) {
  if (!output.includes('[/INST]')) return null;

  const parts = output.split('[/INST]');
  let last = parts[parts.length - 1].trim();

  for (const marker of ['> EOF', '<|endoftext|>']) {
    if (last.includes(marker)) {
      last = last.split(marker)[0];
    }
  }

  return last.trim() || null;
}

function analyzeCode(code, fileName) {
  const prompt = buildPrompt(code, customPrompt, summaryMode, bugCheckMode);

  const result = spawnSync(llamaCliPath, [
    '-m', process.env.MODEL_PATH,
    '-p', prompt,
    '-ngl', '1',
    '-n', '512'
  ], { encoding: 'utf-8' });

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
  const parsed = parseResponse(output);

  console.log(`\n📄 ${fileName}`);
  console.log(parsed || output);
}

if (filePath) {
  if (!fs.existsSync(filePath)) {
    console.error(`❌ File not found: ${filePath}`);
    process.exit(1);
  }
  const code = fs.readFileSync(filePath, 'utf-8');
  analyzeCode(code, path.basename(filePath));
} else if (dirPath) {
  if (!fs.existsSync(dirPath)) {
    console.error(`❌ Directory not found: ${dirPath}`);
    process.exit(1);
  }

  const files = fs.readdirSync(dirPath).filter(f => {
    const fullPath = path.join(dirPath, f);
    return fs.statSync(fullPath).isFile();
  });

  if (files.length === 0) {
    console.error('❌ No code files found in directory:', dirPath);
    process.exit(1);
  }

  for (const file of files) {
    const fullPath = path.join(dirPath, file);
    try {
      const code = fs.readFileSync(fullPath, 'utf-8');
      analyzeCode(code, file);
    } catch (e) {
      console.warn(`⚠️ Skipping unreadable file ${file}: ${e.message}`);
    }
  }
}
