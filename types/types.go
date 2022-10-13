package types

type Config struct {
	// for internal use
	Files map[string]string
	// use as:
	// config.Files["about"] -> about markdown
	// config.Files["rules"] -> rules explanation markdown
	// config.Files["verification"] -> verification explanation

	Community struct {
		Name        string `json:"name"`
		Link        string `json:"link"`
		ConductLink string `json:"conduct_url"`
	} `json:"general"`

	Theme struct {
		Background string `json:"background"`
		Foreground string `json:"foreground"`
		Links      string `json:"links"`
	} `json:"theme"`

	Documents struct {
		LogoPath                    string `json:"logo"`
		AboutPath                   string `json:"about"`
		RegisterRulesPath           string `json:"rules"`
		VerificationExplanationPath string `json:"verification_explanation"`
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
