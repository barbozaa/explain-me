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

- A **LLaMA-compatible GGUF model** (such as [Mistral 7B GGUF](https://huggingface.co/TheBloke/mistral-7B-Instruct-GGUF))
- Set the environment variable `MODEL_PATH` to the absolute path of your model file:

```bash
export MODEL_PATH=/path/to/your/model.gguf
