package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	code := "func add(a int, b int) int { return a + b }"

	wantDefault := "[INST] Explain what this code does:\n\nfunc add(a int, b int) int { return a + b }\n\n[/INST]"
	got := buildPrompt(code, "")
	if got != wantDefault {
		t.Errorf("buildPrompt() = %q, want %q", got, wantDefault)
	}

	custom := "Describe this Go function:"
	wantCustom := "[INST] Describe this Go function:\n\nfunc add(a int, b int) int { return a + b }\n\n[/INST]"
	got = buildPrompt(code, custom)
	if got != wantCustom {
		t.Errorf("buildPrompt(custom) = %q, want %q", got, wantCustom)
	}
}

func TestParseResponse(t *testing.T) {
	response := "[INST] Some prompt [/INST] This is the answer > EOF by user"
	want := "This is the answer"
	got, err := parseResponse(response)
	if err != nil {
		t.Fatalf("parseResponse() error = %v", err)
	}
	if got != want {
		t.Errorf("parseResponse() = %q, want %q", got, want)
	}

	badResponse := "no closing tag"
	_, err = parseResponse(badResponse)
	if err == nil {
		t.Errorf("parseResponse() expected error for bad input")
	}
}

func TestExtractLlamaCli(t *testing.T) {
	// Esta prueba verifica que extractLlamaCli cree un archivo ejecutable temporal
	exePath, err := extractLlamaCli()
	if err != nil {
		t.Fatalf("extractLlamaCli() error = %v", err)
	}
	defer os.RemoveAll(filepath.Dir(exePath))

	info, err := os.Stat(exePath)
	if err != nil {
		t.Fatalf("extracted file does not exist: %v", err)
	}

	if info.IsDir() {
		t.Errorf("expected file but got directory")
	}

	mode := info.Mode()
	if mode&0100 == 0 {
		t.Errorf("expected executable permission, mode = %v", mode)
	}
}
