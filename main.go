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
	"net" // to split IP's and PORT's
	"strings" // UpperCase and LowerCase functions
	"fmt" // to generate emails body
	"bytes" // string buffer to use string-templates
	"time" // access to system date and time - to control uuid's expirations
	"golang.org/x/text/language" // to detect user-perferred language
	"golang.org/x/text/language/display" // to output national names of languages
	"github.com/Shaked/gomobiledetect" // to detect mobile browsers
	"github.com/satori/go.uuid" // generate UUIDs
	"gopkg.in/mgo.v2" // to connect to MongoDB
	"gopkg.in/mgo.v2/bson" // to use BSON data format
)

// Global variables
var URI string //URI MongoDB database
var DB string //MongoDB database name
var MAILHOST string //smtp-server hostname
var MAILPORT string //smtp-server port number (usually 25)
var MAILLOGIN string //login of smtp-server account
var MAILPASSWORD string //password of smtp-server account
var MAILFROM string //E-mail address, from which will be sent emails
var SERVERHOSTNAME string //Hostname for web-server
var PORT string //port on which the web server listens


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
	PasswordHash string
}

// ================================================================================
// WEB-PAGE: main page of web-application
// GET /
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

	// Detect mobile devices
	detect := mobiledetect.NewMobileDetect(req, nil)
	webFormData.IsMobile = detect.IsMobile() && !detect.IsTablet()

    // Load supported languages list
    webFormData.Langs = make([]typeLang,len(SupportedLangs))
	for i, tag := range SupportedLangs {
		webFormData.Langs[i].EnglishName = display.English.Tags().Name(tag)
		webFormData.Langs[i].NationalName = display.Self.Name(tag)
	}

	// Detect user-language and load it labels
	_, webFormData.UserLang, webFormData.Labels = DetectLanguageAndLoadLabels(req)

	// Apply HTML-template
	res.Header().Set("Content-type", "text/html")
	t := template.New("tasks.html")
	t, err := t.ParseFiles("templates/tasks.html")
	if err != nil { panic(err) }
	err = t.Execute(res, webFormData)
	if err != nil { panic(err) }
}

// ===========================================================================================================================
// API: try to sign-up a new user. 
// POST /SignUp
// In case of success, a link is sent to the user, after which he can set password and complete the registration. 
// Without opening the link, the account is not valid.
// IN: JSON: { email : string }
// OUT: JSON: { result : string ["EMailEmpty", "UserJustExistsButEmailSent", "UserSignedUpAndEmailSent"] }
// ===========================================================================================================================

// Structure JSON-request for Sign Up new user
type typeSignUpJSONRequest struct {
	EMail string
}
// Structure JSON-response for Sign Up new user
type typeSignUpJSONResponse struct {
	Result string
}

// Structure to fill template string of mail message body
type typeEMailBodyData struct {
	RequestIP string // IP-address of site-user
	HostName string // Hostname of http-server
	SetPasswordLink string // Generated link to change password
}

func webSignUp(res http.ResponseWriter, req *http.Request) {

    // Parse request to struct
    var request typeSignUpJSONRequest
    err := json.NewDecoder(req.Body).Decode(&request)
    if err != nil { panic(err) }

    // We store email-addresses only in lowecase
    request.EMail = strings.ToLower(request.EMail)

    // Preparing to response
    var response typeSignUpJSONResponse
    
    // Check request fields
    if request.EMail == "" { 
    	response.Result = "EMailEmpty"
    	return
    } 

	// Connect to database
	session, err := mgo.Dial(URI)
	if err != nil { panic(err) }
	defer session.Close()
	c := session.DB(DB).C("Users")

	// Detect user-language and load it labels
	_, _, labels := DetectLanguageAndLoadLabels(req)

   	// Serarch for this user exist
	var user typeUser
	var subject, bodyTemplate string
	err = c.Find(bson.M{"email": request.EMail}).One(&user)
	if err != nil { 
		response.Result = "UserSignedUpAndEmailSent"
		subject = labels["mailSignUpSubject"]
		bodyTemplate = labels["mailSignUpBody"]
	} else
	{
		response.Result = "UserJustExistsButEmailSent"
		subject = labels["mailRestorePasswordSubject"]
		bodyTemplate = labels["mailRestorePasswordBody"]
	}

	// Prepare struct for email body-template
	var eMailBodyData typeEMailBodyData
	eMailBodyData.RequestIP, _, err = net.SplitHostPort(req.RemoteAddr)
	if err != nil { panic(err) }
	eMailBodyData.HostName = strings.ToLower(SERVERHOSTNAME)

    // Generate change_password link
	c = session.DB(DB).C("SetPasswordLinks")
	u := uuid.Must(uuid.NewV4())
	c.Insert( bson.M {
		"email" : request.EMail,
		"uuid" : u.String(),
		"expired" : time.Now().UTC().AddDate(0,0,1) } )
	eMailBodyData.SetPasswordLink = "https://"+SERVERHOSTNAME+"/SetPassword?uuid="+u.String()

 	// Send email
 	SendEmail (labels["mailFrom"], MAILFROM, request.EMail, subject, bodyTemplate, eMailBodyData)

    // Create json response from struct
    resJSON, err := json.Marshal(response)
    if err != nil { panic(err) }
    res.Header().Set("Content-type", "application/json; charset=utf-8")
    res.Write(resJSON)
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
	SERVERHOSTNAME = os.Getenv("SERVER_HOST_NAME")
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