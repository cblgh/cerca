package types

type Config struct {
	// for internal use
	Files map[string]string
	// use as:
	// config.Files["about"] -> about markdown
	// config.Files["rules"] -> rules explanation markdown
	// config.Files["verification"] -> verification explanation

	Community struct {
		Name        string
		Link        string
		ConductLink string
	} `json:"general"`

	Theme struct {
		Background string
		Foreground string
		Links      string
	} `json:"theme"`

	Documents struct {
		LogoPath                    string
		AboutPath                   string
		RegisterRulesPath           string
		VerificationExplanationPath string
	} `json:"documents"`
}

/*
config.Community.Name
config.Community.Link
config.Community.ConductLink

config structure
["General"]
Name = "Merveilles"
Link = "https://wiki.xxiivv.com/site/merveilles.html"
ConductLink = "https://github.com/merveilles/Resources/blob/master/CONDUCT.md"


["Documents"]
LogoPath = "./logo.svg"
AboutPath = "./about.md"
RegisterRulesPath = "./rules.md"
VerificationExplanationPath = "./verification-instructions.md"
*/
