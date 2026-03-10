package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
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

func worker(id int, jobs <-chan string, results chan<- Result, client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range jobs {
		result := checkURL(client, url)
		results <- result
	}
}

func main() {
	outputFile := flag.String("o", "", "output file")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Usage: url-checker [-o output.txt] <urls_file>")
	}

	filePath := flag.Arg(0)

	urls, err := readURLs(filePath)
	if err != nil {
		log.Fatal("Error reading file: ", err)
	}

	var writer *os.File = os.Stdout
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			log.Fatal("Error creating file: ", err)
		}
		defer file.Close()
		writer = file
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	jobs := make(chan string, len(urls))
	results := make(chan Result, len(urls))

	var wg sync.WaitGroup

	workers := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, client, &wg)
	}

	for _, url := range urls {
		jobs <- url
	}

	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.Err != nil {
			_, err := fmt.Fprintf(writer, "%s ERROR: %v\n", result.URL, result.Err)
			if err != nil {
				log.Println("Write error:", err)
			}
		}

		_, err := fmt.Fprintf(writer, "%s STATUS: %d LATENCY: %v\n", result.URL, result.StatusCode, result.Latency)
		if err != nil {
			log.Println("Write error:", err)
		}
	}
}
