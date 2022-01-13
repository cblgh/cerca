package util

import (
	"html/template"
	"net/url"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

func Contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

var contentGuardian = bluemonday.UGCPolicy()
var strictContentGuardian = bluemonday.StrictPolicy()

// Turns Markdown input into HTML
func Markup(md template.HTML) template.HTML {
	mdBytes := []byte(string(md))
	// fix newlines
	mdBytes = markdown.NormalizeNewlines(mdBytes)
	maybeUnsafeHTML := markdown.ToHTML(mdBytes, nil, nil)
	// guard against malicious code being embedded
	html := contentGuardian.SanitizeBytes(maybeUnsafeHTML)
	return template.HTML(html)
}

func SanitizeStringStrict(s string) string {
	return strictContentGuardian.Sanitize(s)
}

// make a string be suitable for use as part of a url
func SanitizeURL(input string) string {
	input = strings.ReplaceAll(input, " ", "-")
	input = url.PathEscape(input)
	// TODO(2022-01-08): evaluate use of strict content guardian?
	return strings.ToLower(input)
}
