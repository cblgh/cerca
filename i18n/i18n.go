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
	"SortRecentThreads": "recent threads",

	"modlogCreateInvites":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> created a batch of invites`,
	"modlogDeleteInvites":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> deleted a batch of invites`,
	"modlogResetPassword":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> reset a user's password`,
	"modlogResetPasswordAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> reset <b> {{ .Data.RecipientUsername}}</b>'s password`,
	"modlogRemoveUser":         `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> removed a user's account`,
	"modlogMakeAdmin":          `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> made <b> {{ .Data.RecipientUsername}}</b> an admin`,
	"modlogAddUser":            `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> manually registered an account for a new user`,
	"modlogAddUserAdmin":       `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> manually registered an account for <b> {{ .Data.RecipientUsername }}</b>`,
	"modlogDemoteAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> demoted <b> 
	{{ if eq .Data.ActingUsername .Data.RecipientUsername }} themselves 
	{{ else }} {{ .Data.RecipientUsername}} {{ end }}</b> from admin back to normal user`,
	"modlogXProposedY":          `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> proposed: {{ .Data.Action }}`,
	"modlogProposalMakeAdmin":   `Make <b> {{ .Data.RecipientUsername}}</b> admin`,
	"modlogProposalDemoteAdmin": `Demote <b> {{ .Data.RecipientUsername}}</b> from role admin`,
	"modlogProposalRemoveUser":  `Remove user <b> {{ .Data.RecipientUsername }} </b>`,
	"modlogConfirm":             "{{ .Data.Action }} <i>confirmed by {{ .Data.ActingUsername }}</i>",
	"modlogVeto":                "<s>{{ .Data.Action }}</s> <i>vetoed by {{ .Data.ActingUsername }}</i>",

	"Admins":                        "admins",
	"AdminVeto":                     "Veto",
	"AdminConfirm":                  "Confirm",
	"AdminForumAdministration":      "Forum Administration",
	"AdminYou":                      "you!",
	"AdminUsers":                    "Users",
	"AdminNoAdmins":                 "There are no admins",
	"AdminNoUsers":                  "There are no other users",
	"AdminNoPendingProposals":       "There are no pending proposals",
	"AdminAddNewUser":               "Add new user",
	"AdminAddNewUserQuestion":       "Does someone wish attendence? You can ",
	"AdminStepDown":                 "Step down",
	"AdminStepDownExplanation":      "If you want to stop being an admin, you can",
	"AdminViewPastActions":          "View past actions in the",
	"ModerationLog":                 "moderation log",
	"AdminDemote":                   "Demote",
	"DeletedUser":                   "deleted user",
	"RemoveAccount":                 "remove account",
	"AdminMakeAdmin":                "Make admin",
	"Submit":                        "Submit",
	"AdminSelfConfirmationsHover":   "a week must pass before self-confirmations are ok",
	"Proposal":                      "Proposal",
	"PendingProposals":              "Pending Proposals",
	"AdminSelfProposalsBecomeValid": "Date self-proposals become valid",
	"AdminPendingExplanation": `Two admins are required for <i>making a user an admin</i>, <i>demoting an existing
															admin</i>, or <i>removing a user</i>. The first proposes the action, the second confirms
															(or vetos) it. If enough time elapses without a veto, the proposer may confirm their own
															proposal.`,

	"AdminAddUserExplanation":          "Register a new user account. After registering the account you will be given a generated password and instructions to pass onto the user.",
	"AdminForumHasAdmins":              "The forum currently has the following admins",
	"AdminOnlyLoggedInMayView":         "Only logged in users may view the forum's admins.",
	"AdminPasswordSuccessInstructions": `Instructions: %s's password was set to: %s. After logging in, please change your password by going to /reset`,

	"ModLogNoActions":           "there are no logged moderation actions",
	"ModLogExplanation":         `This resource lists the moderation actions taken by the forum's administrators.`,
	"ModLogExplanationAdmin":    `You are viewing this page as an admin, you will see slightly more details.`,
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

	"Posts":             "posts",
	"Threads":           "threads",
	"ThreadNew":         "new thread",
	"ThreadThe":         "the thread",
	"Index":             "index",
	"GoBackToTheThread": "Go back to the thread",
	"ThreadsViewEmpty":  "There are currently no threads.",

	"ThreadCreate":        "Create thread",
	"Title":               "Title",
	"Content":             "Content",
	"Private":             "Private",
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
	"PostPrivate":                 "This is a private thread, only logged-in users can see it and read its posts.",

	"AriaPostMeta":          "Post meta",
	"AriaDeletePost":        "Delete this post",
	"AriaRespondIntoThread": "Respond into this thread",
	"PromptDeleteQuestion":  "Delete post for all posterity?",
	"Delete":                "delete",
	"Edit":                  "edit",
	"EditedAt":              "edited at",
	"Post":                  "post",
	"Save":                  "Save",
	"Author":                "Author",
	"Responded":             "responded",
	"YourAnswer":            "Your answer",

	"AriaHome":       "Home",
	"ThreadStartNew": "Start a new thread",

	"RegisterHTMLMessage": `You now have an account! Welcome. Visit the <a href="/">index</a> to read and reply to threads, or start a new one.`,

	"RegisterInviteInstructionsTitle": "How to get an invite code",
	"RegisterConductCodeBoxOne":       `I have refreshed my memory of the <a target="_blank" href="{{ .Data.Link }}">{{ .Data.Name }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo":       `Yes, I have actually <a target="_blank" href="{{ .Data.Link }}">read it</a>`,

	"NewPassword":    "new password",
	"ChangePassword": "change password",
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

	/* begin 2025-03-26: to translate to swedish */
	"modlogCreateInvites":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> created a batch of invites`,
	"modlogDeleteInvites":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> deleted a batch of invites`,
	"modlogResetPassword":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> reset a user's password`,
	"modlogResetPasswordAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> reset <b> {{ .Data.RecipientUsername}}</b>'s password`,
	"modlogRemoveUser":         `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> removed a user's account`,
	"modlogMakeAdmin":          `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> made <b> {{ .Data.RecipientUsername}}</b> an admin`,
	"modlogAddUser":            `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> manually registered an account for a new user`,
	"modlogAddUserAdmin":       `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> manually registered an account for <b> {{ .Data.RecipientUsername }}</b>`,
	"modlogDemoteAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> demoted <b> 
	{{ if eq .Data.ActingUsername .Data.RecipientUsername }} themselves 
	{{ else }} {{ .Data.RecipientUsername}} {{ end }}</b> from admin back to normal user`,
	"modlogXProposedY":          `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> proposed: {{ .Data.Action }}`,
	"modlogProposalMakeAdmin":   `Make <b> {{ .Data.RecipientUsername}}</b> admin`,
	"modlogProposalDemoteAdmin": `Demote <b> {{ .Data.RecipientUsername}}</b> from role admin`,
	"modlogProposalRemoveUser":  `Remove user <b> {{ .Data.RecipientUsername }} </b>`,
	"modlogConfirm":             "{{ .Data.Action }} <i>confirmed by {{ .Data.ActingUsername }}</i>",
	"modlogVeto":                "<s>{{ .Data.Action }}</s> <i>vetoed by {{ .Data.ActingUsername }}</i>",

	"Admins":                        "admins",
	"AdminVeto":                     "Veto",
	"AdminConfirm":                  "Confirm",
	"AdminForumAdministration":      "Forum Administration",
	"AdminYou":                      "you!",
	"AdminUsers":                    "Users",
	"AdminNoAdmins":                 "There are no admins",
	"AdminNoUsers":                  "There are no other users",
	"AdminNoPendingProposals":       "There are no pending proposals",
	"AdminAddNewUser":               "Add new user",
	"AdminAddNewUserQuestion":       "Does someone wish attendence? You can ",
	"AdminStepDown":                 "Step down",
	"AdminStepDownExplanation":      "If you want to stop being an admin, you can",
	"AdminViewPastActions":          "View past actions in the",
	"ModerationLog":                 "moderation log",
	"AdminDemote":                   "Demote",
	"DeletedUser":                   "deleted user",
	"RemoveAccount":                 "remove account",
	"AdminMakeAdmin":                "Make admin",
	"Submit":                        "Submit",
	"AdminSelfConfirmationsHover":   "a week must pass before self-confirmations are ok",
	"Proposal":                      "Proposal",
	"PendingProposals":              "Pending Proposals",
	"AdminSelfProposalsBecomeValid": "Date self-proposals become valid",
	"AdminPendingExplanation": `Two admins are required for <i>making a user an admin</i>, <i>demoting an existing
															admin</i>, or <i>removing a user</i>. The first proposes the action, the second confirms
															(or vetos) it. If enough time elapses without a veto, the proposer may confirm their own
															proposal.`,

	"AdminAddUserExplanation":          "Register a new user account. After registering the account you will be given a generated password and instructions to pass onto the user.",
	"AdminForumHasAdmins":              "The forum currently has the following admins",
	"AdminOnlyLoggedInMayView":         "Only logged in users may view the forum's admins.",
	"AdminPasswordSuccessInstructions": `Instructions: %s's password was set to: %s. After logging in, please change your password by going to /reset`,

	"ModLogNoActions":           "there are no logged moderation actions",
	"ModLogExplanation":         `This resource lists the moderation actions taken by the forum's administrators.`,
	"ModLogExplanationAdmin":    `You are viewing this page as an admin, you will see slightly more details.`,
	"ModLogOnlyLoggedInMayView": "Only logged in users may view the moderation log.",

	/* end 2025-03-26: to translate to swedish */

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

	"Posts":             "poster",
	"Threads":           "trådar",
	"ThreadNew":         "ny tråd",
	"ThreadThe":         "tråden",
	"Index":             "index",
	"GoBackToTheThread": "Gå tillbaka till tråden",
	"ThreadsViewEmpty":  "Det finns för närvarande inga trådar",

	"ThreadCreate":        "Skapa en tråd",
	"Title":               "Titel",
	"Content":             "Innehåll",
	"Create":              "Skapa",
	"Private":             "Privat",
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

	"RegisterHTMLMessage": `Du har nu ett konto! Välkommen. Besök <a href="/">trådindexet</a> för att läsa och svara på trådar, eller för att starta en ny.`,

	"RegisterInviteInstructionsTitle": "Instruktioner för invitationskod",
	"RegisterConductCodeBoxOne":       `I have refreshed my memory of the <a target="_blank" href="{{ .Data.Link }}">{{ .Data.Name }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo":       `Yes, I have actually <a target="_blank" href="{{ .Data.Link }}">read it</a>`,

	"PasswordResetDescription":      "På denna sida går vi igenom ett par steg för att säkert nollställa ditt lösenord—utan att behöva ta till mejl!",
	"PasswordResetUsernameQuestion": "För de första: hur löd användarnamnet?",
	"NewPassword":                   "nytt lösenord",
	"ChangePassword":                "ändra lösenord",
}

var Danish = map[string]string{
	"About":    "om",
	"Login":    "log ind",
	"Logout":   "log ud",
	"Sort":     "sorter",
	"Enter":    "Indsend",
	"Register": "registrer",
	"Bottom":   "gå til bunden",

	"LoggedIn":    "logget ind",
	"NotLoggedIn": "ikke logget ind",
	"LogIn":       "log ind",
	"GoBack":      "gå tilbage",

	"SortRecentPosts":   "sorter efter nyeste opslag",
	"SortRecentThreads": "sorter efter nyeste tråd",

	"modlogCreateInvites":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> lavede en række af invitationer`,
	"modlogDeleteInvites":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> slettede en række af invitationer`,
	"modlogResetPassword":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> nulstillede en brugers password`,
	"modlogResetPasswordAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> nulstillede <b> {{ .Data.RecipientUsername}}</b>'s password`,
	"modlogRemoveUser":         `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> fjernede en brugers konto`,
	"modlogMakeAdmin":          `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> gjorde <b> {{ .Data.RecipientUsername}}</b> til admin`,
	"modlogAddUser":            `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> registerede manuels en konto til en bruger`,
	"modlogAddUserAdmin":       `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> registerede manuelt en konto til <b> {{ .Data.RecipientUsername }}</b>`,
	"modlogDemoteAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> fratog <b> 
	{{ if eq .Data.ActingUsername .Data.RecipientUsername }} dem selv deres admin status
	{{ else }} {{ .Data.RecipientUsername}} {{ end }}</b> admin rang og gjorde dem til en normal bruger`,
	"modlogXProposedY":          `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> foreslog : {{ .Data.Action }}`,
	"modlogProposalMakeAdmin":   `Gør <b> {{ .Data.RecipientUsername}}</b> til admin`,
	"modlogProposalDemoteAdmin": `Fratag <b> {{ .Data.RecipientUsername}}</b>'s admin status`,
	"modlogProposalRemoveUser":  `Fjern bruger <b> {{ .Data.RecipientUsername }} </b>`,
	"modlogConfirm":             "{{ .Data.Action }} <i>blev gennemført af {{ .Data.ActingUsername }}</i>",
	"modlogVeto":                "<s>{{ .Data.Action }}</s> <i>blev vetoet af {{ .Data.ActingUsername }}</i>",

	"Admins":                        "admins",
	"AdminVeto":                     "Veto",
	"AdminConfirm":                  "Bekræft",
	"AdminForumAdministration":      "Forum Administration",
	"AdminYou":                      "dig!",
	"AdminUsers":                    "Brugere",
	"AdminNoAdmins":                 "Der er ingen admins",
	"AdminNoUsers":                  "Der er ingen andre admins",
	"AdminNoPendingProposals":       "Der er ingen afventende forslag",
	"AdminAddNewUser":               "Tilføj en ny bruger",
	"AdminAddNewUserQuestion":       "Er der nogen som ønsker at deltage i forummet? Du kan ",
	"AdminStepDown":                 "Træd fra",
	"AdminStepDownExplanation":      "Hvis du gerne vil stoppe med at være en admin, så kan du",
	"AdminViewPastActions":          "Se tidligere handinger i ",
	"ModerationLog":                 "moderations log",
	"AdminDemote":                   "Fratag status",
	"DeletedUser":                   "slettede bruger",
	"RemoveAccount":                 "fjern konto",
	"AdminMakeAdmin":                "Gør til admin",
	"Submit":                        "Indsend",
	"AdminSelfConfirmationsHover":   "der skal gå uge før at admin selv-bekræftelse er ok",
	"Proposal":                      "Forslag",
	"PendingProposals":              "Afventende forslag",
	"AdminSelfProposalsBecomeValid": "Dato selvindsendte forslag træder i kræft",
	"AdminPendingExplanation": `To admins er krævet for <i>at gøre en bruger til admin</i>, <i>fratage en eksisterende 
															admin sin admin status</i>, eller <i>for at fjerne en bruger</i>. Den første admin foreslår en administrations handling, den anden admin bekræfter
															(eller vetoer) handlingen. Hvis der går længe nok uden et veto, kan den første admin bekræfte og gennemføre deres eget
															foreslag.`,

	"AdminAddUserExplanation":          "Registrer en ny bruger konto. Efter at have registreret den nye konto vil du blive givet et generet password og instruktioner som du kan videre sende til brugeren.",
	"AdminForumHasAdmins":              "Forummet har i øjeblikket følgende admins.",
	"AdminOnlyLoggedInMayView":         "Kun brugere som er logget ind, kan se forummets admins.",
	"AdminPasswordSuccessInstructions": `Instruktioner: %s's password blev sat til: %s. Efter at du har logget ind, ændre venligst dit password ved at gå til /reset`,

	"ModLogNoActions":           "der er ingen logged moderations handlinger",
	"ModLogExplanation":         `Denne liste viser de moderations handlinger som forummet's adminis har taget.`,
	"ModLogExplanationAdmin":    `Du er logged ind på denne side som en admin, du vil se flere detaljer end normalt. `,
	"ModLogOnlyLoggedInMayView": "Kun brugere som er logget ind kan se moderations loggen.",

	/* end 2025-03-26: to translate to swedish */

	"LoginNoAccount":       "Mangler du en konto? <a href='/register'>Ansøg om en her</a>.",
	"LoginFailure":         "<b>Mislykket login forsøg:</b> forkert password, forkert brugernavn, eller ikke-eksisterende bruger.",
	"LoginAlreadyLoggedIn": `Du er allerede logget ind. Vil du <a href="/logout">logge ud</a>?`,

	"Username":                  "brugernavn",
	"Current":                   "nuværende",
	"New":                       "ny",
	"ChangePasswordDescription": "På den her side kan du ændre dit password.",
	"Password":                  "password",
	"PasswordMin":               "Skal være mindst 9 tegn langt",
	"PasswordForgot":            "Glemt dit password?",

	"Posts":             "opslag",
	"Threads":           "tråde",
	"ThreadNew":         "ny tråd",
	"ThreadThe":         "tråden",
	"Index":             "index",
	"GoBackToTheThread": "Gå tilbage til tråden",
	"ThreadsViewEmpty":  "Der findes iøjeblikket ingen tråde",

	"ThreadCreate":        "Lav en tråd",
	"Title":               "Titel",
	"Content":             "Indhold",
	"Create":              "Udgiv opslaget",
	"Private":             "Privat",
	"TextareaPlaceholder": "Tabula rasa",

	"PasswordReset":                   "nulstil password",
	"PasswordResetSuccess":            "Nulstilning af password—lykkedes!",
	"PasswordResetSuccessMessage":     "Du har nulstillet dit password!",
	"PasswordResetSuccessLinkMessage": "Giv det et forsøg",

	"RegisterMessage":     "Du har allerede en konto (du er logged in med den).",
	"RegisterLinkMessage": "Besøg",
	"RegisterSuccess":     "konto registeret successfuld",

	"ErrUnaccepted":        "Fejl, anmodningen ikke accepteret",
	"ErrThread404":         "Tråd ej fundet",
	"ErrThread404Message":  "Denne tråde findes ikke (måske ikke længere?)",
	"ErrGeneric404":        "Siden kunne ikke findes",
	"ErrGeneric404Message": "Den besøgte side findes ikke (længere?). Fejlkode %d.",

	"NewThreadMessage":            "Enbart medlemmarna av detta forum får skapa nya trådar",
	"NewThreadLinkMessage":        "Om du er et medlem,",
	"NewThreadCreateError":        "Fejl opstod ved oprettelsen af tråden",
	"NewThreadCreateErrorMessage": "Det opstod en datafejl mens tråden blev lavet, årsag.",
	"PostEdit":                    "Post preview",

	"AriaPostMeta":          "Post meta",
	"AriaDeletePost":        "Slet dette opslag",
	"AriaRespondIntoThread": "Skriv tilbage i denne tråd",
	"PromptDeleteQuestion":  "Slet opslaget for altid?",
	"Delete":                "slet",
	"Edit":                  "rediger",
	"EditedAt":              "redigere ved",
	"Post":                  "indsend",
	"Author":                "Forfatter",
	"Responded":             "svarede",
	"YourAnswer":            "Dit svar",

	"AriaHome":       "Hjem",
	"ThreadStartNew": "Start en ny tråd",

	"RegisterHTMLMessage": `Du har nu en konto! Velkommen. Besøg <a href="/">tråd indexet</a> for at læse og svar på tråde, eller for at start en ny.`,

	"RegisterInviteInstructionsTitle": "Instruktioner til at få en invitationskode",
	"RegisterConductCodeBoxOne":       `Jeg har læst <a target="_blank" href="{{ .Data.Link }}">{{ .Data.Name }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo":       `Ja, jeg har faktisk læst <a target="_blank" href="{{ .Data.Link }}">læst den</a>`,

	"PasswordResetDescription":      "På denne side går vi igennem et par ting for at sikre nulstillingen af dit password—uden at behøve at bruge mail!",
	"PasswordResetUsernameQuestion": "For det første: hvad er dit brugernavn?",
	"NewPassword":                   "nyt password",
	"ChangePassword":                "ændre password",
}


var EspanolLATAM = map[string]string{
	"About":    "acerca de",
	"Login":    "acceder",
	"Logout":   "salir",
	"Sort":     "orden",
	"Register": "registro",
	"Enter":    "entrar",
	"Bottom":   "fin(al)",

	"LoggedIn":    "¡Accediste!",
	"NotLoggedIn": "No accediste :(",
	"LogIn":       "Accede",
	"GoBack":      "Regresa",

	"SortRecentPosts":   "publicaciones recientes",
	"SortRecentThreads": "hilos recientes",

	"modlogCreateInvites":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> creó un conjunto de invitaciones`,
	"modlogDeleteInvites":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> borró un conjunto de invitaciones`,
	"modlogResetPassword":      `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> reseteó una contraseña de usuarie`,
	"modlogResetPasswordAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> reseteó la contraseña de <b> {{ .Data.RecipientUsername}}</b>`,
	"modlogRemoveUser":         `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> removió una cuenta de usuarie`,
	"modlogMakeAdmin":          `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> hizo a <b> {{ .Data.RecipientUsername}}</b> une admin`,
	"modlogAddUser":            `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> manualmente registró una cuenta para une usuarie nuevx`,
	"modlogAddUserAdmin":       `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> manualmente registró una cuenta para <b> {{ .Data.RecipientUsername }}</b>`,
	"modlogDemoteAdmin": `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> destituyo a <b>
	{{ if eq .Data.ActingUsername .Data.RecipientUsername }}
	{{ else }} {{ .Data.RecipientUsername}} {{ end }}</b> (elle misme) de admin a usuarie normal`,
	"modlogXProposedY":          `<code>{{ .Data.Time }}</code> <b>{{ .Data.ActingUsername }}</b> propuso: {{ .Data.Action }}`,
	"modlogProposalMakeAdmin":   `Hacer a <b> {{ .Data.RecipientUsername}}</b> une admin`,
	"modlogProposalDemoteAdmin": `Deponer a <b> {{ .Data.RecipientUsername}}</b> del rol admin`,
	"modlogProposalRemoveUser":  `Remover a <b> {{ .Data.RecipientUsername }} </b>`,
	"modlogConfirm":             "{{ .Data.Action }} <i>Confirmado por {{ .Data.ActingUsername }}</i>",
	"modlogVeto":                "<s>{{ .Data.Action }}</s> <i>Rechazado por {{ .Data.ActingUsername }}</i>",

	"Admins":                        "Administradorxs",
	"AdminVeto":                     "Rechaza",
	"AdminConfirm":                  "Confirma",
	"AdminForumAdministration":      "Panel de Admin",
	"AdminYou":                      "¡Tu!",
	"AdminUsers":                    "Usuaries",
	"AdminNoAdmins":                 "No hay admins",
	"AdminNoUsers":                  "No hay otres usuaries",
	"AdminNoPendingProposals":       "No hay propuestas pendientes",
	"AdminAddNewUser":               "Crea usuarie nuevx",
	"AdminAddNewUserQuestion":       "¿Alguien desea formar parte del foro?",
	"AdminStepDown":                 "dimitir",
	"AdminStepDownExplanation":      "Si quieres dejar de ser admin puedes",
	"AdminViewPastActions":          "Ver acciones pasadas en el",
	"ModerationLog":                 "Registro de moderación",
	"AdminDemote":                   "Destituir",
	"DeletedUser":                   "Usuarie removidx",
	"RemoveAccount":                 "remover cuenta",
	"AdminMakeAdmin":                "Hacer admin",
	"Submit":                        "Activar",
	"AdminSelfConfirmationsHover":   "Una semana debe pasar antes de que las autoconfirmaciones sean válidas",
	"Proposal":                      "Propuesta",
	"PendingProposals":              "Propuestas admin pendientes",
	"AdminSelfProposalsBecomeValid": "Fecha de autovalidación",
	"AdminPendingExplanation": `Dos admins se necesitan para <i>hacer une usuarie como admin</i>, <i>deponer une admin
															existente</i>, o <i>remover une usuarie</i>. Le primere propone la acción, le segunde confirma
															(o veta). Si suficiente tiempo pasa sin un veto quien propone puede confirmar su propia
															propuesta.`,

	"AdminAddUserExplanation":          "Registra una cuenta nueva de usuarie. Luego de registrar la cuenta vas a recibir una contraseña e instrucciones para comptartirle a le usuarie.",
	"AdminForumHasAdmins":              "Este foro tiene les siguientes admins",
	"AdminOnlyLoggedInMayView":         "Solo les usuaries con sesión iniciada pueden ver les admins del foro.",
	"AdminPasswordSuccessInstructions": `Instrucciones: La contraseña de %s fue establecida a %s. Luego de iniciar sesión puede cambiar su contraseña entrando a la configuración de su cuenta.`,

	"ModLogNoActions":           "No hay registro de acciones de moderadore",
	"ModLogExplanation":         `Este recurso hace lista de las acciones de moderadore hechas por admins de este foro.`,
	"ModLogExplanationAdmin":    `Como estas viendo esta página como admin vas a poder ver más detalles.`,
	"ModLogOnlyLoggedInMayView": "Solo usuaries con sesión iniciada van a poder ver el registro de moderación.",

	"LoginNoAccount":       "¿No tienes una cuenta? <a href='/register'>Regístrate</a>.",
	"LoginFailure":         "<b>Falló el intento de acceder:</b> contraseña incorrecta, usuario equivocado o no existe el nombre de usuario.",
	"LoginAlreadyLoggedIn": `Ya estás dentro. ¿Quisieras <a href="/logout">salir</a>?`,

	"Username":                  "Usuarie",
	"Current":                   "Actual",
	"New":                       "Nuevo",
	"ChangePasswordDescription": "Usa esta página para cambiar la contraseña.",
	"Password":                  "Contraseña",
	"PasswordMin":               "Debe tener por lo menos 9 caracteres.",
	"PasswordForgot":            "¿Olvidaste tu contraseña?",

	"Posts":             "publicaciones",
	"Threads":           "hilos",
	"ThreadNew":         "Nuevo hilo",
	"ThreadThe":         "El hilo",
	"Index":             "Inicio",
	"GoBackToTheThread": "Regresa al hilo",
	"ThreadsViewEmpty":  "Actualmente no hay hilos.",

	"ThreadCreate":        "Crea un hilo",
	"Title":               "Título",
	"Content":             "Contenido",
	"Private":	       "Privado",
	"Create":              "Crear",
	"TextareaPlaceholder": "Escribe aquí",

	"PasswordReset":                   "Cambia la contraseña",
	"PasswordResetSuccess":            "¡Contraseña cambiada exitosamente!",
	"PasswordResetSuccessMessage":     "¡Cambiaste la contraseña!",
	"PasswordResetSuccessLinkMessage": "Inténtalo y",

	"RegisterMessage":     "Ya tienes una cuenta (ya iniciaste sesión).",
	"RegisterLinkMessage": "Visita",
	"RegisterSuccess":     "Registro exitoso",

	"ErrUnaccepted":        "Solicitud no aceptada",
	"ErrGeneric401":        "No autorizado",
	"ErrGeneric401Message": "No tienes permiso para hacer esta acción.",
	"ErrEdit404":           "Publicación no encontrada",
	"ErrEdit404Message":    "Publicación no fue encontrada y no se puede editar",
	"ErrThread404":         "Hilo no encontrado",
	"ErrThread404Message":  "El hilo no existe (¿más?)",
	"ErrGeneric404":        "Página no encontrada",
	"ErrGeneric404Message": "La página visitada no existe (¿más?). Código de error %d.",

	"NewThreadMessage":            "Solo miembros de este foro pueden crear hilos nuevos",
	"NewThreadLinkMessage":        "Si eres miembro,",
	"NewThreadCreateError":        "Error creando hilo",
	"NewThreadCreateErrorMessage": "Disculpa, hubo un error en la base de datos cuando se creó el hilo.",
	"PostEdit":                    "Vista previa",
	"PostPrivate":                 "Este es un hilo privado, solo personas con sesión iniciada pueden ver y leer su contenido.",

	"AriaPostMeta":          "Post meta",
	"AriaDeletePost":        "Borrar este post",
	"AriaRespondIntoThread": "Responde dentro de este hilo",
	"PromptDeleteQuestion":  "¿Borrar esta publicación para la posteridad?",
	"Delete":                "borrar",
	"Edit":                  "editar",
	"EditedAt":              "editado a las",
	"Post":                  "publicar",
	"Save":                  "guardar",
	"Author":                "autore",
	"Responded":             "respondió",
	"YourAnswer":            "Tu respuesta",

	"AriaHome":       	 "Inicio",
	"ThreadStartNew":        "Empieza un hilo nuevo",

	"RegisterHTMLMessage": `¡Ahora tienes una cuenta! Bienvenide. Visita el <a href="/">inicio</a> para leer o responder los hilos, o empieza uno nuevo.`,

	"RegisterInviteInstructionsTitle": "Cómo obtener un código de invitación",
	"RegisterConductCodeBoxOne":       `He refrescado mi memoria del <a target="_blank" href="{{ .Data.Link }}">{{ .Data.Name }} Code of Conduct</a>`,
	"RegisterConductCodeBoxTwo":       `Sí, realmente lo <a target="_blank" href="{{ .Data.Link }}">he leído</a>.`,

	"NewPassword":                   "Nueva contraseña",
	"ChangePassword":                "Cambiar contraseña",
	}

var translations = map[string]map[string]string{
	"English":         English,
	"EspañolLATAM":   EspanolLATAM,
	"Swedish":         Swedish,
	"Danish":          Danish,
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
// 	  fmt.Println(tr.TranslateWithData("LoginDescription", General{"Merveilles", "https://merveill.es"}))
