package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	Latency    time.Duration
	Err        error
}

func readURLs(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		urls = append(urls, line)
	}

	return urls, scanner.Err()
}

func checkURL(client *http.Client, url string) Result {
	start := time.Now()
	resp, err := client.Get(url)
	latency := time.Since(start)

	if err != nil {
		return Result{
			URL:     url,
			Latency: latency,
			Err:     err,
		}
	}

	defer resp.Body.Close()

	return Result{
		URL:        url,
		StatusCode: resp.StatusCode,
		Latency:    latency,
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: url-checker <url>")
	}

	filePath := os.Args[1]

	urls, err := readURLs(filePath)
	if err != nil {
		log.Fatal("Error reading file: ", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, url := range urls {
		result := checkURL(client, url)

		if result.Err != nil {
			fmt.Printf("%s ERROR: %v\n", result.URL, result.Err)
		}

		fmt.Printf("%s STATUS: %d LATENCT: %v\n", result.URL, result.StatusCode, result.Latency)
	}
}
