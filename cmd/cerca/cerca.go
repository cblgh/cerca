package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"cerca/server"
	"cerca/util"
)

func readAllowlist(location string) []string {
	ed := util.Describe("read allowlist")
	data, err := os.ReadFile(location)
	ed.Check(err, "read file")
	list := strings.Split(strings.TrimSpace(string(data)), "\n")
	var processed []string
	for _, fullpath := range list {
		u, err := url.Parse(fullpath)
		if err != nil {
			continue
		}
		processed = append(processed, u.Host)
	}
	return processed
}

func inform(msg string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("%s\n", fmt.Sprintf(msg, args...))
	} else {
		fmt.Printf("%s\n", msg)
	}
}

func complain(msg string, args ...interface{}) {
	if len(args) > 0 {
		inform(msg, args)
	} else {
		inform(msg)
	}
	os.Exit(0)
}

func cerca() {
	// TODO (2022-01-10): somehow continually update veri sites by scraping merveilles webring sites || webring domain
	var allowlistLocation string
	var sessionKey string
	var configPath string
	var dataDir string
	var dev bool

	flag.BoolVar(&dev, "dev", false, "trigger development mode")
	flag.StringVar(&allowlistLocation, "allowlist", "", "domains which can be used to read verification codes from during registration")
	flag.StringVar(&sessionKey, "authkey", "", "session cookies authentication key")
	flag.StringVar(&configPath, "config", "cerca.toml", "config and settings file containing cerca's customizations")
	flag.StringVar(&dataDir, "data", "./data", "directory where cerca will dump its database")

	help := `NAME:
  cerca - lean forum software

USAGE:
  cerca [command] [options]

COMMANDS:
  admin     promote user to admin
  migrate   manage database migrations
  user      create new user
  reset     reset user password

GLOBAL OPTIONS:
  -help
        show help (default: false)

OPTIONS:
`
	usage := func() {
		fmt.Fprintf(os.Stderr, help)
		flag.PrintDefaults()
	}
	flag.Usage = usage

	flag.Parse()

	if len(sessionKey) == 0 {
		if !dev {
			complain("please pass a random session auth key with --authkey")
		}
		sessionKey = "0"
	}
	if len(allowlistLocation) == 0 {
		if !dev {
			complain("please pass a file containing the verification code domain allowlist")
		}
		allowlistLocation = "temp-allowlist.txt"
		created, err := util.CreateIfNotExist(allowlistLocation, "")
		if err != nil {
			complain(fmt.Sprintf("couldn't create %s: %s", allowlistLocation, err))
		}
		if created {
			fmt.Println(fmt.Sprintf("Created %s", allowlistLocation))
		}
	}

	err := os.MkdirAll(dataDir, 0750)
	if err != nil {
		complain(fmt.Sprintf("couldn't create dir '%s'", dataDir))
	}
	allowlist := readAllowlist(allowlistLocation)
	allowlist = append(allowlist, "merveilles.town")
	config := util.ReadConfig(configPath)
	server.Serve(allowlist, sessionKey, dev, dataDir, config)
}

func main() {
	switch os.Args[1] {
	case "migrate":
		migrate()
	case "admin":
		admin()
	case "user":
		user()
	case "reset":
		reset()
	default:
		cerca()
	}
}
