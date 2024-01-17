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
	"Bottom":   "bottom",

	"LoggedIn":    "logged in",
	"NotLoggedIn": "Not logged in",
	"LogIn":       "log in",
	"GoBack":      "Go back",

	"SortRecentPosts":   "recent posts",
	"SortRecentThreads": "most recent threads",

  "modlogResetPassword": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> reset a user's password`,
  "modlogResetPasswordAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> reset <b> {{ .Data.RecipientUsername}}</b>'s password`,
  "modlogRemoveUser": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> removed a user's account`,
  "modlogMakeAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> made <b> {{ .Data.RecipientUsername}}</b> an admin`,
  "modlogAddUser": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> manually registered an account for a new user`,
  "modlogAddUserAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> manually registered an account for <b> {{ .Data.RecipientUsername }}</b>`,
  "modlogDemoteAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> demoted <b> 
	{{ if eq .Data.ActingUsername .Data.RecipientUsername }} themselves 
	{{ else }} {{ .Data.RecipientUsername}} {{ end }}</b> from admin back to normal user`,
	"modlogXProposedY": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> proposed: {{ .Data.Action }}`,
	"modlogProposalMakeAdmin": `Make <b> {{ .Data.RecipientUsername}}</b> admin`,
	"modlogProposalDemoteAdmin": `Demote <b> {{ .Data.RecipientUsername}}</b> from role admin`,
	"modlogProposalRemoveUser": `Remove user <b> {{ .Data.RecipientUsername }} </b>`,
	"modlogConfirm": "{{ .Data.Action }} <i>confirmed by {{ .Data.ActingUsername }}</i>",
	"modlogVeto": "<s>{{ .Data.Action }}</s> <i>vetoed by {{ .Data.ActingUsername }}</i>",



	"Admins": "admins",
	"AdminVeto": "Veto",
	"AdminConfirm": "Confirm",
	"AdminForumAdministration": "Forum Administration",
	"AdminYou": "you!",
	"AdminUsers": "Users", 
	"AdminNoAdmins": "There are no admins",
	"AdminNoUsers": "There are no other users",
	"AdminNoPendingProposals": "There are no pending proposals",
	"AdminAddNewUser": "Add new user",
	"AdminAddNewUserQuestion": "Does someone wish attendence? You can ",
	"AdminStepDown": "Step down",
	"AdminStepDownExplanation": "If you want to stop being an admin, you can",
	"AdminViewPastActions": "View past actions in the",
	"ModerationLog": "moderation log",
	"AdminDemote": "Demote",
	"DeletedUser": "deleted user",
	"RemoveAccount": "remove account",
	"AdminMakeAdmin": "Make admin", 
	"Submit": "Submit",
	"AdminSelfConfirmationsHover": "a week must pass before self-confirmations are ok",
	"Proposal": "Proposal",
	"PendingProposals": "Pending Proposals",
	"AdminSelfProposalsBecomeValid": "Date self-proposals become valid",
	"AdminPendingExplanation": `Two admins are required for <i>making a user an admin</i>, <i>demoting an existing
															admin</i>, or <i>removing a user</i>. The first proposes the action, the second confirms
															(or vetos) it. If enough time elapses without a veto, the proposer may confirm their own
															proposal.`,
	
	"AdminAddUserExplanation": "Register a new user account. After registering the account you will be given a generated password and instructions to pass onto the user.",
	"AdminForumHasAdmins": "The forum currently has the following admins",
	"AdminOnlyLoggedInMayView": "Only logged in users may view the forum's admins.",
	"AdminPasswordSuccessInstructions": `Instructions: %s's password was set to: %s. After logging in, please change your password by going to /reset`,

	"ModLogNoActions": "there are no logged moderation actions",
	"ModLogExplanation": `This resource lists the moderation actions taken by the forum's administrators.`,
	"ModLogExplanationAdmin": `You are viewing this page as an admin, you will see slightly more details.`,
	"ModLogOnlyLoggedInMayView": "Only logged in users may view the moderation log.",

	"LoginNoAccount":       "Don't have an account yet? <a href='/register'>Register</a> one.",
	"LoginFailure":         "<b>Failed login attempt:</b> incorrect password, wrong username, or a non-existent user.",
	"LoginAlreadyLoggedIn": `You are already logged in. Would you like to <a href="/logout">log out</a>?`,

	"Username":                  "username",
	"Current":                   "current",
	"New":                       "new",
	"ChangePasswordDescription": "Use this page to change your password.",
	"Password":                  "password",
	"PasswordMin":               "Must be at least 9 characters long",
	"PasswordForgot":            "Forgot your password?",

	"Threads":   "threads",
	"ThreadNew": "new thread",
	"ThreadThe": "the thread",
	"Index":     "index",
	"GoBackToTheThread": "Go back to the thread",

	"ThreadCreate":        "Create thread",
	"Title":               "Title",
	"Content":             "Content",
	"Create":              "Create",
	"TextareaPlaceholder": "Tabula rasa",

	"PasswordReset":                   "reset password",
	"PasswordResetSuccess":            "Reset password—success!",
	"PasswordResetSuccessMessage":     "You reset your password!",
	"PasswordResetSuccessLinkMessage": "Give it a try and",

	"RegisterMessage":     "You already have an account (you are logged in with it).",
	"RegisterLinkMessage": "Visit the",
	"RegisterSuccess":     "registered successfully",

	"ErrUnaccepted":        "Unaccepted request",
	"ErrGeneric401":        "Unauthorized",
	"ErrGeneric401Message": "You do not have permissions to perform this action.",
	"ErrEdit404":           "Post not found",
	"ErrEdit404Message":    "This post cannot be found for editing",
	"ErrThread404":         "Thread not found",
	"ErrThread404Message":  "The thread does not exist (anymore?)",
	"ErrGeneric404":        "Page not found",
	"ErrGeneric404Message": "The visited page does not exist (anymore?). Error code %d.",

	"NewThreadMessage":            "Only members of this forum may create new threads",
	"NewThreadLinkMessage":        "If you are a member,",
	"NewThreadCreateError":        "Error creating thread",
	"NewThreadCreateErrorMessage": "There was a database error when creating the thread, apologies.",
	"PostEdit":                    "Post preview",

	"AriaPostMeta":          "Post meta",
	"AriaDeletePost":        "Delete this post",
	"AriaRespondIntoThread": "Respond into this thread",
	"PromptDeleteQuestion":  "Delete post for all posterity?",
	"Delete":                "delete",
	"Edit":                  "edit",
	"EditedAt":							 "edited at",
	"Post":                  "post",
	"Save":                  "Save",
	"Author":                "Author",
	"Responded":             "responded",
	"YourAnswer":            "Your answer",

	"AriaHome":       "Home",
	"ThreadStartNew": "Start a new thread",

	"RegisterHTMLMessage":               `You now have an account! Welcome. Visit the <a href="/">index</a> to read and reply to threads, or start a new one.`,

	"RegisterVerificationCode":              "Your verification code is",
	"RegisterVerificationInstructionsTitle": "Verification instructions",
	"RegisterVerificationLink":              "Verification link",
	"RegisterConductCodeBoxOne":             `I have refreshed my memory of the <a target="_blank" href="{{ .Data.Link }}">{{ .Data.Name }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo":             `Yes, I have actually <a target="_blank" href="{{ .Data.Link }}">read it</a>`,

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
	"Bottom":   "hoppa ner",

	"LoggedIn":    "inloggad",
	"NotLoggedIn": "Ej inloggad",
	"LogIn":       "logga in",
	"GoBack":      "Go back",

	"SortRecentPosts":   "nyast poster",
	"SortRecentThreads": "nyast trådar",

	"LoginNoAccount":       "Saknar du konto? <a href='/register'>Skapa</a> ett.",
	"LoginFailure":         "<b>Misslyckat inloggningsförsök:</b> inkorrekt lösenord, fel användernamn, eller obefintlig användare.",
	"LoginAlreadyLoggedIn": `Du är redan inloggad. Vill du <a href="/logout">logga ut</a>?`,

	"Username":                  "användarnamn",
	"Current":                   "nuvarande",
	"New":                       "nytt",
	"ChangePasswordDescription": "På den här sidan kan du ändra ditt lösenord.",
	"Password":                  "lösenord",
	"PasswordMin":               "Måste vara minst 9 karaktärer långt",
	"PasswordForgot":            "Glömt lösenordet?",

	"Threads":   "trådar",
	"ThreadNew": "ny tråd",
	"ThreadThe": "tråden",
	"Index":     "index",
	"GoBackToTheThread": "Go back to the thread",

	"ThreadCreate":        "Skapa en tråd",
	"Title":               "Titel",
	"Content":             "Innehåll",
	"Create":              "Skapa",
	"TextareaPlaceholder": "Tabula rasa",

	"PasswordReset":                   "nollställ lösenord",
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
	"PostEdit":                    "Post preview",

	"AriaPostMeta":          "Post meta",
	"AriaDeletePost":        "Delete this post",
	"AriaRespondIntoThread": "Respond into this thread",
	"PromptDeleteQuestion":  "Radera post för alltid?",
	"Delete":                "radera",
	"Edit":                  "redigera",
	"EditedAt":              "redigerat",
	"Post":                  "post",
	"Author":                "Författare",
	"Responded":             "svarade",
	"YourAnswer":            "Ditt svar",

	"AriaHome":       "Hem",
	"ThreadStartNew": "Starta ny tråd",

	"RegisterHTMLMessage":               `Du har nu ett konto! Välkommen. Besök <a href="/">trådindexet</a> för att läsa och svara på trådar, eller för att starta en ny.`,

	"RegisterVerificationCode":              "Din verifikationskod är",
	"RegisterVerificationInstructionsTitle": "Verification instructions",
	"RegisterVerificationLink":              "Verificationsnyckel",
	"RegisterConductCodeBoxOne":             `I have refreshed my memory of the <a target="_blank" href="{{ .Data.Link }}">{{ .Data.Name }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo":             `Yes, I have actually <a target="_blank" href="{{ .Data.Link }}">read it</a>`,

	"PasswordResetDescription":            "På denna sida går vi igenom ett par steg för att säkert nollställa ditt lösenord—utan att behöva ta till mejl!",
	"PasswordResetUsernameQuestion":       "För de första: hur löd användarnamnet?",
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
	"Bottom":   "bottom",

	"LoggedIn":    "logged in",
	"NotLoggedIn": "Not logged in",
	"LogIn":       "log in",
	"GoBack":      "Go back",

	"SortRecentPosts":   "recent posts",
	"SortRecentThreads": "most recent threads",

	"LoginNoAccount":       "¿No tienes una cuenta? <a href='/register'>Registra</a> una. ",
	"LoginFailure":         "<b>Failed login attempt:</b> incorrect password, wrong username, or a non-existent user.",
	"LoginAlreadyLoggedIn": `You are already logged in. Would you like to <a href="/logout">log out</a>?`,

	"Username":                  "usuarie",
	"Current":                   "current",
	"New":                       "new",
	"ChangePasswordDescription": "Use this page to change your password.",
	"Password":                  "contraseña",
	"PasswordMin":               "Debe tener por lo menos 9 caracteres.",
	"PasswordForgot":            "Olvidaste tu contraseña?",

	"Threads":   "threads",
	"ThreadNew": "new thread",
	"ThreadThe": "the thread",
	"Index":     "index",
	"GoBackToTheThread": "Go back to the thread",

	"ThreadCreate":        "Create thread",
	"Title":               "Title",
	"Content":             "Content",
	"Create":              "Create",
	"TextareaPlaceholder": "Tabula rasa",

	"PasswordReset":                   "reset password",
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
	"PostEdit":                    "Post preview",
	"ThreadStartNew":              "Start a new thread",

	"AriaPostMeta":          "Post meta",
	"AriaDeletePost":        "Delete this post",
	"AriaRespondIntoThread": "Respond into this thread",
	"AriaHome":              "Home",
	"PromptDeleteQuestion":  "Delete post for all posterity?",
	"Delete":                "delete",
	"Edit":                  "editar",
	"EditedAt":              "editado a las",
	"Post":                  "post",
	"Save":                  "Save",
	"Author":                "Author",
	"Responded":             "responded",
	"YourAnswer":            "Your answer",

	"RegisterHTMLMessage":               `You now have an account! Welcome. Visit the <a href="/">index</a> to read and reply to threads, or start a new one.`,

	"RegisterVerificationCode":              "Your verification code is",
	"RegisterVerificationInstructionsTitle": "Verification instructions",
	"RegisterVerificationLink":              "Verification link",
	"RegisterConductCodeBoxOne":             `I have refreshed my memory of the <a target="_blank" href="{{ .Data.Link }}">{{ .Data.Name }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo":             `Yes, I have actually <a target="_blank" href="{{ .Data.Link }}">read it</a>`,

	"PasswordResetDescription":            "On this page we'll go through a few steps to securely reset your password—without resorting to any emails!",
	"PasswordResetUsernameQuestion":       "First up: what was your username?",
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
