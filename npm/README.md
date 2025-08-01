# explain-me

> 🧠🚀 A local CLI tool to explain source code using LLaMA models powered by [llama.cpp](https://github.com/ggerganov/llama.cpp).

---

## What is explain-me?

`explain-me` is a developer-friendly command-line tool that helps you **understand and explain source code** by running lightweight, open-source LLMs (Large Language Models) entirely **on your own machine**. No cloud, no data leaks — just fast, private, and powerful code explanations right from your terminal.

---

## Why use explain-me?

- **Privacy first:** All code and model processing happen locally. No data is sent anywhere.
- **Lightweight:** Uses efficient LLaMA-compatible GGUF models via the blazing-fast `llama-cli`.
- **Customizable:** Supply your own prompt to tailor explanations to your needs.
- **Zero dependencies:** Bundled with a precompiled `llama-cli` binary for macOS ARM64 (Apple Silicon).
- **Multi-language:** Works on any source code — Go, Python, JavaScript, TypeScript, JSON, and more.
- **Easy to use:** Simple CLI interface designed for developers, powered by Go.
- **Open source:** Contributions and improvements are welcome!

---

## Current Status

- ✅ Supports macOS ARM64 (Apple Silicon) with embedded `llama-cli` binary.
- 🚧 Linux and Windows support **coming soon** (requires cross-compiling `llama-cli`).
- Requires a local LLaMA-compatible model in GGUF format (see below).

---

## Prerequisites

Before running `explain-me`, you need:

- A **LLaMA-compatible GGUF model** (such as [Mistral 7B GGUF](https://huggingface.co/TheBloke/phi-2-GGUF/tree/main))
- Set the environment variable `MODEL_PATH` to the absolute path of your model file:

```bash
export MODEL_PATH=/path/to/your/model.gguf

## Usage

# Basic usage: explain a source code file
npx explain-me -f ./path/to/file.go

# Use a custom prompt to tailor the explanation
npx explain-me -f ./path/to/file.py --prompt "Summarize this function briefly:"

# Analyze all files in a directory (non-recursive)
npx explain-me -d ./path/to/project

# Generate a concise summary explanation (default prompt in English)
npx explain-me -f ./path/to/file.js --summary

# Combine summary mode with custom prompt
npx explain-me -f ./path/to/file.js --summary --prompt "Focus on security implications:"

# Show help and usage info
npx explain-me --help
```
---

## Features

**Summary Mode** (**--summary**)
When the **--summary** flag is provided, explain-me will ask the model to summarize the code in English, highlighting its purpose and main functions instead of a detailed explanation.

This mode can be combined with a custom prompt (**--prompt**) for even more tailored summaries.

**Directory Analysis** (**-d <directory_path>**)
Instead of analyzing a single file, use the -d flag to analyze all files (non-recursive) inside a directory. The tool will process each file separately and print individual explanations.

Useful for quick codebase overviews or bulk analysis.

**Bug Check Mode** (**--bug-check**)
A specialized mode that instructs the model to analyze the code for potential bugs, security issues, or logic errors.

Example:
```bash
npx explain-me -f ./src/main.go --bug-check
```
This mode is ideal for quick static analysis powered by LLMs to detect subtle problems.

## What happens if you run without arguments?

If you run npx explain-me without specifying -f or -d, the CLI will:

- Display an error message explaining that a file or directory path is required.

- Show usage instructions and available options.

- Exit without running the model.

```bash
$ npx explain-me

❌ Usage: explain-me -f <file_path> OR -d <directory_path> [--prompt "your prompt"] [--summary] [--bug-check]

Usage: explain-me [options]

Options:
  -f string
        Path to a single file to analyze
  -d string
        Path to a directory to analyze all files
  -prompt string
        Custom prompt to use
  -summary
        Use default summary prompt in English
  -bug-check
        Analyze code for bugs and potential issues
  -h, --help
        Show help message
```
