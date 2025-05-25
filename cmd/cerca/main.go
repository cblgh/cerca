package main

import (
	"flag"
	"fmt"
	"os"
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
const DEFAULT_CONFIG_NAME = "cerca.toml"
const DEFAULT_DATA_DIR = "./data"

func run() {
	var sessionKey string
	var configPath string
	var dataDir string
	var dev bool
	var port int

	flag.BoolVar(&dev, "dev", false, "trigger development mode")
	flag.IntVar(&port, "port", DEFAULT_PORT, "port to run the forum on")
	flag.StringVar(&sessionKey, "authkey", "", "session cookies authentication key")
	flag.StringVar(&configPath, "config", DEFAULT_CONFIG_NAME, "config and settings file containing cerca's customizations")
	flag.StringVar(&dataDir, "data", DEFAULT_DATA_DIR, "directory where cerca will dump its database")

	help := createHelpString("run", []string{
		"cerca -authkey \"CHANGEME\"",
		"cerca -dev",
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

	if len(sessionKey) == 0 {
		if !dev {
			complain("please pass a random session auth key with --authkey")
		}
		sessionKey = "0"
	}

	cercaRoot := util.GetEnvCercaRoot()

	// if --config isn't set and CERCA_ROOT is set, set config path to use the cerca root/cerca.toml
	if configPath == DEFAULT_CONFIG_NAME && cercaRoot != "" {
		configPath = util.JoinWithBase(cercaRoot, DEFAULT_CONFIG_NAME)
	}

	config := util.ReadConfig(configPath)
	// if CERCA_ROOT env var is set but config.Tooling.CercaRoot is not set, then use the CERCA_ROOT value.
	// note that implicitly, config.Tooling.CercaRoot takes precedence over any env variable CERCA_ROOT if both are set
	if cercaRoot != "" && config.Tooling.CercaRoot == "" {
		config.Tooling.CercaRoot = cercaRoot
	}

	// if --data is not passed, then join the default dataDir with the cerca root
	if dataDir == DEFAULT_DATA_DIR && config.Tooling.CercaRoot != "" {
		dataDir = config.JoinWithRoot(dataDir)
	}

	err := os.MkdirAll(dataDir, 0750)
	if err != nil {
		complain(fmt.Sprintf("couldn't create dir '%s'", dataDir))
	}

	_, err = util.CreateIfNotExist(config.JoinWithRoot("html", "assets", "theme.css"), defaults.DEFAULT_THEME)
	if err != nil {
		complain("couldn't output default theme.css")
	}
	server.Serve(sessionKey, port, dev, dataDir, config)
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
		authkey()
	case "version":
		version()
	default:
		fmt.Printf("ERR: no such subcommand '%s'\n", command)
		run()
	}
}
