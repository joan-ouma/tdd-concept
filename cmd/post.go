package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var postCmd = &cobra.Command{
	Use:   "post [URL]",
	Short: "Perform HTTP POST request",
	Args:  cobra.ExactArgs(1),
	Run:   runPost,
}

var (
	postData    string
	contentType string
)

func init() {
	rootCmd.AddCommand(postCmd)
	postCmd.Flags().StringVarP(&postData, "data", "d", "", "Data to send in POST request")
	postCmd.Flags().StringVarP(&contentType, "content-type", "c", "application/x-www-form-urlencoded", "Content type")
}

func runPost(cmd *cobra.Command, args []string) {
	url := args[0]
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	var body io.Reader
	if postData != "" {
		body = strings.NewReader(postData)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", contentType)

	for _, header := range headers {
		key, value := parseHeader(header)
		if key != "" {
			req.Header.Set(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if outputFile != "" {
		err = os.WriteFile(outputFile, responseBody, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Response written to %s\n", outputFile)
	} else {
		fmt.Println(string(responseBody))
	}
}
