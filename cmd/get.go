package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [URL]",
	Short: "Perform HTTP GET request",
	Args:  cobra.ExactArgs(1),
	Run:   runGet,
}

var (
	outputFile string
	timeout    int
	headers    []string
)

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write output to file")
	getCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Request timeout in seconds")
	getCmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "HTTP headers")
}

func runGet(cmd *cobra.Command, args []string) {
	url := args[0]
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	for _, header := range headers {
		req.Header.Set(parseHeader(header))
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if outputFile != "" {
		err = os.WriteFile(outputFile, body, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Response written to %s\n", outputFile)
	} else {
		fmt.Println(string(body))
	}
}

func parseHeader(header string) (string, string) {
	for i, ch := range header {
		if ch == ':' {
			return header[:i], header[i+1:]
		}
	}
	return "", ""
}
