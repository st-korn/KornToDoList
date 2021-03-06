package main

import (
	"net/http"      // for HTTP-server
	"text/template" // for use HTML-page templates

	"golang.org/x/text/language/display" // to output national names of languages
	"gopkg.in/mgo.v2/bson"               // to use BSON queries format
)

// ===========================================================================================================================
// WEB-PAGE: main page of web-application
// GET /
// GET /YYYY-MM-DD
// Cookies: User-Language  : string
// Cookies: User-Filter : string ["all", "created-only", "created-not-wait-not-remind"]
// ===========================================================================================================================

// Structure to fill HTML-template of main web-page
type typeWebFormData struct {
	UserLang    string     // english_name of current language, on which to display the page
	UserLangTag string     // tag of current language, on which to display the page
	Langs       []typeLang // global list of supported languages
	Labels      typeLabels // strings-table of current language for HTML
	ListToOpen  string     // selected list name (by URL path)
	UserFilter  string     // user-selected filter
}

func webFormShow(res http.ResponseWriter, req *http.Request) {

	// If necessary, we redirect requests received via the unprotected HTTP protocol to HTTPS and redirect to hostname without "www.""
	if RedirectIncomingHTTPandWWW(res, req) {
		return
	}

	// Prepare main structure of HTML-template
	var webFormData typeWebFormData

	// All calls to unknown url paths should return 404
	if req.URL.Path != "/" {
		// Check incoming parameters
		webFormData.ListToOpen = req.URL.Path[1:]
		passed, _ := TestTaskListName(webFormData.ListToOpen)
		if !passed {
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
	webFormData.UserLangTag, webFormData.UserLang, webFormData.Labels = DetectLanguageAndLoadLabels(req)

	// Detect user-selected filter from cookie
	filterCookie, err := req.Cookie("User-Filter")
	if err == nil {
		webFormData.UserFilter = filterCookie.Value
	} else {
		webFormData.UserFilter = "all"
	}

	// Apply HTML-template
	res.Header().Set("Content-type", "text/html")
	t := template.New("tasks.html")
	t, err = t.ParseFiles("templates/tasks.html")
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
	UUID        string     // UUID set-password-link
	UserLang    string     // english_name of current language, on which to display the page
	UserLangTag string     // tag of current language, on which to display the page
	Langs       []typeLang // global list of supported languages
	Labels      typeLabels // strings-table of current language for HTML
	Result      string     // A string that passes a precomputed and predefined result to a form, for example "Link is expired or account is not found"
}

func webChangePasswordFormShow(res http.ResponseWriter, req *http.Request) {

	// If necessary, we redirect requests received via the unprotected HTTP protocol to HTTPS and redirect to hostname without "www."
	if RedirectIncomingHTTPandWWW(res, req) {
		return
	}

	// Prepare main structure of HTML-template
	var changePasswordFormData typeChangePasswordFormData

	// Parse http GET-request params
	q := req.URL.Query()
	changePasswordFormData.UUID = q.Get("uuid")

	// Load supported languages list
	changePasswordFormData.Langs = make([]typeLang, len(SupportedLangs))
	for i, tag := range SupportedLangs {
		changePasswordFormData.Langs[i].EnglishName = display.English.Tags().Name(tag)
		changePasswordFormData.Langs[i].NationalName = display.Self.Name(tag)
	}

	// Detect user-language and load it labels
	changePasswordFormData.UserLangTag, changePasswordFormData.UserLang, changePasswordFormData.Labels = DetectLanguageAndLoadLabels(req)

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
