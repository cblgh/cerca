package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cerca/defaults"
	"cerca/server"
	"cerca/util"
)

var commandExplanations = map[string]string{
	"run":        "run the forum",
	"adduser":    "create a new user",
	"makeadmin":  "make an existing user an admin",
	"migrate":    "manage database migrations",
	"resetpw":    "reset a user's password",
	"genauthkey": "generate and output an authkey for use with `cerca run`",
	"version":    "output version information",
}

func createHelpString(commandName string, usageExamples []string) string {
	helpString := fmt.Sprintf("USAGE:\n  %s\n\n  %s\n",
		commandExplanations[commandName],
		strings.Join(usageExamples, "\n  "))

	if commandName == "run" {
		helpString += "\nCOMMANDS:\n"
		cmds := []string{"adduser", "makeadmin", "migrate", "resetpw", "genauthkey", "version"}
		for _, key := range cmds {
			// pad first string with spaces to the right instead, set its expected width = 11
			helpString += fmt.Sprintf("  %-11s%s\n", key, commandExplanations[key])
		}
	}

	helpString += "\nOPTIONS:\n"
	return helpString
}

func usage(help string, fset *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, help)
	if fset != nil {
		fset.PrintDefaults()
		return
	}
	flag.PrintDefaults()
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

const DEFAULT_PORT = 8272
const DEFAULT_DEV_PORT = 8277

func run() {
	var configPath string
	var dev bool
	var port int

	flag.BoolVar(&dev, "dev", false, "trigger development mode")
	flag.IntVar(&port, "port", DEFAULT_PORT, "port to run the forum on")
	flag.StringVar(&configPath, "config", "cerca.toml", "config and settings file containing cerca's customizations")

	help := createHelpString("run", []string{
		"cerca -config <path-to-cerca.toml>",
		"cerca -config <path-to-cerca.toml> -dev",
	})
	flag.Usage = func() { usage(help, nil) }
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}
	// if dev mode and port not specified then use the default dev port to prevent collision with default serving port
	if dev && port == DEFAULT_PORT {
		port = DEFAULT_DEV_PORT
	}

	config := util.ReadConfig(configPath)
	_, err := util.CreateIfNotExist(filepath.Join("html", "assets", "theme.css"), defaults.DEFAULT_THEME)
	if err != nil {
		complain("couldn't output default theme.css")
	}

	if len(config.General.AuthKey) == 0 {
		if !dev {
			complain("please add an auth_key to the config file")
		}
		config.General.AuthKey = "0"
	}

	server.Serve(port, dev, config)
}

func main() {
	command := "run"
	if len(os.Args) > 1 && (os.Args[1][0] != '-') {
		command = os.Args[1]
	}

	switch command {
	case "adduser":
		user()
	case "makeadmin":
		admin()
	case "migrate":
		migrate()
	case "resetpw":
		reset()
	case "run":
		run()
	case "genauthkey":
		genauthkey()
	case "version":
		version()
	default:
		fmt.Printf("ERR: no such subcommand '%s'\n", command)
		run()
	}
}
