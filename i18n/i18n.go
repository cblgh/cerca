package i18n

import (
	"cerca/util"
	"text/template"
	"strings"
	"log"
	"fmt"
)

var English = map[string]string{
	"About": "about",
	"Login": "login",
	"Logout": "logout",
	"Sort": "sort",
	"SortPostsRecent": "recent posts",
	"SortThreadsRecent": "most recent threads",
	"LoginDescription": "This forum is for the <a href='{{ .CommunityLink }}'>{{.CommunityName}}</a> community.",
	"LoginNoAccount": "Don't have an account yet? <a href='/register'>Register</a> one.",
	"Username": "username",
	"Password": "password",
	"PasswordMin": "Must be at least 9 characters long",
	"PasswordForgot": "Forgot your password?",
	"Enter": "enter",
}

var EspanolMexicano = map[string]string{
	"About": "acerca de",
	"Login": "loguearse",
	"Logout": "logout",
	"Sort": "sort",
	"SortPostsRecent": "recent posts",
	"SortThreadsRecent": "most recent threads",
	"LoginDescription": "Este foro es principalmente para las personas de la comunidad <a href='{{.CommunityLink}}>{{.CommunityName}}</a>.",
	"LoginNoAccount": "¿No tienes una cuenta? <a href='/register'>Registra</a> una. ",
	"Username": "usuarie",
	"Password": "contraseña",
	"PasswordMin": "Debe tener por lo menos 9 caracteres.",
	"PasswordForgot": "Olvidaste tu contraseña?",
	"Enter": "enter",
}

var translations = map[string]map[string]string{
	"English": English,
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

func main() {
	tr := Init("EnglishSwedish")
	fmt.Println(tr.Translate("LoginNoAccount"))
	fmt.Println(tr.TranslateWithData("LoginDescription", Community{"Merveilles", "https://merveill.es"}))
}
