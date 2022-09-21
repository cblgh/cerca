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

	"ForumDescription":     "This forum is for the <a href='{{ .CommunityLink }}'>{{.CommunityName}}</a> community.",
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

	"RegisterRules": `To register, you need to either belong to the <a href="https://webring.xxiivv.com">Merveilles Webring</a> or the <a href="https://merveilles.town">Merveilles Fediverse instance</a>`,

	"RegisterVerificationCode":              "Your verification code is",
	"RegisterVerificationInstructionsTitle": "Verification instructions",
	// TODO (2022-09-20): make verification instructions another md file to load, pass path from config
	"RegisterVerificationInstructions": `<p>You can use either your mastodon profile or your webring site to verify your registration.</p>
    <ul>
        <li><b>Mastodon:</b> temporarily add a new metadata item to <a href="https://merveilles.town/settings/profile">your profile</a> containing the verification code
            displayed above. Pass your profile as the verification link.</li>
        <li><b>Webring site:</b> Upload a plaintext file somewhere on your webring domain (incl. subdomain) containing
            the verification code from above. Pass the link to the uploaded file as the verification link (make sure it is viewable in a browser).</li>
    </ul>
  `,

	"RegisterVerificationLink":  "Verification link",
	"RegisterConductCodeBoxOne": `I have refreshed my memory of the <a target="_blank" href="https://github.com/merveilles/Resources/blob/master/CONDUCT.md">{{ .CommunityName }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo": `Yes, I have actually <a target="_blank" href="https://github.com/merveilles/Resources/blob/master/CONDUCT.md">read it</a>`,

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

	"ForumDescription":     "Este foro es principalmente para las personas de la comunidad <a href='{{ .CommunityLink }}'>{{ .CommunityName }}</a>.",
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

	"RegisterRules": `To register, you need to either belong to the <a href="https://webring.xxiivv.com">Merveilles Webring</a> or the <a href="https://merveilles.town">Merveilles Fediverse instance</a>`,

	"RegisterVerificationCode":              "Your verification code is",
	"RegisterVerificationInstructionsTitle": "Verification instructions",
	// TODO (2022-09-20): make verification instructions another md file to load, pass path from config
	"RegisterVerificationInstructions": `<p>You can use either your mastodon profile or your webring site to verify your registration.</p>
    <ul>
        <li><b>Mastodon:</b> temporarily add a new metadata item to <a href="https://merveilles.town/settings/profile">your profile</a> containing the verification code
            displayed above. Pass your profile as the verification link.</li>
        <li><b>Webring site:</b> Upload a plaintext file somewhere on your webring domain (incl. subdomain) containing
            the verification code from above. Pass the link to the uploaded file as the verification link (make sure it is viewable in a browser).</li>
    </ul>
  `,
	"RegisterVerificationLink":  "Verification link",
	"RegisterConductCodeBoxOne": `I have refreshed my memory of the <a target="_blank" href="https://github.com/merveilles/Resources/blob/master/CONDUCT.md">{{ .CommunityName }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo": `Yes, I have actually <a target="_blank" href="https://github.com/merveilles/Resources/blob/master/CONDUCT.md">read it</a>`,

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
}

type Community struct {
	CommunityName string
	CommunityLink string
}

func (tr *Translator) TranslateWithData(key string, data Community) string {
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
	var empty Community
	return tr.TranslateWithData(key, empty)
}

type Translator struct {
	Language string
}

func Init(lang string) Translator {
	if _, ok := translations[lang]; !ok {
		log.Fatalln(lang + " is not translated yet")
	}
	return Translator{lang}
}

// usage:
// 	  tr := Init("EnglishSwedish")
// 	  fmt.Println(tr.Translate("LoginNoAccount"))
// 	  fmt.Println(tr.TranslateWithData("LoginDescription", Community{"Merveilles", "https://merveill.es"}))
