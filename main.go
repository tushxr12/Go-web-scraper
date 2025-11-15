package main

import (
	"encoding/json"
	"flag" // Now used!
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const MAX_WORKERS = 5

// ScrapeResult will hold the data extracted for each URL.
type ScrapeResult struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Links int    `json:"link_count"`
	Error string `json:"error,omitempty"` 
}

func main() {
	log.Println("Starting Concurrent Web Scraper...")

	// --- 1. CLI Input Implementation ---
	// Define flags to accept input from the command line
	var urlFlag string
	var outputFilename string

	flag.StringVar(&urlFlag, "urls", "", "Comma-separated list of URLs to scrape (e.g., https://example.com,https://go.dev)")
	flag.StringVar(&outputFilename, "output", "scrape_results.json", "Filename to save the results (e.g., results.json)")

	flag.Parse()

	if urlFlag == "" {
		// If no URLs are provided, show usage and exit gracefully
		fmt.Println("Error: No URLs provided.")
		flag.Usage()
		os.Exit(1)
	}

	// Convert the comma-separated string into a slice of URLs
	rawUrls := strings.Split(urlFlag, ",")
	urls := make([]string, 0, len(rawUrls))
	for _, u := range rawUrls {
		trimmedURL := strings.TrimSpace(u)
		if trimmedURL != "" {
			urls = append(urls, trimmedURL)
		}
	}
	
	if len(urls) == 0 {
		fmt.Println("Error: The provided URL list is empty or invalid.")
		flag.Usage()
		os.Exit(1)
	}
	// --- End CLI Input Implementation ---
	
	// --- 2. Setup concurrency primitives ---
	var wg sync.WaitGroup
	// The size of the channels is now the number of provided URLs
	urlsToProcess := make(chan string, len(urls))
	results := make(chan ScrapeResult, len(urls))

	startTime := time.Now()

	// --- 3. Start the worker pool (MAX_WORKERS goroutines) ---
	log.Printf("Starting %d concurrent workers...", MAX_WORKERS)
	for i := 0; i < MAX_WORKERS; i++ {
		wg.Add(1) 
		go scrapeWorker(urlsToProcess, results, &wg, i) 
	}

	// --- 4. Feed URLs into the input channel ---
	for _, url := range urls {
		// Basic validation: ensure URL has a schema
		if !strings.HasPrefix(url, "http") {
			url = "https://" + url
		}
		urlsToProcess <- url
	}
	close(urlsToProcess) 

	// --- 5. Wait for all workers to finish ---
	wg.Wait()
	close(results) 

	// --- 6. Collect and report results ---
	var allResults []ScrapeResult
	for result := range results {
		allResults = append(allResults, result)
	}

	// --- 7. Write results to file ---
	if err := writeResultsToFile(outputFilename, allResults); err != nil {
		log.Fatalf("Failed to write results to file: %v", err)
	}
	
	// Final Summary
	fmt.Printf("\n--- Scraping Complete ---\n")
	for _, res := range allResults {
		if res.Error != "" {
			fmt.Printf("[FAIL] %s: %s\n", res.URL, res.Error)
		} else {
			fmt.Printf("[OK]   %s: Title='%s', Links=%d\n", res.URL, res.Title, res.Links)
		}
	}
	fmt.Printf("\nSuccessfully saved %d results to %s\n", len(allResults), outputFilename)
	fmt.Printf("Total Time Taken: %s\n", time.Since(startTime).Round(time.Millisecond))
}

// scrapeWorker is the concurrent goroutine function.
func scrapeWorker(urls <-chan string, results chan<- ScrapeResult, wg *sync.WaitGroup, id int) {
	defer wg.Done()

	for url := range urls {
		log.Printf("[Worker %d] Scraping: %s", id, url)

		result := ScrapeResult{URL: url}
		client := http.Client{Timeout: 10 * time.Second} // Add a timeout for robustness

		// 1. Fetch the content
		resp, err := client.Get(url)
		if err != nil {
			result.Error = fmt.Sprintf("HTTP Request Failed: %v", err)
			results <- result
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			result.Error = fmt.Sprintf("Status Code Error: %d %s", resp.StatusCode, resp.Status)
			results <- result
			continue
		}

		// 2. Use goquery to parse the document
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			result.Error = fmt.Sprintf("Parsing Error: %v", err)
			results <- result
			continue
		}

		// 3. Extract data (Page Title and Link Count)
		result.Title = doc.Find("title").Text()
		result.Links = doc.Find("a").Length()

		// 4. Send the result to the output channel
		results <- result
		log.Printf("[Worker %d] Finished: %s", id, url)
	}
}

// writeResultsToFile takes the filename and the slice of results and writes them as JSON.
func writeResultsToFile(filename string, results []ScrapeResult) error {
	// Marshal the struct slice into a pretty-printed JSON byte array
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write the JSON bytes to the specified file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}