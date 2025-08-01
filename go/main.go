package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed assets/bin/mac-arm64/llama-cli
var llamaCliBytes []byte

func extractLlamaCli() (string, error) {
	tmpDir, err := ioutil.TempDir("", "llama-cli")
	if err != nil {
		return "", err
	}
	exePath := filepath.Join(tmpDir, "llama-cli")

	err = ioutil.WriteFile(exePath, llamaCliBytes, 0755)
	if err != nil {
		return "", err
	}

	return exePath, nil
}

func buildPrompt(code string, customPrompt string) string {
	if customPrompt != "" {
		return fmt.Sprintf("[INST] %s\n\n%s\n\n[/INST]", customPrompt, code)
	}
	return fmt.Sprintf("[INST] Explain what this code does:\n\n%s\n\n[/INST]", code)
}

func parseResponse(response string) (string, error) {
	parts := strings.SplitN(response, "[/INST]", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("could not find [/INST] in output")
	}

	answer := strings.TrimSpace(parts[1])
	answer = strings.TrimSuffix(answer, "> EOF by user")
	answer = strings.ReplaceAll(answer, "[/INST]", "")
	answer = strings.TrimSpace(answer)

	return answer, nil
}

func runLlamaCli(llamaCLI string, modelPath string, prompt string) (string, error) {
	cmd := exec.Command(llamaCLI,
		"-m", modelPath,
		"-p", prompt,
		"-ngl", "1",
		"-n", "512")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running llama-cli: %v\n%s", err, stderr.String())
	}

	return out.String(), nil
}

func main() {
	modelPath := os.Getenv("MODEL_PATH")
	if modelPath == "" {
		log.Fatal("❌ MODEL_PATH environment variable is not set")
	}

	customPrompt := flag.String("prompt", "", "Custom prompt for the model")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <file_path> [options]\n\n", filepath.Base(os.Args[0]))
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("❌ Usage: explain-me <file_path> [--prompt \"custom prompt\"]")
		flag.Usage()
		os.Exit(1)
	}
	path := args[0]

	codeBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("❌ Error reading file: %s\n", err)
		os.Exit(1)
	}
	code := string(codeBytes)

	prompt := buildPrompt(code, *customPrompt)

	llamaCLI, err := extractLlamaCli()
	if err != nil {
		log.Fatalf("Failed to extract llama-cli binary: %v", err)
	}
	defer os.RemoveAll(filepath.Dir(llamaCLI))

	response, err := runLlamaCli(llamaCLI, modelPath, prompt)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	answer, err := parseResponse(response)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println(response)
		os.Exit(1)
	}

	fmt.Println(answer)
}
