package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "basic get",
			args:    []string{server.URL},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := getCmd
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("getCmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			out, err := io.ReadAll(b)
			if err != nil {
				t.Fatalf("Error reading output: %v", err)
			}

			if string(out) != "OK\n" {
				t.Errorf("Expected 'OK\\n', got '%s'", string(out))
			}
		})
	}
}

func TestGetCommandWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "value" {
			http.Error(w, "Header not found", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	cmd := getCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{server.URL, "-H", "X-Test: value"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatalf("Error reading output: %v", err)
	}

	if string(out) != "OK\n" {
		t.Errorf("Expected 'OK\\n', got '%s'", string(out))
	}
}

func TestGetCommandOutputToFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "File content")
	}))
	defer server.Close()

	tempFile, err := os.CreateTemp("", "test_output")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	cmd := getCmd
	cmd.SetArgs([]string{server.URL, "-o", tempFile.Name()})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Error reading temp file: %v", err)
	}

	if string(content) != "File content\n" {
		t.Errorf("Expected 'File content\\n', got '%s'", string(content))
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		input string
		key   string
		value string
	}{
		{"Key:Value", "Key", "Value"},
		{"Key: Value", "Key", " Value"},
		{"Key", "", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, value := parseHeader(tt.input)
			if key != tt.key || value != tt.value {
				t.Errorf("parseHeader(%q) = (%q, %q), want (%q, %q)", tt.input, key, value, tt.key, tt.value)
			}
		})
	}
}

func TestGetCommandInvalidURL(t *testing.T) {
	cmd := getCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{"invalid://url"})

	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	var exitCode int
	osExit = func(code int) {
		exitCode = code
		panic(fmt.Sprintf("exit %d", code))
	}

	defer func() {
		if r := recover(); r != nil {
			if exitCode != 1 {
				t.Errorf("Expected exit code 1, got %d", exitCode)
			}
		}
	}()

	cmd.Execute()
	t.Error("Expected os.Exit(1) to be called")
}
