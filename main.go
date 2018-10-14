package main

import (
	"bytes"           // string buffer to use string-templates
	"encoding/base64" // to use UTF-8 in emails bodys
	"encoding/json"   // to parse JSON (language strings-tables)
	"fmt"             // to generate emails body
	"io/ioutil"       // for read text files from the server
	"mime"            // to use UTF-8 in emails headers
	"net/http"        // for HTTP-server
	"net/mail"        // to generate emails
	"net/smtp"        // to send emails
	"os"              // to get OS environment variables
	"regexp"          // to validation todo-list names
	"strings"         // to parse http-request headers
	"text/template"   // for use HTML-page templates
	"time"            // to define MongoDB time type for structure elements, for timers, to validation todo-list names

	"golang.org/x/text/language"         // to detect user-perferred language
	"golang.org/x/text/language/display" // to output national names of languages
	"gopkg.in/mgo.v2"                    // to connect to MongoDB
	"gopkg.in/mgo.v2/bson"               // to use BSON queries format
)

// Global envirnment variables
var URI string               //URI MongoDB database
var DB string                //MongoDB database name
var MAILHOST string          //smtp-server hostname
var MAILPORT string          //smtp-server port number (usually 25)
var MAILLOGIN string         //login of smtp-server account
var MAILPASSWORD string      //password of smtp-server account
var MAILFROM string          //E-mail address, from which will be sent emails
var SERVERHTTPADDRESS string //HTTP-address for web-server with http or https prefix, and without any path: (eg. http://127.0.0.1:9000 or https://todo.works)
var LISTENPORT string        //port on which the web server listens

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
		if err != nil {
			panic(err)
		}
		SESSION.SetMode(mgo.Monotonic, true)
	}
	return SESSION.Clone()
}

// Connect to the database, test cookie and return EMail of active session's user.
// Return "" if session UUID empty, expired or not found.
// IN:	req : *http.Request - http request, from which the cookie is read from the UUID of the user session.
// OUT:	email : string - email of current authenticated user, or "" if session UUID empty, session is expired or not found.
//		session : *mgo.Session - session variable, cloned from global MongoDB session.
//								 Returns for possible later use for new database queries.
func TestSession(req *http.Request) (email string, session *mgo.Session) {

	// Connect to database
	session = GetMongoDBSession()
	c := session.DB(DB).C("Sessions")

	// Try to detect current user session
	sessionCookie, err := req.Cookie("User-Session")
	if err != nil {
		return "", session
	}

	// Try to find active session
	var sid typeSession
	err = c.Find(bson.M{"uuid": sessionCookie.Value}).One(&sid)
	if err != nil {
		return "", session
	}

	email = sid.EMail
	return email, session
}

// Regularly clears the base of expired sessions
func ClearExpiredSessiond() {

	// Connect to database
	session := GetMongoDBSession()
	defer session.Close()

	// Remove all expired set-password-links
	c := session.DB(DB).C("SetPasswordLinks")
	c.RemoveAll(bson.M{"expired": bson.M{"$lte": time.Now().UTC()}})

	// Remove all expired sessions
	c = session.DB(DB).C("Sessions")
	c.RemoveAll(bson.M{"expired": bson.M{"$lte": time.Now().UTC()}})

	// Write log to STDOUT
	fmt.Println(time.Now().Format("2006-01-02 15:04:05 MST") + " Expired sesions cleaned.")
}

// ===========================================================================================================================
// Language defenition and detection
// ===========================================================================================================================

// All supported languages
var SupportedLangs = []language.Tag{
	language.English, // The first language is used as default
	language.Russian}

