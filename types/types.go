package types

type Config struct {
	Community struct {
		Name        string `json:"name"`
		ConductLink string `json:"conduct_url"`
    Language string `json:"language"`
	} `json:"general"`

	Documents struct {
		LogoPath                    string `json:"logo"`
		AboutPath                   string `json:"about"`
		RegisterRulesPath           string `json:"rules"`
		VerificationExplanationPath string `json:"verification_instructions"`
	} `json:"documents"`
}

/*
config structure
["General"]
Name = "Merveilles"
ConductLink = "https://github.com/merveilles/Resources/blob/master/CONDUCT.md"


["Documents"]
LogoPath = "./logo.svg"
AboutPath = "./about.md"
RegisterRulesPath = "./rules.md"
VerificationExplanationPath = "./verification-instructions.md"
*/
