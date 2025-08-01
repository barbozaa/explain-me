package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test buildPrompt with different flags
func TestBuildPrompt(t *testing.T) {
	code := "func main() {}"

	// Custom prompt overrides everything
	custom := "Custom prompt test"
	got := buildPrompt(code, custom, false, false)
	want := "[INST] Custom prompt test\n\nfunc main() {}\n\n[/INST]"
	if got != want {
		t.Errorf("Custom prompt failed.\nGot: %s\nWant: %s", got, want)
	}

	// Bug check mode prompt
	got = buildPrompt(code, "", false, true)
	if !strings.Contains(got, "Analyze this code for bugs") {
		t.Errorf("Bug check prompt missing expected text. Got: %s", got)
	}

	// Summary mode prompt
	got = buildPrompt(code, "", true, false)
	if !strings.Contains(got, "Summarize this code in English") {
		t.Errorf("Summary prompt missing expected text. Got: %s", got)
	}

	// Default prompt (no flags)
	got = buildPrompt(code, "", false, false)
	if !strings.Contains(got, "Explain what this code does") {
		t.Errorf("Default prompt missing expected text. Got: %s", got)
	}
}

// Test parseResponse with valid and invalid inputs
func TestParseResponse(t *testing.T) {
	validResp := "some output [/INST] This is the answer > EOF by user"
	answer, err := parseResponse(validResp)
	if err != nil {
		t.Errorf("parseResponse returned error on valid input: %v", err)
	}
	if answer != "This is the answer" {
		t.Errorf("parseResponse returned wrong answer: %s", answer)
	}

	invalidResp := "no end tag here"
	_, err = parseResponse(invalidResp)
	if err == nil {
		t.Errorf("parseResponse did not return error on invalid input")
	}
}

// Test listFilesInDir creates temp files and verifies listing
func TestListFilesInDir(t *testing.T) {
	dir := t.TempDir()

	// Create some files and a subdir
	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")
	subdir := filepath.Join(dir, "subdir")

	os.WriteFile(file1, []byte("hello"), 0644)
	os.WriteFile(file2, []byte("world"), 0644)
	os.Mkdir(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "file3.txt"), []byte("!"), 0644)

	files, err := listFilesInDir(dir)
	if err != nil {
		t.Fatalf("listFilesInDir error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
	// Check returned files contain file1 and file2
	found1 := false
	found2 := false
	for _, f := range files {
		if f == file1 {
			found1 = true
		} else if f == file2 {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Errorf("Returned files do not match expected files")
	}
}

// Test sha256sum returns expected checksum or error string
func TestSha256sum(t *testing.T) {
	tmpfile := t.TempDir() + "/test.txt"
	content := []byte("test content")
	err := os.WriteFile(tmpfile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Compute expected hash
	want := fmt.Sprintf("%x", sha256.Sum256(content))
	got := sha256sum(tmpfile)
	if got != want {
		t.Errorf("sha256sum returned %s, want %s", got, want)
	}

	// Non-existent file returns error string
	got = sha256sum("/non/existent/file")
	if !strings.HasPrefix(got, "error reading") {
		t.Errorf("sha256sum expected error prefix, got: %s", got)
	}
}
