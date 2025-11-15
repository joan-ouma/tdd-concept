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

func TestPostCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "test=data" {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	cmd := postCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{server.URL, "-d", "test=data"})

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

func TestPostCommandWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "OK")
	}))
	defer server.Close()

	cmd := postCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{server.URL, "-c", "application/json", "-d", `{"key":"value"}`})

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

func TestPostCommandOutputToFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Post response")
	}))
	defer server.Close()

	tempFile, err := os.CreateTemp("", "test_post_output")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	cmd := postCmd
	cmd.SetArgs([]string{server.URL, "-d", "data=test", "-o", tempFile.Name()})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Error reading temp file: %v", err)
	}

	if string(content) != "Post response\n" {
		t.Errorf("Expected 'Post response\\n', got '%s'", string(content))
	}
}

func TestPostCommandInvalidURL(t *testing.T) {
	cmd := postCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{"invalid://url", "-d", "test=data"})

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
