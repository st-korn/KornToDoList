package main

import (
	"net/http" // for HTTP-server
	"os" // to get OS environment variables
	"text/template" // for use HTML-page templates
	"io/ioutil" // for read text files from the server
	"encoding/json" // to parse JSON (language strings-tables)
	"net/mail" // to generate emails
	"encoding/base64" // to use UTF-8 in emails bodys
	"mime" // to use UTF-8 in emails headers
	"net/smtp" // to send emails
	"fmt" // to generate emails body
	"bytes" // string buffer to use string-templates
	"time" // to define MongoDB time type for structure elements
	"golang.org/x/text/language" // to detect user-perferred language
	"golang.org/x/text/language/display" // to output national names of languages
	"gopkg.in/mgo.v2" // to connect to MongoDB
)

// Global envirnment variables
var URI string //URI MongoDB database
var DB string //MongoDB database name
var MAILHOST string //smtp-server hostname
var MAILPORT string //smtp-server port number (usually 25)
var MAILLOGIN string //login of smtp-server account
var MAILPASSWORD string //password of smtp-server account
var MAILFROM string //E-mail address, from which will be sent emails
var SERVERHTTPADDRESS string //HTTP-address for web-server with http or https prefix, and without any path: (eg. http://127.0.0.1:9000 or https://todo.works)
var LISTENPORT string //port on which the web server listens

const DefaultSessionLifetimeDays = 365 // default user-session UUID lifetime in database (days)

// ===========================================================================================================================
// MongoDB session working
// ===========================================================================================================================

var SESSION *mgo.Session // global MongoDB session variable

// Always return active MongoDB session object
// If necessary, reconnect to the database
func GetMongoDBSession() *mgo.Session {
	if SESSION == nil {
		var err error
		SESSION, err = mgo.Dial(URI)
		if err != nil { panic(err) }
	}
	return SESSION.Clone()
}

// ===========================================================================================================================
// Return JSON in response to a http-request
// ===========================================================================================================================
// IN: 
func ReturnJSON(res http.ResponseWriter, structJSON interface{}) {
    // Create json response from struct
    resJSON, err := json.Marshal(structJSON)
    if err != nil { panic(err) }
    res.Header().Set("Content-type", "application/json; charset=utf-8")
    res.Write(resJSON)

}

// ===========================================================================================================================
// Language defenition and detection
// ===========================================================================================================================

// All supported languages
var SupportedLangs = []language.Tag{
    language.English,   // The first language is used as default
    language.Russian}

// Language structure: english_name_of_language and national_name_of_language
type typeLang struct {
	EnglishName string
	NationalName string
}

// Map of national labels for selected language
type typeLabels map[string]string

// Internal function: detect language from http-request, and load labels for detected language
// IN:
//		req *http.Request // http-request
// OUT: 
//		langTag language.Tag // detected language
//		langEnglishName string // english name of detected language
//		labels typeLabels // labels of these language
func DetectLanguageAndLoadLabels(req *http.Request) (langTag language.Tag, langEnglishName string, labels typeLabels) {

	// Start detect user-language from cookie
	var langCookieEnglishName string
	var langTagCode string
	langCookie, err := req.Cookie("User-Language")
	if err == nil { langCookieEnglishName = langCookie.Value }

    // Finduser-language from supported languages list
	for _, tag := range SupportedLangs {
		if display.English.Tags().Name(tag) == langCookieEnglishName {langTagCode = tag.String()}
	}

	// Finish detect user-language
    accept := req.Header.Get("Accept-Language")
    matcher := language.NewMatcher(SupportedLangs)
    langTag, _ = language.MatchStrings(matcher, langTagCode, accept)

    langEnglishName = display.English.Tags().Name(langTag);

	// Load strings-table for user-language
	jsonFile, err := os.Open("templates/"+langEnglishName+".json")
	if err != nil { panic(err) }
	defer jsonFile.Close()
	jsonText, err := ioutil.ReadAll(jsonFile)
	if err != nil { panic(err) }
	json.Unmarshal(jsonText, &labels)

	return langTag, langEnglishName, labels
}

// ===========================================================================================================================
// E-Mail sending
// ===========================================================================================================================

// Internal function: try to sends a letter from the system to the specified mailbox
// IN:
//		fromName string // the name from which the letter will be sent
//		fromAddress string // the address from which the letter will be sent
//		toAddress string // the address to which the letter will be sent
//		subject string // subject of the letter
//		templateString string // template of the letter body
//		dataForTemplate struct // struct of data, which is used to fill email-body template
// OUT: nothing
func SendEmail (fromName string, fromAddress string, toAddress string, subject string, templateString string, dataForTemplate interface{}) {

	// Set up smtp-authentication information.
	auth := smtp.PlainAuth("",MAILLOGIN,MAILPASSWORD,MAILHOST)

	// Collect mail headers
	header := make(map[string]string)
	from := mail.Address{fromName, fromAddress}
	header["From"] = from.String()
	header["To"] = toAddress
	header["Subject"] = mime.QEncoding.Encode("utf-8", subject)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	// Generate mail body from template string
	t := template.New("eMailBody")
	t, err := t.Parse(templateString)
	if err != nil { panic(err) }
	var body bytes.Buffer
	err = t.Execute(&body, dataForTemplate)
	if err != nil { panic(err) }

	// Generate the whole mail message
	message := ""
	for k, v := range header { message += fmt.Sprintf("%s: %s\r\n", k, v) }
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body.String()))

	// Connect to the smtp-server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err = smtp.SendMail(
		MAILHOST+":"+MAILPORT,
		auth,
		fromAddress,
		[]string{toAddress},
		[]byte(message) )
	if err != nil { panic(err) }
}

// ===========================================================================================================================
// Database collections defenition
// ===========================================================================================================================

type typeUser struct {
	EMail string
	PasswordMD5 string
}

type typeSetPasswordLink struct {
	EMail string
	UUID string
	Expired time.Time
}


// ===========================================================================================================================
// Main program: start the web-server
// ===========================================================================================================================
func main() {

	// Read environment variables
	URI = os.Getenv("MONGODB_ADDON_URI")
	DB = os.Getenv("MONGODB_ADDON_DB")
	MAILHOST = os.Getenv("MAIL_HOST")
	MAILPORT = os.Getenv("MAIL_PORT")
	MAILLOGIN = os.Getenv("MAIL_LOGIN")
	MAILPASSWORD = os.Getenv("MAIL_PASSWORD")
	MAILFROM = os.Getenv("MAIL_FROM")
	SERVERHTTPADDRESS = os.Getenv("SERVER_HTTP_ADDRESS")
	LISTENPORT := os.Getenv("LISTEN_PORT")

	// Assign handlers for web requests
 	http.HandleFunc("/SignUp",webSignUp)
 	http.HandleFunc("/ChangePassword",webChangePasswordFormShow)
 	http.HandleFunc("/SetPassword",webSetPassword)
 	http.HandleFunc("/LogIn",webLogIn)
 	http.HandleFunc("/LogOut",webLogOut)
	http.HandleFunc("/",webFormShow)
	
	// Register a HTTP file server for delivery static files from the static directory
	fs := http.FileServer(http.Dir("./static"))
 	http.Handle("/static/", http.StripPrefix("/static/", fs))

 	// Launch the web server on all interfaces on the PORT port
	http.ListenAndServe(":"+LISTENPORT,nil)
}