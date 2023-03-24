package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// There are four parts to the catechism
type Part struct {
	Title    string
	Sections []Section
}

// A section has many chapters
type Section struct {
	Title    string
	Chapters []Chapter
}

// A chapter has many articles
type Chapter struct {
	Parent   *Section
	Title    string
	Articles []Article
}

// An article has many sub-articles
type Article struct {
	Parent      *Chapter
	Title       string
	SubArticles []SubArticle
}

// A sub-article has many paragraphs
type SubArticle struct {
	Parent     *Article
	Title      string
	Paragraphs []Paragraph
}

// A paragraph has a number (e.g. 484) and text, as well as many
type Paragraph struct {
	Parent     *SubArticle
	Number     int // Paragraph numbers like 484 would correspond to "CCC 484" which starts with 'The Annunciation to Mary inaugurates "the fullness of time"'
	Text       string
	References []string
}

// This is the index of the official Catechism of the Catholic Church, in English
const vatican = "https://www.vatican.va"
const archeng = "/archive/ENG0015"

// This is the first page of the catechism
var vaticanFirstPage, _ = vaticanURL("/__P2.HTM")

func urlToFilename(urlStr string) string {
	// Parse the URL
	u, err := url.Parse(urlStr)
	if err != nil {
		fmt.Printf("error parsing url %s: %s", urlStr, err)
		os.Exit(1)
	}

	// Extract the path
	path := u.Path

	// Replace slashes with underscores and remove trailing slash
	path = strings.TrimRight(strings.ReplaceAll(path, "/", "_"), "_")

	// Remove any illegal characters using a regular expression
	illegalChars := regexp.MustCompile(`[<>:"|?*]`)
	path = illegalChars.ReplaceAllString(path, "")

	// Make the path safe for the filesystem
	return filepath.Clean(path)
}

// getOnce uses httputil.DumpResponse to store the response on disk,
// then uses http.ReadResponse to read the response from disk (./cache/url is the filename)
func getOnce(urlStr string) io.Reader {
	// Check if cached url is in ./cache/url file
	filename := fmt.Sprintf("cache/%s", urlToFilename(urlStr))
	//fmt.Printf("filename = %s\n", filename)
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// file doesn't exist, make an HTTP GET request
			var urlFullStr string = urlStr
			if !strings.HasPrefix(urlStr, "http") {
				urlFullStr, _ = vaticanURL(urlStr)
			}
			res, err := http.Get(urlFullStr)
			if err != nil {
				fmt.Printf("error getting url %s: %s\n", urlFullStr, err)
				os.Exit(1)
			}
			// dump the response body to raw bytes for caching
			body, err := httputil.DumpResponse(res, true)
			if err != nil {
				fmt.Printf("error dumping response: %\n", err)
				os.Exit(1)
			}
			//fmt.Printf("cacheing %s/\n", urlStr)
			// save the bytes to the ./cache folder so we don't have to request again
			file, err := os.Create(filename)
			if err != nil {
				fmt.Printf("error creating cache file %s: %s\n", filename, err)
				os.Exit(1)
			}
			defer file.Close()
			file.Write(body)
			defer res.Body.Close()
		}
	} else {
		//fmt.Printf("fetching %s from cache\n", urlStr)
	}
	// Open and read dumped response, and return the response
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("error reading file %s: %s", filename, data)
	}

	return bufio.NewReader(bytes.NewReader(data))
}

func getCatechism() map[int]Paragraph {
	var urlStr string = vaticanFirstPage
	var paragraphs map[int]Paragraph = make(map[int]Paragraph)

	// Get the first page of the Catechism
	for {
		body := getOnce(urlStr)
		// Create a goquery document
		doc, err := goquery.NewDocumentFromReader(body)
		if err != nil {
			fmt.Printf("error creating new goquery doc: %s", err)
			os.Exit(1)
		}
		// Extract Paragraphs from doc
		doc.Find("p").Each(func(_ int, s *goquery.Selection) {
			// Check for paragraph number
			num, startsWithNumber := extractNumber(s.Text())
			_, isStoredInMap := paragraphs[num]
			if startsWithNumber && !isStoredInMap {
				paragraphs[num] = Paragraph{
					Number: num,
					Text:   s.Text(),
				}
			}
		})
		// Get next link
		next := getNextLink(doc)
		if next == nil {
			//fmt.Printf("next is nil")
			return paragraphs
		} else {
			// Get urlStr to nextLink
			urlPath, _ := next.Attr("href")
			urlStr, err = vaticanURL(urlPath)
			if err != nil {
				fmt.Printf("error generating vaticanURL from urlPath = %s\n", urlPath)
			}
		}
	}
}

func main() {
	// Load the Catechism into the Paragraph array
	var paragraphs map[int]Paragraph = getCatechism()
	// Check for command arguments
	if len(os.Args) > 1 {
		paragraphNumber, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Printf("error parsing 1st arg from os.Args: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(paragraphs[paragraphNumber].Text)

	} else {
		for _, p := range paragraphs {
			text := strings.ReplaceAll(p.Text, "\n", " ")
			fmt.Printf("%s\n", text)
		}
	}
}

func getNextLink(doc *goquery.Document) *goquery.Selection {
	var next *goquery.Selection = nil
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		if s.Text() == "Next" {
			next = s
			return
		}
	})
	return next
}

func extractNumber(str string) (int, bool) {
	re := regexp.MustCompile(`^(\d+)`)
	matches := re.FindStringSubmatch(str)
	if len(matches) > 1 {
		num, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, false
		}
		return num, true
	}
	return 0, false
}

func vaticanURL(relativePath string) (string, error) {
	// Forgive these web developers, some next links are absolute and some are relative
	if strings.HasPrefix(strings.ToLower(relativePath), "http") {
		return relativePath, nil
	}
	u, err := url.Parse(vatican)
	if err != nil {
		return "", err
	}

	rel, err := url.Parse(path.Join(archeng, relativePath))
	if err != nil {
		return "", err
	}

	resolvedURL := u.ResolveReference(rel)
	return resolvedURL.String(), nil
}
