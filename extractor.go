package main

import (
	"bytes"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	log "github.com/sirupsen/logrus"
)

type Issue struct {
	Headers   map[string]int     // header::start_index
	Articles  map[string]Article // header::article
	IssueDate time.Time          // Month Year
	IssueNum  int
	Raw       string //unprocessed
}

type Article struct {
	Author      string
	LegalQuotes []string
	Text        string
}

func extractIssue(issueString string, issueNum int, issueDate time.Time) Issue {
	var issue Issue
	issue.Raw = issueString
	// Order is important
	issue.IssueNum = 15
	issue.IssueDate = formatIssueDate(2021, 10)
	issue.Headers = issue.extractHeaders()
	issue.Articles = issue.extractArticles()
	for header, article := range issue.Articles {
		log.Infoln(header)
		log.Infoln(article.Text[:200])
		log.Infoln(article.Text[len(article.Text)-200:])
		log.Infoln(article.LegalQuotes)
	}
	return issue
}

func (is Issue) extractArticles() map[string]Article {
	log.Infoln("Extracting articles...")
	out := make(map[string]Article)
	keys := make([]string, len(is.Headers))
	ix := make([]int, 0)
	for _, i := range is.Headers {
		ix = append(ix, i)
	}
	slices.Sort(ix)
	for k, page_i := range is.Headers {
		i, _ := slices.BinarySearch(ix, page_i)
		keys = slices.Insert(keys, i, k)
	}
	keys = removeEmptyStrings(keys)

	for i, head := range keys {
		var start int
		var end int
		var article Article
		start = is.Headers[head]
		if i == len(keys)-1 {
			end = len(is.Raw) - 1
		} else {
			end = is.Headers[keys[i+1]]
		}
		article.Text = is.Raw[start:end]
		article.Author = "bpra.info" //TODO is there an automatic way we can extract this? Assume: no
		article.LegalQuotes = article.extractLegalQuotes()

		out[head] = article
	}
	return out
}

func (ar Article) extractLegalQuotes() []string {
	log.Infoln("Extracting Legal Quotes for article...")
	var out []string
	for _, legalAbreviation := range legalDocs {
		refMatch := regexp.MustCompile(legalAbreviation)
		isMentioned := refMatch.MatchString(ar.Text)
		if isMentioned && !slices.Contains(out, legalAbreviation) {
			out = append(out, legalAbreviation)
		}
	}
	return out
}

func (is Issue) extractHeaders() map[string]int {
	is.Headers = make(map[string]int)
	orderedRules := [][]string{
		{"В ТОВА ИЗДАНИЕ ", ""},
		{"^\\s+\\d+\\s+", ""},
		{"(.*)\\s+", "$1"}, // Remove trailing whitespace
	}
	refMatch := regexp.MustCompile("В ТОВА ИЗДАНИЕ.*СТР ")
	rawHeaders := refMatch.FindString(is.Raw)
	headers := strings.Split(rawHeaders, "СТР")
	headers = removeEmptyStrings(headers)
	log.Infof("Found %v headers", len(headers))
	for _, head := range headers {
		head = processStrings(head, orderedRules)
		// out = append(out, head)
		refMatch = regexp.MustCompile(head)
		indices := refMatch.FindAllStringIndex(is.Raw, -1)
		if len(indices) < 2 {
			log.Warnf("Couldn't find header index for %v, skipping", head)
			is.Headers[head] = -1
		} else {
			i := indices[1][0]
			is.Headers[head] = i
		}
	}
	return is.Headers

}

func loadPdf(file string) (string, error) {
	content, err := readPdf(file) // Read local pdf file
	if err != nil {
		return "", err
	}
	//Cleanup
	cleanupRules := [][]string{
		{"\\s+", " "},
	}
	return processStrings(content, cleanupRules), nil
}

func readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}

func formatIssueDate(year, month int) time.Time {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
}

// Removes empty strings from slices.
func removeEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" && str != " " {
			r = append(r, str)
		}
	}
	return r
}

// Helper function for applying a sequential order of regex string replacements
// raw - the string to be processed
// orderedRules - pairs of [MATCH, REPLACE] rules that will be executed in order
func processStrings(raw string, orderedRules [][]string) string {
	for _, rule := range orderedRules {
		refMatch := regexp.MustCompile(rule[0])
		raw = refMatch.ReplaceAllString(raw, rule[1])
	}
	return raw
}
