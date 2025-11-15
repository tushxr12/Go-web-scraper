# Go-web-scraper
This project is a command-line interface (CLI) application built using Go (Golang) designed for rapid, concurrent extraction of data from multiple web pages.


🚀 How to Run

Prerequisites

Go installed on your machine (version 1.18 or higher recommended).

Run go get github.com/PuerkitoBio/goquery in the project directory.

Execution

Use the -urls flag to provide a comma-separated list of targets and the -output flag for the filename.

# Example 1: Basic Scrape
go run main.go -urls "[https://example.com](https://example.com),[https://go.dev](https://go.dev),[https://pkg.go.dev](https://pkg.go.dev)"

# Example 2: Scrape and save to a custom file
go run main.go -urls "[https://google.com](https://google.com),[https://github.com](https://github.com)" -output results/latest_scrape.json


Build (Standalone Binary)

Create a single executable file that can be run anywhere on your system:

go build -o web_scraper
./web_scraper -urls "[https://example.com](https://example.com)"
