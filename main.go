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
)

var URI string //URI MongoDB database
var DB string //MongoDB database name

// All supported languages
var SupportedLangs = []language.Tag{
    language.English,   // The first language is used as default
    language.German,
    language.Russian}

// Language structure: english_name_of_language and national_name_of_language
type typeLang struct {
	EnglishName string
	NationalName string
}

// Structure to fill HTML-template of main web-page
type typeWebFormData struct {
	IsMobile bool // flag: the site was opened from a mobile browser 
	UserLang string // english_name of current language, on which to display the page
	Langs []typeLang // global list of supported languages
	Labels map[string]string // strings-table of current language
}

// ================================================================================
// Web-page: main page of web-application
// ================================================================================
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

	// Detect user-language
	lang, _ := req.Cookie("User-Language")
    accept := req.Header.Get("Accept-Language")
    matcher := language.NewMatcher(SupportedLangs)
    tag, _ := language.MatchStrings(matcher, lang.String(), accept)
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
// Main program: start the web-server
// ================================================================================
func main() {

	// Read environment variables
	URI = os.Getenv("MONGODB_ADDON_URI")
	DB = os.Getenv("MONGODB_ADDON_DB")
	PORT := os.Getenv("PORT")

	// Assign handlers for web requests
	http.HandleFunc("/",webFormShow)
	
	// Register a HTTP file server for delivery static files from the static directory
	fs := http.FileServer(http.Dir("./static"))
 	http.Handle("/static/", http.StripPrefix("/static/", fs))

 	// Launch the web server on all interfaces on the PORT port
	http.ListenAndServe(":"+PORT,nil)
}