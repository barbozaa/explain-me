package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
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
		_ = os.Chmod(exePath, 0777)
		_ = exec.Command("xattr", "-d", "com.apple.quarantine", exePath).Run()
	}

	return exePath, nil
}

func buildPrompt(code string, customPrompt string, summaryMode bool, bugCheckMode bool) string {
	if customPrompt != "" {
		return fmt.Sprintf("[INST] %s\n\n%s\n\n[/INST]", customPrompt, code)
	}
	if bugCheckMode {
		return fmt.Sprintf("[INST] Analyze this code for bugs, vulnerabilities, or bad practices. Explain any issues found:\n\n%s\n\n[/INST]", code)
	}
	if summaryMode {
		return fmt.Sprintf("[INST] Summarize this code in English, explaining its purpose and main functions:\n\n%s\n\n[/INST]", code)
	}
	return fmt.Sprintf("[INST] Explain what this code does:\n\n%s\n\n[/INST]", code)
}

func parseResponse(output string) (string, error) {
	if !strings.Contains(output, "[/INST]") {
		return "", errors.New("no [/INST] found in output")
	}
	parts := strings.Split(output, "[/INST]")
	lastPart := strings.TrimSpace(parts[len(parts)-1])

	cutOffMarkers := []string{"> EOF", "<|endoftext|>"}
	for _, marker := range cutOffMarkers {
		if strings.Contains(lastPart, marker) {
			lastPart = strings.Split(lastPart, marker)[0]
		}
	}
	clean := strings.TrimSpace(lastPart)
	if clean == "" {
		return "", errors.New("could not parse useful response from model output")
	}
	return clean, nil
}

func runLlamaCli(llamaCLI string, modelPath string, prompt string) (string, error) {
	cmd := exec.Command(llamaCLI, "-m", modelPath, "-p", prompt, "-ngl", "1", "-n", "512")
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

func analyzeFile(filePath, llamaCLI, modelPath, customPrompt string, summaryMode, bugCheckMode bool) error {
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
	fmt.Printf("📄 %s\n", answer)
	return nil
}

func startChatMode(llamaCLI, modelPath string) error {
	fmt.Println("🔵 Entering chat mode. Type 'exit' to quit.")

	cmd := exec.Command(llamaCLI,
		"-m", modelPath,
		"--interactive")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	go io.Copy(os.Stdout, stdout)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Question > ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		if text == "exit" {
			break
		}
		fmt.Fprintln(stdin, text)
	}
	cmd.Process.Kill()
	return nil
}

func main() {
	modelPath := os.Getenv("MODEL_PATH")
	if modelPath == "" {
		log.Fatal("❌ MODEL_PATH environment variable is not set")
	}

	fileFlag := flag.String("f", "", "Path to a single file to analyze")
	dirFlag := flag.String("d", "", "Path to a directory to analyze all files")
	promptFlag := flag.String("prompt", "", "Custom prompt to use")
	summaryFlag := flag.Bool("summary", false, "If set, use default summary prompt in English")
	bugCheckFlag := flag.Bool("bug-check", false, "If set, analyze the code for bugs and bad practices")
	chatModeFlag := flag.Bool("chat-mode", false, "If set, starts the model in chat mode")

	flag.Parse()

	llamaCLI, err := extractLlamaCli()
	if err != nil {
		log.Fatalf("❌ Failed to extract llama-cli: %v", err)
	}
	defer os.RemoveAll(filepath.Dir(llamaCLI))

	// Handle chat mode
	if *chatModeFlag {
		if err := startChatMode(llamaCLI, modelPath); err != nil {
			log.Fatalf("❌ Chat mode failed: %v", err)
		}
		return
	}

	// If neither file nor directory provided
	if *fileFlag == "" && *dirFlag == "" {
		fmt.Println("❌ Usage: explain-me -f <file_path> OR -d <directory_path> [--prompt \"your prompt\"] [--summary] [--bug-check] [--chat-mode]")
		os.Exit(1)
	}

	if *fileFlag != "" {
		err := analyzeFile(*fileFlag, llamaCLI, modelPath, *promptFlag, *summaryFlag, *bugCheckFlag)
		if err != nil {
			log.Fatal(err)
		}
	} else {
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
