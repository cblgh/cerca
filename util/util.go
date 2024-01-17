package util

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/komkom/toml"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/exp/utf8string"

	"cerca/defaults"
	"cerca/types"
)

/* util.Eout example invocations
if err != nil {
  return util.Eout(err, "reading data")
}
if err = util.Eout(err, "reading data"); err != nil {
  return nil, err
}
*/

type ErrorDescriber struct {
	environ string // the basic context that is potentially generating errors (like a GetThread function, the environ would be "get thread")
}

// parametrize Eout/Check such that error messages contain a defined context/environ
func Describe(environ string) ErrorDescriber {
	return ErrorDescriber{environ}
}

func (ed ErrorDescriber) Eout(err error, msg string, args ...interface{}) error {
	msg = fmt.Sprintf("%s: %s", ed.environ, msg)
	return Eout(err, msg, args...)
}

func (ed ErrorDescriber) Check(err error, msg string, args ...interface{}) {
	msg = fmt.Sprintf("%s: %s", ed.environ, msg)
	Check(err, msg, args...)
}

// format all errors consistently, and provide context for the error using the string `msg`
func Eout(err error, msg string, args ...interface{}) error {
	if err != nil {
		// received an invocation of e.g. format:
		// Eout(err, "reading data for %s and %s", "database item", "weird user")
		if len(args) > 0 {
			return fmt.Errorf("%s (%w)", fmt.Sprintf(msg, args...), err)
		}
		return fmt.Errorf("%s (%w)", msg, err)
	}
	return nil
}

func Check(err error, msg string, args ...interface{}) {
	if len(args) > 0 {
		err = Eout(err, msg, args...)
	} else {
		err = Eout(err, msg)
	}
	if err != nil {
		log.Fatalln(err)
	}
}

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
func Markup(md string) template.HTML {
	mdBytes := []byte(string(md))
	// fix newlines
	mdBytes = markdown.NormalizeNewlines(mdBytes)
	mdParser := parser.NewWithExtensions(parser.CommonExtensions ^ parser.MathJax)
	maybeUnsafeHTML := markdown.ToHTML(mdBytes, mdParser, nil)
	// guard against malicious code being embedded
	html := contentGuardian.SanitizeBytes(maybeUnsafeHTML)
	// lazy load images
	pattern := regexp.MustCompile("<img")
	if pattern.Match(html) {
		html = pattern.ReplaceAll(html, []byte(`<img loading="lazy"`))
	}
	return template.HTML(html)
}

func SanitizeStringStrict(s string) string {
	return strictContentGuardian.Sanitize(s)
}

func VerificationPrefix(name string) string {
	pattern := regexp.MustCompile("A|E|O|U|I|Y")
	upper := strings.ToUpper(name)
	replaced := string(pattern.ReplaceAll([]byte(upper), []byte("")))
	if len(replaced) < 3 {
		replaced += "XYZ"
	}
	return replaced[0:3]
}

func GetThreadSlug(threadid int, title string, threadLen int) string {
	return fmt.Sprintf("/thread/%d/%s-%d/", threadid, SanitizeURL(title), threadLen)
}

func Hex2Base64(s string) (string, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString(b)
	return b64, nil
}

// make a string be suitable for use as part of a url
func SanitizeURL(input string) string {
	input = strings.ReplaceAll(input, " ", "-")
	input = url.PathEscape(input)
	// TODO(2022-01-08): evaluate use of strict content guardian?
	return strings.ToLower(input)
}

// returns an id from a url path, and a boolean. the boolean is true if we're returning what we expect; false if the
// operation failed
func GetURLPortion(req *http.Request, index int) (int, bool) {
	var desiredID int
	parts := strings.Split(strings.TrimSpace(req.URL.Path), "/")
	if len(parts) < index || parts[index] == "" {
		return -1, false
	}
	desiredID, err := strconv.Atoi(parts[index])
	if err != nil {
		return -1, false
	}
	return desiredID, true
}

func Capitalize(s string) string {
	// utf8 safe capitalization
	str := utf8string.NewString(s)
	first := string(str.At(0))
	rest := string(str.Slice(1, str.RuneCount()))
	return strings.ToUpper(first) + rest
}

func CreateIfNotExist(docpath, content string) (bool, error) {
	err := os.MkdirAll(filepath.Dir(docpath), 0750)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(docpath)
	if err != nil {
		// if the file doesn't exist, create it
		if errors.Is(err, fs.ErrNotExist) {
			err = os.WriteFile(docpath, []byte(content), 0777)
			if err != nil {
				return false, err
			}
			// created file successfully
			return true, nil
		} else {
			return false, err
		}
	}
	return false, nil
}

const solarYearSecs = 31556926

func RelativeTime(t time.Time) string {
	d := time.Since(t)
	var metric string
	var amount int
	if d.Seconds() < 60 {
		amount = int(d.Seconds())
		metric = "second"
	} else if d.Minutes() < 60 {
		amount = int(d.Minutes())
		metric = "minute"
	} else if d.Hours() < 24 {
		amount = int(d.Hours())
		metric = "hour"
	} else if d.Seconds() < solarYearSecs {
		amount = int(d.Hours()) / 24
		metric = "day"
	} else {
		amount = int(d.Seconds()) / solarYearSecs
		metric = "year"
	}
	if amount == 1 {
		return fmt.Sprintf("%d %s ago", amount, metric)
	} else {
		return fmt.Sprintf("%d %ss ago", amount, metric)
	}
}

func ReadConfig(confpath string) types.Config {
	ed := Describe("config")
	_, err := CreateIfNotExist(confpath, defaults.DEFAULT_CONFIG)
	ed.Check(err, "create default config")

	data, err := os.ReadFile(confpath)
	ed.Check(err, "read file")

	var conf types.Config
	decoder := json.NewDecoder(toml.New(bytes.NewBuffer(data)))

	err = decoder.Decode(&conf)
	ed.Check(err, "decode toml with json decoder")

	return conf
}

func LoadFile(key, docpath, defaultContent string) ([]byte, error) {
	ed := Describe("load file")
	_, err := CreateIfNotExist(docpath, defaultContent)
	err = ed.Eout(err, "create if not exist (%s) %s", key, docpath)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(docpath)
	err = ed.Eout(err, "read %s", docpath)
	if err != nil {
		return nil, err
	}
	return data, nil
}
