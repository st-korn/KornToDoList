package main

import (
	"net/http"      // for HTTP-server
	"regexp"        // to validation todo-list name
	"text/template" // for use HTML-page templates
	"time"          // to validation todo-list name

	"golang.org/x/text/language/display" // to output national names of languages
	"gopkg.in/mgo.v2/bson"               // to use BSON queries format
)

// ===========================================================================================================================
// WEB-PAGE: main page of web-application
// GET /
// GET /YYYY-MM-DD
// Cookies: User-Language  : string
// ===========================================================================================================================

// Structure to fill HTML-template of main web-page
type typeWebFormData struct {
	UserLang   string     // english_name of current language, on which to display the page
	Langs      []typeLang // global list of supported languages
	Labels     typeLabels // strings-table of current language for HTML
	ListToOpen string     // selected list name (by URL path)
}

func webFormShow(res http.ResponseWriter, req *http.Request) {

	// If necessary, we redirect requests received via the unprotected HTTP protocol to HTTPS
	if RedirectIncomingHTTPtoHTTPS(res, req) {
		return
	}

	// Prepare main structure of HTML-template
	var webFormData typeWebFormData

	// All calls to unknown url paths should return 404
	if req.URL.Path != "/" {
		// Check incoming parameters
		webFormData.ListToOpen = req.URL.Path[1:]
		regexp, _ := regexp.Compile(`\d\d\d\d-\d\d-\d\d`)
		if !regexp.MatchString(webFormData.ListToOpen) {
			http.NotFound(res, req)
			return
		}
		_, err := time.Parse("2006-01-02", webFormData.ListToOpen)
		if err != nil {
			http.NotFound(res, req)
			return
		}
	}

	// Load supported languages list
	webFormData.Langs = make([]typeLang, len(SupportedLangs))
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
	if err != nil {
		panic(err)
	}
	err = t.Execute(res, webFormData)
	if err != nil {
		panic(err)
	}
}

// ===========================================================================================================================
// WEB-PAGE: web-page to set user-s password
// GET /SetPassword?uuid=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
// Cookies: User-Language : string
// IN: uuid : string
// ===========================================================================================================================

// Structure to fill HTML-template of main web-page
type typeChangePasswordFormData struct {
	UUID   string     // UUID set-password-link
	Labels typeLabels // strings-table of current language for HTML
	Result string     // A string that passes a precomputed and predefined result to a form, for example "Link is expired or account is not found"
}

func webChangePasswordFormShow(res http.ResponseWriter, req *http.Request) {

	// If necessary, we redirect requests received via the unprotected HTTP protocol to HTTPS
	if RedirectIncomingHTTPtoHTTPS(res, req) {
		return
	}

	// Prepare main structure of HTML-template
	var changePasswordFormData typeChangePasswordFormData

	// Parse http GET-request params
	q := req.URL.Query()
	changePasswordFormData.UUID = q.Get("uuid")

	// Detect user-language and load it labels
	_, _, changePasswordFormData.Labels = DetectLanguageAndLoadLabels(req)

	// Connect to database
	session := GetMongoDBSession()
	defer session.Close()
	c := session.DB(DB).C("SetPasswordLinks")

	// Try to find current set-password-link
	var setPasswordLink typeSetPasswordLink
	err := c.Find(bson.M{"uuid": changePasswordFormData.UUID}).One(&setPasswordLink)
	if err != nil {
		changePasswordFormData.Result = changePasswordFormData.Labels["resultUUIDExpiredOrNotFound"]
	}

	// Apply HTML-template
	t := template.New("changepassword.html")
	t, err = t.ParseFiles("templates/changepassword.html")
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-type", "text/html")
	err = t.Execute(res, changePasswordFormData)
	if err != nil {
		panic(err)
	}
}
