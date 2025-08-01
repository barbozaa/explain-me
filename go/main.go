package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed assets/bin/mac-arm64/llama-cli
var llamaCliBytes []byte

func sha256sum(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "error reading: " + err.Error()
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

func extractLlamaCli() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	binDir := filepath.Join(homeDir, ".local", "bin")
	err = os.MkdirAll(binDir, 0755)
	if err != nil {
		return "", err
	}

	exePath := filepath.Join(binDir, "llama-cli")
	existingHash := sha256sum(exePath)
	newHash := fmt.Sprintf("%x", sha256.Sum256(llamaCliBytes))
	if existingHash != newHash {
		err = os.WriteFile(exePath, llamaCliBytes, 0755)
		if err != nil {
			return "", err
		}
		os.Chmod(exePath, 0777)
		exec.Command("xattr", "-d", "com.apple.quarantine", exePath).Run()
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
		return "", fmt.Errorf("error running llama-cli: %v\nSTDERR: %s\nSTDOUT: %s", err, stderr.String(), out.String())
	}

	return out.String(), nil
}

func main() {
	modelPath := os.Getenv("MODEL_PATH")
	if modelPath == "" {
		log.Fatal("❌ MODEL_PATH environment variable is not set")
	}

	var filePath string
	var customPrompt string

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--prompt" && i+1 < len(os.Args) {
			customPrompt = os.Args[i+1]
			i++
		} else if !strings.HasPrefix(arg, "--") && filePath == "" {
			filePath = arg
		}
	}

	if filePath == "" {
		fmt.Println("❌ Usage: explain-me <file_path> [--prompt \"custom prompt\"]")
		os.Exit(1)
	}

	codeBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ Error reading file: %s\n", err)
		os.Exit(1)
	}
	code := string(codeBytes)

	prompt := buildPrompt(code, customPrompt)

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
