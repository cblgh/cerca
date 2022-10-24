package i18n

import (
	"cerca/util"
	"fmt"
	"html/template"
	"log"
	"strings"
)

const toolURL = "https://github.com/cblgh/cerca/releases/tag/pwtool-v1"

var English = map[string]string{
	"About":    "about",
	"Login":    "login",
	"Logout":   "logout",
	"Sort":     "sort",
	"Enter":    "enter",
	"Register": "register",

	"LoggedIn":    "logged in",
	"NotLoggedIn": "Not logged in",
	"LogIn":       "log in",
	"GoBack":      "Go back",

	"SortPostsRecent":   "recent posts",
	"SortThreadsRecent": "most recent threads",

	"LoginNoAccount":       "Don't have an account yet? <a href='/register'>Register</a> one.",
	"LoginFailure":         "<b>Failed login attempt:</b> incorrect password, wrong username, or a non-existent user.",
	"LoginAlreadyLoggedIn": `You are already logged in. Would you like to <a href="/logout">log out</a>?`,

	"Username":       "username",
	"Password":       "password",
	"PasswordMin":    "Must be at least 9 characters long",
	"PasswordForgot": "Forgot your password?",

	"Threads":   "threads",
	"ThreadNew": "new thread",
	"ThreadThe": "the thread",
	"Index":     "index",

	"ThreadCreate":        "Create thread",
	"Title":               "Title",
	"Content":             "Content",
	"Create":              "Create",
	"TextareaPlaceholder": "Tabula rasa",

	"PasswordReset":                   "reset password",
	"PasswordResetMessage":            "You are logged in, log out to reset password using proof",
	"PasswordResetSuccess":            "Reset password—success!",
	"PasswordResetSuccessMessage":     "You reset your password!",
	"PasswordResetSuccessLinkMessage": "Give it a try and",

	"RegisterMessage":     "You already have an account (you are logged in with it).",
	"RegisterLinkMessage": "Visit the",
	"RegisterSuccess":     "registered successfully",

	"ErrUnaccepted":        "Unaccepted request",
	"ErrThread404":         "Thread not found",
	"ErrThread404Message":  "The thread does not exist (anymore?)",
	"ErrGeneric404":        "Page not found",
	"ErrGeneric404Message": "The visited page does not exist (anymore?). Error code %d.",

	"NewThreadMessage":            "Only members of this forum may create new threads",
	"NewThreadLinkMessage":        "If you are a member,",
	"NewThreadCreateError":        "Error creating thread",
	"NewThreadCreateErrorMessage": "There was a database error when creating the thread, apologies.",

	"AriaPostMeta":          "Post meta",
	"AriaDeletePost":        "Delete this post",
	"AriaRespondIntoThread": "Respond into this thread",
	"PromptDeleteQuestion":  "Delete post for all posterity?",
	"Delete":                "delete",
	"Post":                  "post",
	"Author":                "Author",
	"Responded":             "responded",
	"YourAnswer":            "Your answer",

	"AriaHome":       "Home",
	"ThreadStartNew": "Start a new thread",

	"RegisterHTMLMessage":               `You now have an account! Welcome. Visit the <a href="/">index</a> to read and reply to threads, or start a new one.`,
	"RegisterKeypairExplanationStart":   `There's just one more thing: <b>save the key displayed below</b>. It is a <a href="https://en.wikipedia.org/wiki/Public-key_cryptography">keypair</a> describing your forum identity, with a private part that only you know; the forum only stores the public portion.`,
	"RegisterViewKeypairExplanationEnd": `With this keypair you will be able to reset your account if you ever lose your password—and without having to share your email (or require email infrastructure on the forum's part).`,
	"RegisterKeypairWarning":            "This keypair will only be displayed once",

	"RegisterVerificationCode":              "Your verification code is",
	"RegisterVerificationInstructionsTitle": "Verification instructions",
	"RegisterVerificationLink":              "Verification link",
	"RegisterConductCodeBoxOne":             `I have refreshed my memory of the <a target="_blank" href="{{ .Data.Link }}">{{ .Data.Name }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo":             `Yes, I have actually <a target="_blank" href="{{ .Data.Link }}">read it</a>`,

	"PasswordResetDescription":            "On this page we'll go through a few steps to securely reset your password—without resorting to any emails!",
	"PasswordResetUsernameQuestion":       "First up: what was your username?",
	"PasswordResetCopyPayload":            `Now, first copy the snippet (aka <i>proof payload</i>) below`,
	"PasswordResetFollowToolInstructions": `Follow the <b>tool instructions</b> to finalize the password reset.`,
	"ToolInstructions":                    `tool instructions`,
	"PasswordResetToolInstructions": fmt.Sprintf(`
    <ul>
        <li><a href="%s">Download the tool</a></li>
        <li>Run as:<br><code>pwtool --payload &lt;proof payload from above&gt; --keypair &lt;path to file with yr keypair from registration&gt;</code>
        </li>
        <li>Copy the generated proof and paste below</li>
        <li>(Remember to save your password :)</li>
    </ul>
    `, toolURL),
	"GeneratePayload": "generate payload",
	"Proof":           "proof",
	"NewPassword":     "new password",
	"ChangePassword":  "change password",
}

var Swedish = map[string]string{
	"About":    "om",
	"Login":    "logga in",
	"Logout":   "logga ut",
	"Sort":     "sortera",
	"Enter":    "skicka",
	"Register": "registrera",

	"LoggedIn":    "inloggad",
	"NotLoggedIn": "Ej inloggad",
	"LogIn":       "logga in",
	"GoBack":      "Go back",

	"SortRecentPosts":   "nyast poster",
	"SortRecentThreads": "nyast trådar",

	"LoginNoAccount":       "Saknar du konto? <a href='/register'>Skapa</a> ett.",
	"LoginFailure":         "<b>Misslyckat inloggningsförsök:</b> inkorrekt lösenord, fel användernamn, eller obefintlig användare.",
	"LoginAlreadyLoggedIn": `Du är redan inloggad. Vill du <a href="/logout">logga ut</a>?`,

	"Username":       "användarnamn",
	"Password":       "lösenord",
	"PasswordMin":    "Måste vara minst 9 karaktärer långt",
	"PasswordForgot": "Glömt lösenordet?",

	"Threads":   "trådar",
	"ThreadNew": "ny tråd",
	"ThreadThe": "tråden",
	"Index":     "index",

	"ThreadCreate":        "Skapa en tråd",
	"Title":               "Titel",
	"Content":             "Innehåll",
	"Create":              "Skapa",
	"TextareaPlaceholder": "Tabula rasa",

	"PasswordReset":                   "nollställ lösenord",
	"PasswordResetMessage":            "Du är inloggad, logga ut för att nollställga lösenordet med skapat lösenordsbevis",
	"PasswordResetSuccess":            "Nollställning av lösenord—lyckades!",
	"PasswordResetSuccessMessage":     "Du har nollställt ditt lösenord!",
	"PasswordResetSuccessLinkMessage": "Ge det ett försök och",

	"RegisterMessage":     "Du har redan ett konto (du är inloggad med det).",
	"RegisterLinkMessage": "Besök",
	"RegisterSuccess":     "konto skapat",

	"ErrUnaccepted":        "Ej accepterat request",
	"ErrThread404":         "Tråd ej funnen",
	"ErrThread404Message":  "Denna tråden finns ej (längre?)",
	"ErrGeneric404":        "Sida ej funnen",
	"ErrGeneric404Message": "Den besökta sidan finns ej (längre?). Felkod %d.",

	"NewThreadMessage":            "Enbart medlemmarna av detta forum får skapa nya trådar",
	"NewThreadLinkMessage":        "Om du är en medlem,",
	"NewThreadCreateError":        "Fel uppstod vid trådskapning",
	"NewThreadCreateErrorMessage": "Det uppstod ett databasfel under trådskapningen, ursäkta.",

	"AriaPostMeta":          "Post meta",
	"AriaDeletePost":        "Delete this post",
	"AriaRespondIntoThread": "Respond into this thread",
	"PromptDeleteQuestion":  "Radera post för alltid?",
	"Delete":                "radera",
	"Post":                  "post",
	"Author":                "Författare",
	"Responded":             "svarade",
	"YourAnswer":            "Ditt svar",

	"AriaHome":       "Hem",
	"ThreadStartNew": "Starta ny tråd",

	"RegisterHTMLMessage":               `Du har nu ett konto! Välkommen. Besök <a href="/">trådindexet</a> för att läsa och svara på trådar, eller för att starta en ny.`,
	"RegisterKeypairExplanationStart":   `En grej till: <b>spara nyckeln du ser nedan</b>. Det är ett <a href="https://en.wikipedia.org/wiki/Public-key_cryptography">nyckelpar</a> som tillhandahåller din forumidentitet, och inkluderar en hemlig del som bara du vet om och endast visas nu; forumdatabasen kommer enbart ihåg den publika delen.`,
	"RegisterViewKeypairExplanationEnd": `Med detta nyckelpar kan du återställa ditt lösenord om du skulle tappa bort det—och detta utan att behöva samla in din email (eller kräva emailinfrastruktur på forumets sida).`,
	"RegisterKeypairWarning":            "Detta nyckelpar visas enbart denna gång",

	"RegisterVerificationCode":              "Din verifikationskod är",
	"RegisterVerificationInstructionsTitle": "Verification instructions",
	"RegisterVerificationLink":              "Verificationsnyckel",
	"RegisterConductCodeBoxOne":             `I have refreshed my memory of the <a target="_blank" href="{{ .Data.Link }}">{{ .Data.Name }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo":             `Yes, I have actually <a target="_blank" href="{{ .Data.Link }}">read it</a>`,

	"PasswordResetDescription":            "På denna sida går vi igenom ett par steg för att säkert nollställa ditt lösenord—utan att behöva ta till mejl!",
	"PasswordResetUsernameQuestion":       "För de första: hur löd användarnamnet?",
	"PasswordResetCopyPayload":            `Kopiera nu textsnutten nedan (aka <i>beviset</i>)`,
	"PasswordResetFollowToolInstructions": `Följ <b>verktygsinstruktionerna</b> för att finalisera nollställningen.`,
	"ToolInstructions":                    `verktygsinstruktionerna`,
	"PasswordResetToolInstructions": fmt.Sprintf(`
    <ul>
        <li><a href="%s">Ladda ned verktyget</a></li>
        <li>Kör det så hör:<br><code>pwtool --payload &lt;payload från ovan&gt; --keypair &lt;filvägen innehållandes ditt nyckelpar från när du registrerade dig&gt;</code>
        </li>
        <li>Kopiera det genererade beviset och klistra in nedan</li>
        <li>(Kom ihåg att spara ditt lösenord:)</li>
    </ul>
    `, toolURL),
	"GeneratePayload": "skapa payload",
	"Proof":           "bevis",
	"NewPassword":     "nytt lösenord",
	"ChangePassword":  "ändra lösenord",
}

var EspanolMexicano = map[string]string{
	"About":    "acerca de",
	"Login":    "loguearse",
	"Logout":   "logout",
	"Sort":     "sort",
	"Register": "register",
	"Enter":    "entrar",

	"LoggedIn":    "logged in",
	"NotLoggedIn": "Not logged in",
	"LogIn":       "log in",
	"GoBack":      "Go back",

	"SortRecentPosts":   "recent posts",
	"SortRecentThreads": "most recent threads",

	"LoginNoAccount":       "¿No tienes una cuenta? <a href='/register'>Registra</a> una. ",
	"LoginFailure":         "<b>Failed login attempt:</b> incorrect password, wrong username, or a non-existent user.",
	"LoginAlreadyLoggedIn": `You are already logged in. Would you like to <a href="/logout">log out</a>?`,

	"Username":       "usuarie",
	"Password":       "contraseña",
	"PasswordMin":    "Debe tener por lo menos 9 caracteres.",
	"PasswordForgot": "Olvidaste tu contraseña?",

	"Threads":   "threads",
	"ThreadNew": "new thread",
	"ThreadThe": "the thread",
	"Index":     "index",

	"ThreadCreate":        "Create thread",
	"Title":               "Title",
	"Content":             "Content",
	"Create":              "Create",
	"TextareaPlaceholder": "Tabula rasa",

	"PasswordReset":                   "reset password",
	"PasswordResetMessage":            "You are logged in, log out to reset password using proof",
	"PasswordResetSuccess":            "Reset password—success!",
	"PasswordResetSuccessMessage":     "You reset your password!",
	"PasswordResetSuccessLinkMessage": "Give it a try and",

	"RegisterMessage":     "You already have an account (you are logged in with it).",
	"RegisterLinkMessage": "Visit the",
	"RegisterSuccess":     "registered successfully",

	"ErrUnaccepted":        "Unaccepted request",
	"ErrThread404":         "Thread not found",
	"ErrThread404Message":  "The thread does not exist (anymore?)",
	"ErrGeneric404":        "Page not found",
	"ErrGeneric404Message": "The visited page does not exist (anymore?). Error code %d.",

	"NewThreadMessage":            "Only members of this forum may create new threads",
	"NewThreadLinkMessage":        "If you are a member,",
	"NewThreadCreateError":        "Error creating thread",
	"NewThreadCreateErrorMessage": "There was a database error when creating the thread, apologies.",
	"ThreadStartNew":              "Start a new thread",

	"AriaPostMeta":          "Post meta",
	"AriaDeletePost":        "Delete this post",
	"AriaRespondIntoThread": "Respond into this thread",
	"AriaHome":              "Home",
	"PromptDeleteQuestion":  "Delete post for all posterity?",
	"Delete":                "delete",
	"Post":                  "post",
	"Author":                "Author",
	"Responded":             "responded",
	"YourAnswer":            "Your answer",

	"RegisterHTMLMessage":               `You now have an account! Welcome. Visit the <a href="/">index</a> to read and reply to threads, or start a new one.`,
	"RegisterKeypairExplanationStart":   `There's just one more thing: <b>save the key displayed below</b>. It is a <a href="https://en.wikipedia.org/wiki/Public-key_cryptography">keypair</a> describing your forum identity, with a private part that only you know; the forum only stores the public portion.`,
	"RegisterViewKeypairExplanationEnd": `With this keypair you will be able to reset your account if you ever lose your password—and without having to share your email (or require email infrastructure on the forum's part).`,
	"RegisterKeypairWarning":            "This keypair will only be displayed once",

	"RegisterVerificationCode":              "Your verification code is",
	"RegisterVerificationInstructionsTitle": "Verification instructions",
	"RegisterVerificationLink":              "Verification link",
	"RegisterConductCodeBoxOne":             `I have refreshed my memory of the <a target="_blank" href="{{ .Data.Link }}">{{ .Data.Name }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo":             `Yes, I have actually <a target="_blank" href="{{ .Data.Link }}">read it</a>`,

	"PasswordResetDescription":            "On this page we'll go through a few steps to securely reset your password—without resorting to any emails!",
	"PasswordResetUsernameQuestion":       "First up: what was your username?",
	"PasswordResetCopyPayload":            `Now, first copy the snippet (aka <i>proof payload</i>) below`,
	"PasswordResetFollowToolInstructions": `Follow the <b>tool instructions</b> to finalize the password reset.`,
	"ToolInstructions":                    `tool instructions`,
	"PasswordResetToolInstructions": fmt.Sprintf(`
    <ul>
        <li><a href="%s">Download the tool</a></li>
        <li>Run as:<br><code>pwtool --payload &lt;proof payload from above&gt; --keypair &lt;path to file with yr keypair from registration&gt;</code>
        </li>
        <li>Copy the generated proof and paste below</li>
        <li>(Remember to save your password :)</li>
    </ul>
    `, toolURL),
	"GeneratePayload": "generate payload",
	"Proof":           "proof",
	"NewPassword":     "new password",
	"ChangePassword":  "change password",
}

var translations = map[string]map[string]string{
	"English":         English,
	"EspañolMexicano": EspanolMexicano,
	"Swedish":         Swedish,
}

type TranslationData struct {
	Data interface{}
}

func (tr *Translator) TranslateWithData(key string, data TranslationData) string {
	phrase := translations[tr.Language][key]
	t, err := template.New(key).Parse(phrase)
	ed := util.Describe("i18n translation")
	ed.Check(err, "parse translation phrase")
	sb := new(strings.Builder)
	err = t.Execute(sb, data)
	ed.Check(err, "execute template with data")
	return sb.String()
}

func (tr *Translator) Translate(key string) string {
	var empty TranslationData
	return tr.TranslateWithData(key, empty)
}

type Translator struct {
	Language string
}

func Init(lang string) Translator {
	if _, ok := translations[lang]; !ok {
		log.Fatalln(fmt.Sprintf("language '%s' is not translated yet", lang))
	}
	return Translator{lang}
}

// usage:
// 	  tr := Init("EnglishSwedish")
// 	  fmt.Println(tr.Translate("LoginNoAccount"))
// 	  fmt.Println(tr.TranslateWithData("LoginDescription", Community{"Merveilles", "https://merveill.es"}))
