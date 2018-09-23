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

	"golang.org/x/text/language" // to detect user-perferred language
	"golang.org/x/text/language/display" // to output national names of languages
	"github.com/Shaked/gomobiledetect" // to detect mobile browsers
	"gopkg.in/mgo.v2" // to connect to MongoDB
	"gopkg.in/mgo.v2/bson" // to use BSON data format
)

var URI string //URI MongoDB database
var DB string //MongoDB database name
var PORT string //port on which the web server listens
var MAILHOST string //smtp-server hostname
var MAILPORT string //smtp-server port number (usually 25)
var MAILLOGIN string //login of smtp-server account
var MAILPASSWORD string //password of smtp-server account
var MAILFROM string //E-mail address, from which will be sent emails


// ================================================================================
// Language defenition and detection
// ================================================================================

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
// IN: req *http.Request
// OUT: lang *typeLang, labels *typeLabels
func DetectLanguageAndLoadLabels(req *http.Request, langtag *language.Tag, labels *typeLabels) {

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
    *langtag, _ = language.MatchStrings(matcher, langTagCode, accept)

    langEnglishName := display.English.Tags().Name(*langtag);

	// Load strings-table for user-language
	jsonFile, err := os.Open("templates/"+langEnglishName+".json")
	if err != nil { panic(err) }
	defer jsonFile.Close()
	jsonText, err := ioutil.ReadAll(jsonFile)
	if err != nil { panic(err) }
	json.Unmarshal(jsonText, labels)
}

// ================================================================================
// Database collections defenition
// ================================================================================

type typeUser struct {
	EMail string
	PasswordHash string
}

// ================================================================================
// Web-page: main page of web-application
// ================================================================================

// Structure to fill HTML-template of main web-page
type typeWebFormData struct {
	IsMobile bool // flag: the site was opened from a mobile browser 
	UserLang string // english_name of current language, on which to display the page
	Langs []typeLang // global list of supported languages
	Labels typeLabels // strings-table of current language for HTML
}

func webFormShow(res http.ResponseWriter, req *http.Request) {

   	// Prepare main structure of HTML-template
	var webFormData typeWebFormData
	var langTag language.Tag

	// Detect mobile devices
	detect := mobiledetect.NewMobileDetect(req, nil)
	webFormData.IsMobile = detect.IsMobile() && !detect.IsTablet()

    // Load supported languages list
    webFormData.Langs = make([]typeLang,len(SupportedLangs))
	for i, tag := range SupportedLangs {
		webFormData.Langs[i].EnglishName = display.English.Tags().Name(tag)
		webFormData.Langs[i].NationalName = display.Self.Name(tag)
	}

	// Detect user-language
	DetectLanguageAndLoadLabels(req, &langTag, &webFormData.Labels)
    webFormData.UserLang = display.English.Tags().Name(langTag);

	// Apply HTML-template
	res.Header().Set("Content-type", "text/html")
	t := template.New("tasks.html")
	t, err := t.ParseFiles("templates/tasks.html")
	if err != nil { panic(err) }
	err = t.Execute(res, webFormData)
	if err != nil { panic(err) }
}

// ================================================================================
// API: try to sign-up a new user. 
// In case of success, a link is sent to the user, after which he can set password and complete the registration. 
// Without opening the link, the account is not valid.
// IN: JSON: { email : "" }
// OUT: JSON: { result : string ["EMailEmpty", "UserJustExistsButEmailSent", "UserSignedUpEmailSent"] }
// ================================================================================

// Structure JSON-request for Sign Up new user
type typeSignUpJSONRequest struct {
	EMail string
}
// Structure JSON-response for Sign Up new user
type typeSignUpJSONResponse struct {
	Result string
}

func webSignUp(res http.ResponseWriter, req *http.Request) {

    // Parse request to struct
    var request typeSignUpJSONRequest
    err := json.NewDecoder(req.Body).Decode(&request)
    if err != nil { panic(err) }

    // Try to sign-up new user
    var response typeSignUpJSONResponse
    
    // Check request fields
    if request.EMail == "" { response.Result = "EMailEmpty" } else 
    {
    	// Serarch for this user exist

		// Connect to database
		session, err := mgo.Dial(URI)
    	if err != nil { panic(err) }
    	defer session.Close()
   		c := session.DB(DB).C("Users")

   		var user typeUser
   		err = c.Find(bson.M{"email": request.EMail}).One(&user)
   		if err != nil { 
   			response.Result = "UserSignedUpEmailSent"
   		} else
   		{
   			response.Result = "UserJustExistsButEmailSent"
   		}
    }

    // Generate change_password link

	// Detect user-language
	var labels typeLabels
	var langTag language.Tag
	DetectLanguageAndLoadLabels(req, &langTag, &labels)

   	// Set up smtp-authentication information.
	auth := smtp.PlainAuth("",MAILLOGIN,MAILPASSWORD,MAILHOST)
	// Collect mail headers and body
	from := mail.Address{labels["mailFrom"], MAILFROM}
	subject := labels["mailSignUpSubject"]
	body := labels["mailSignUpBody1"] + req.RemoteAddr + labels["mailSignUpBody2"];
	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = request.EMail
	header["Subject"] = mime.QEncoding.Encode("utf-8", subject)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"
	message := ""
	for k, v := range header { message += fmt.Sprintf("%s: %s\r\n", k, v) }
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// Connect to the smtp-server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err = smtp.SendMail(
		MAILHOST+":"+MAILPORT,
		auth,
		MAILFROM,
		[]string{request.EMail},
		[]byte(message) )
	if err != nil { panic(err) }

    // Create json response from struct
    resJSON, err := json.Marshal(response)
    if err != nil { panic(err) }
    res.Write(resJSON)
}

// ================================================================================
// Main program: start the web-server
// ================================================================================
func main() {

	// Read environment variables
	URI = os.Getenv("MONGODB_ADDON_URI")
	DB = os.Getenv("MONGODB_ADDON_DB")
	MAILHOST = os.Getenv("MAIL_HOST")
	MAILPORT = os.Getenv("MAIL_PORT")
	MAILLOGIN = os.Getenv("MAIL_LOGIN")
	MAILPASSWORD = os.Getenv("MAIL_PASSWORD")
	MAILFROM = os.Getenv("MAIL_FROM")
	PORT := os.Getenv("PORT")

	// Assign handlers for web requests
	http.HandleFunc("/SignUp",webSignUp)
	http.HandleFunc("/",webFormShow)
	
	// Register a HTTP file server for delivery static files from the static directory
	fs := http.FileServer(http.Dir("./static"))
 	http.Handle("/static/", http.StripPrefix("/static/", fs))

 	// Launch the web server on all interfaces on the PORT port
	http.ListenAndServe(":"+PORT,nil)
}