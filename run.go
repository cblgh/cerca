package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"cerca/server"

	"github.com/antchfx/htmlquery"
)

func readAllowlist() []string {
	doc, err := htmlquery.LoadURL("https://webring.xxiivv.com")
	if err != nil {
		log.Println("err fetching webring", err)
	}

	var processed []string
	// query for links in the ordered list (ol), that do not contain a class
	// (otherwise, we'd get duplicates such as liked webrings with twtxt or rss
	// feeds)
	list := htmlquery.Find(doc, "//ol//a[not(@class)]/@href")
	for _, n := range list {
		ring_url := htmlquery.SelectAttr(n, "href")

		u, err := url.Parse(ring_url)
		if err != nil {
			log.Fatal(err)
		}
		processed = append(processed, u.Host)
	}
	return processed
}

func complain(msg string) {
	fmt.Printf("cerca: %s\n", msg)
	os.Exit(0)
}

func main() {
	// TODO (2022-01-10): somehow continually update veri sites by scraping merveilles webring sites || webring domain
	var sessionKey string
	var dev bool
	flag.BoolVar(&dev, "dev", false, "trigger development mode")
	flag.StringVar(&sessionKey, "authkey", "", "session cookies authentication key")
	flag.Parse()
	if len(sessionKey) == 0 {
		complain("please pass a random session auth key with --authkey")
	}
	allowlist := readAllowlist()
	allowlist = append(allowlist, "merveilles.town")
	server.Serve(allowlist, sessionKey, dev)
}
