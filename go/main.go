package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed assets/bin/mac-arm64/llama-cli
var llamaCliBytes []byte

// compute the sha256 checksum of a file (used to verify llama-cli binary)
func sha256sum(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "error reading: " + err.Error()
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

// extract embedded llama-cli binary to ~/.local/bin if not already there or outdated
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

	// Only write the file if it has changed
	if existingHash != newHash {
		err = os.WriteFile(exePath, llamaCliBytes, 0755)
		if err != nil {
			return "", err
		}
		_ = os.Chmod(exePath, 0777)
		_ = exec.Command("xattr", "-d", "com.apple.quarantine", exePath).Run()
	}

	return exePath, nil
}

// build prompt for the model based on flags
func buildPrompt(code string, customPrompt string, summaryMode bool, bugCheckMode bool) string {
	if customPrompt != "" {
		// if custom prompt is provided, use it verbatim
		return fmt.Sprintf("[INST] %s\n\n%s\n\n[/INST]", customPrompt, code)
	}
	if bugCheckMode {
		// default bug check prompt in English if --bug-check is set and no custom prompt
		return fmt.Sprintf("[INST] Analyze this code for bugs, vulnerabilities, or bad practices. Explain any issues found:\n\n%s\n\n[/INST]", code)
	}
	if summaryMode {
		// default summary prompt in English if --summary is set and no custom prompt
		return fmt.Sprintf("[INST] Summarize this code in English, explaining its purpose and main functions:\n\n%s\n\n[/INST]", code)
	}
	// default explanation prompt in English if no prompt or special mode
	return fmt.Sprintf("[INST] Explain what this code does:\n\n%s\n\n[/INST]", code)
}

// parse the model's response from the raw output
func parseResponse(response string) (string, error) {
	parts := strings.SplitN(response, "[/INST]", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("could not find [/INST] in output")
	}
	answer := strings.TrimSpace(parts[1])
	answer = strings.TrimSuffix(answer, "> EOF by user")
	answer = strings.ReplaceAll(answer, "[/INST]", "")
	return strings.TrimSpace(answer), nil
}

// run llama-cli with the given model and prompt, capturing stdout and stderr
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

// list all non-directory files in the given directory (non-recursive)
func listFilesInDir(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

// analyze a single file: read, build prompt, run model, parse and print output
func analyzeFile(filePath string, llamaCLI string, modelPath string, customPrompt string, summaryMode bool, bugCheckMode bool) error {
	codeBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("❌ Error reading file %s: %v", filePath, err)
	}
	code := string(codeBytes)
	prompt := buildPrompt(code, customPrompt, summaryMode, bugCheckMode)

	response, err := runLlamaCli(llamaCLI, modelPath, prompt)
	if err != nil {
		return err
	}
	answer, err := parseResponse(response)
	if err != nil {
		fmt.Printf("❌ Failed to parse model response for %s\nRaw output:\n%s\n", filePath, response)
		return err
	}

	fmt.Printf("📄 %s\n\n%s\n\n---\n", filePath, answer)
	return nil
}

func main() {
	// get model path from environment variable
	modelPath := os.Getenv("MODEL_PATH")
	if modelPath == "" {
		log.Fatal("❌ MODEL_PATH environment variable is not set")
	}

	// CLI flags: -f file, -d dir, --prompt custom prompt, --summary for summary mode, --bug-check for bug detection mode
	fileFlag := flag.String("f", "", "Path to a single file to analyze")
	dirFlag := flag.String("d", "", "Path to a directory to analyze all files")
	promptFlag := flag.String("prompt", "", "Custom prompt to use")
	summaryFlag := flag.Bool("summary", false, "If set, use default summary prompt in English")
	bugCheckFlag := flag.Bool("bug-check", false, "If set, analyze the code for bugs and bad practices")

	flag.Parse()

	if *fileFlag == "" && *dirFlag == "" {
		fmt.Println("❌ Usage: explain-me -f <file_path> OR -d <directory_path> [--prompt \"your prompt\"] [--summary] [--bug-check]")
		os.Exit(1)
	}

	// extract the llama-cli binary
	llamaCLI, err := extractLlamaCli()
	if err != nil {
		log.Fatalf("❌ Failed to extract llama-cli: %v", err)
	}
	defer os.RemoveAll(filepath.Dir(llamaCLI)) // optional cleanup

	if *fileFlag != "" {
		// analyze single file
		err := analyzeFile(*fileFlag, llamaCLI, modelPath, *promptFlag, *summaryFlag, *bugCheckFlag)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// analyze all files in directory (non-recursive)
		files, err := listFilesInDir(*dirFlag)
		if err != nil {
			log.Fatalf("❌ Failed to read directory: %v", err)
		}
		if len(files) == 0 {
			fmt.Println("⚠️ No files found in the directory")
			return
		}
		for _, file := range files {
			fmt.Printf("🔍 Analyzing %s...\n", file)
			err := analyzeFile(file, llamaCLI, modelPath, *promptFlag, *summaryFlag, *bugCheckFlag)
			if err != nil {
				fmt.Printf("❌ Error analyzing %s: %v\n", file, err)
			}
		}
	}
}
