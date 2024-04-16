package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const baseUrl = "https://e-flux.com/journal/"

func generateIssueURL(issue int) string {
	var suffix string
	if issue < 10 {
		suffix = fmt.Sprintf("0%d", issue)
	} else {
		suffix = fmt.Sprintf("%d", issue)
	}
	return baseUrl + suffix
}

func findPDFLink(issueURL string) (string, error) {
	resp, err := http.Get(issueURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	var pdfLink string
	doc.Find("div.current-issue__content div.button-wrap--current-download a.button").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.Contains(href, ".pdf") {
			pdfLink = href
			return
		}
	})

	return pdfLink, nil
}

func downloadPDF(pdfURL, filePath string) error {
	resp, err := http.Get(pdfURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func main() {
	outputDir := flag.String("output", ".", "Directory to save downloaded PDFs")
	startIssue := flag.Int("startIssue", 1, "The starting issue number")
	endIssue := flag.Int("endIssue", 150, "The ending issue number") // Default assumes 150 issues available
	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Println("Failed to create output directory:", err)
		return
	}

	for i := *startIssue; i <= *endIssue; i++ {
		issueURL := generateIssueURL(i)
		fmt.Println("Processing:", issueURL)
		pdfLink, err := findPDFLink(issueURL)
		if err != nil {
			fmt.Println("Failed to find PDF link for issue", i, ":", err)
			continue
		}
		fmt.Println("Downloading PDF from:", pdfLink)

		fileName := fmt.Sprintf("e-flux-journal-issue-%d.pdf", i)
		filePath := filepath.Join(*outputDir, fileName)
		err = downloadPDF(pdfLink, filePath)
		if err != nil {
			fmt.Println("Failed to download PDF for issue", i, ":", err)
			continue
		}
		fmt.Println("Successfully downloaded:", filePath)
	}
}