// Language structure: english_name_of_language and national_name_of_language
type typeLang struct {
	EnglishName  string
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
	if err == nil {
		langCookieEnglishName = langCookie.Value
	}

	// Finduser-language from supported languages list
	for _, tag := range SupportedLangs {
		if display.English.Tags().Name(tag) == langCookieEnglishName {
			langTagCode = tag.String()
		}
	}

	// Finish detect user-language
	accept := req.Header.Get("Accept-Language")
	matcher := language.NewMatcher(SupportedLangs)
	langTag, _ = language.MatchStrings(matcher, langTagCode, accept)

	langEnglishName = display.English.Tags().Name(langTag)

	// Load strings-table for user-language
	jsonFile, err := os.Open("templates/" + langEnglishName + ".json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	jsonText, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(jsonText, &labels)

	return langTag, langEnglishName, labels
}

// ===========================================================================================================================
// Task lists common routines
// ===========================================================================================================================

// Check if the string "list" matches the format "YYYY-MM-DD"
// If it matches, returns true as result, and it's value in time.Time ad datetime
// If don't matches, returns false as result
func TestTaskListName(list string) (result bool, datetime time.Time) {
	regexp, _ := regexp.Compile(`\d\d\d\d-\d\d-\d\d`)
	if !regexp.MatchString(list) {
		return false, datetime
	}
	datetime, err := time.Parse("2006-01-02", list)
	if err != nil {
		return false, datetime
	}
	return true, datetime
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
// OUT: -
func SendEmail(fromName string, fromAddress string, toAddress string, subject string, templateString string, dataForTemplate interface{}) {

	// Set up smtp-authentication information.
	auth := smtp.PlainAuth("", MAILLOGIN, MAILPASSWORD, MAILHOST)

	// Collect mail headers
	header := make(map[string]string)
	from := mail.Address{Name: fromName, Address: fromAddress}
	header["From"] = from.String()
	header["To"] = toAddress
	header["Subject"] = mime.QEncoding.Encode("utf-8", subject)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	// Generate mail body from template string
	t := template.New("eMailBody")
	t, err := t.Parse(templateString)
	if err != nil {
		panic(err)
	}
	var body bytes.Buffer
	err = t.Execute(&body, dataForTemplate)
	if err != nil {
		panic(err)
	}

	// Generate the whole mail message
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body.String()))

	// Connect to the smtp-server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err = smtp.SendMail(
		MAILHOST+":"+MAILPORT,
		auth,
		fromAddress,
		[]string{toAddress},
		[]byte(message))
	if err != nil {
		panic(err)
	}
}

// ===========================================================================================================================
// Common Web-server routines
// ===========================================================================================================================

// Debug function: output to STDOUT all HTTP-headers, received in the HTTP-request
// IN: *http.Request
// OUT: -
func FormatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}
	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}

// If found incoming HTTP-request through cloud HTTPS-proxy - then redirect user's browser from http to https
// IN: http.ResponseWriter, *http.Request
// OUT: bool [true - redirect required, false - you can continue to serve the request]
func RedirectIncomingHTTPandWWW(res http.ResponseWriter, req *http.Request) bool {
	hostOriginal := strings.ToLower(req.Host)
	hostWithoutWWW := hostOriginal
	if hostOriginal[0:4] == "www." {
		hostWithoutWWW = hostOriginal[4:]
	}
	if (strings.ToLower(req.Header.Get("X-Forwarded-Proto")) == "http") || (hostOriginal != hostWithoutWWW) {
		target := "https://" + hostWithoutWWW + req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			target += "?" + req.URL.RawQuery
		}
		http.Redirect(res, req, target, http.StatusMovedPermanently)
		return true
	} else {
		return false
	}
}

// Return JSON in response to a http-request
// IN:	res - http.Responsewriter, in which the returned json is written
// 		structJSON - structure to convert to JSON
// OUT: -
func ReturnJSON(res http.ResponseWriter, structJSON interface{}) {
	// Create json response from struct
	resJSON, err := json.Marshal(structJSON)
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-type", "application/json; charset=utf-8")
	res.Write(resJSON)

}

// ===========================================================================================================================
// Database collections defenition
// ===========================================================================================================================

type typeUser struct {
	EMail       string
	PasswordMD5 string
}

type typeSetPasswordLink struct {
	EMail     string
	UUID      string
	Expired   time.Time
	Anonymous string
}

type typeSession struct {
	UUID    string
	EMail   string
	Expired time.Time
}

type typeTask struct {
	Id        bson.ObjectId `json:"Id" bson:"_id"`
	EMail     string
	List      string
	Text      string
	Section   string
	Status    string
	Icon      string
	Timestamp time.Time
}

type typeTodaysTaks struct {
	EMail     string
	List      string
	Tasks     []string
	Timestamp time.Time
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
	http.HandleFunc("/SignUp", webSignUp)
	http.HandleFunc("/ChangePassword", webChangePasswordFormShow)
	http.HandleFunc("/SetPassword", webSetPassword)
	http.HandleFunc("/LogIn", webLogIn)
	http.HandleFunc("/LogOut", webLogOut)
	http.HandleFunc("/GoAnonymous", webGoAnonymous)
	http.HandleFunc("/UserInfo", webUserInfo)
	http.HandleFunc("/GetLists", webGetLists)
	http.HandleFunc("/GetTasks", webGetTasks)
	http.HandleFunc("/SendTask", webSendTask)
	http.HandleFunc("/CreateList", webCreateList)
	http.HandleFunc("/SaveTodayTasks", webSaveTodayTasks)
	http.HandleFunc("/NeedUpdate", webNeedUpdate)
	http.HandleFunc("/", webFormShow)

	// Register a HTTP file server for delivery static files from the static directory
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start timer to clean expired session every hour
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	go func() {
		for _ = range ticker.C {
			ClearExpiredSessiond()
		}
	}()

	// Launch the web server on all interfaces on the PORT port
	http.ListenAndServe(":"+LISTENPORT, nil)
}
