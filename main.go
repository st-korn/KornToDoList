package main

import (
	"net/http" // for HTTP-server
	"os" // to get OS environment variables
	"text/template" // for use HTML-page templates
	"io/ioutil" // for read text files from the server
	"encoding/json" // to parse JSON (language strings-tables)
	"golang.org/x/text/language" // to detect user-perferred language
	"golang.org/x/text/language/display" // to output national names of languages
	"github.com/Shaked/gomobiledetect" // to detect mobile browsers
	"gopkg.in/mgo.v2" // to connect to MongoDB
	"gopkg.in/mgo.v2/bson" // to use BSON data format
)

var URI string //URI MongoDB database
var DB string //MongoDB database name

// ================================================================================
// Language defenition
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
	Labels map[string]string // strings-table of current language for HTML
}

func webFormShow(res http.ResponseWriter, req *http.Request) {

   	// Prepare main structure of HTML-template
	var webFormData typeWebFormData

	// Detect mobile devices
	detect := mobiledetect.NewMobileDetect(req, nil)
	webFormData.IsMobile = detect.IsMobile() && !detect.IsTablet()

	// Start detect user-language from cookie
	var langCookieEnglishName string
	var langTagCode string
	langCookie, err := req.Cookie("User-Language")
	if err == nil { langCookieEnglishName = langCookie.Value }

    // Load supported languages list
    webFormData.Langs = make([]typeLang,len(SupportedLangs))
	for i, tag := range SupportedLangs {
		webFormData.Langs[i].EnglishName = display.English.Tags().Name(tag)
		webFormData.Langs[i].NationalName = display.Self.Name(tag)
		if webFormData.Langs[i].EnglishName == langCookieEnglishName {langTagCode = tag.String()}
	}

	// Finish detect user-language
    accept := req.Header.Get("Accept-Language")
    matcher := language.NewMatcher(SupportedLangs)
    tag, _ := language.MatchStrings(matcher, langTagCode, accept)
    webFormData.UserLang = display.English.Tags().Name(tag);

	// Load strings-table for user-language
	jsonFile, err := os.Open("templates/"+webFormData.UserLang+".json")
	if err != nil { panic(err) }
	defer jsonFile.Close()
	jsonText, err := ioutil.ReadAll(jsonFile)
	if err != nil { panic(err) }
	json.Unmarshal(jsonText, &webFormData.Labels)

	// Apply HTML-template
	res.Header().Set("Content-type", "text/html")
	t := template.New("tasks.html")
	t, err = t.ParseFiles("templates/tasks.html")
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