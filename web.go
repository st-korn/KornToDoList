package main

import (
	"net/http" // for HTTP-server
	"text/template" // for use HTML-page templates
	"time" // access to system date and time - to control uuid's expirations
	"golang.org/x/text/language/display" // to output national names of languages
	"github.com/Shaked/gomobiledetect" // to detect mobile browsers
	"gopkg.in/mgo.v2/bson" // to use BSON data format
)


// ===========================================================================================================================
// WEB-PAGE: main page of web-application
// GET /
// Cookies: User-Language : string
// ===========================================================================================================================

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
// WEB-PAGE: web-page to set user-s password
// GET /SetPassword?uuid=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
// Cookies: User-Language : string
// IN: uuid : string
// ===========================================================================================================================

// Structure to fill HTML-template of main web-page
type typeChangePasswordFormData struct {
	UUID string // UUID set-password-link
	IsMobile bool // flag: the site was opened from a mobile browser
	Labels typeLabels // strings-table of current language for HTML
	Result string // A string that passes a precomputed and predefined result to a form, for example "Link is expired or account is not found"
}

func webChangePasswordFormShow(res http.ResponseWriter, req *http.Request) {

   	// Prepare main structure of HTML-template
	var changePasswordFormData typeChangePasswordFormData

	// Parse http GET-request params
	q := req.URL.Query()
	changePasswordFormData.UUID = q.Get("uuid")

	// Detect mobile devices
	detect := mobiledetect.NewMobileDetect(req, nil)
	changePasswordFormData.IsMobile = detect.IsMobile() && !detect.IsTablet()

	// Detect user-language and load it labels
	_, _, changePasswordFormData.Labels = DetectLanguageAndLoadLabels(req)

	// Connect to database
	session := GetMongoDBSession()
	defer session.Close()
	c := session.DB(DB).C("SetPasswordLinks")

	// Remove all expired set-password-links
	c.RemoveAll(bson.M{"expired" : bson.M{"$lte":time.Now().UTC()} })

	// Try to find current set-password-link
	var setPasswordLink typeSetPasswordLink
	err := c.Find(bson.M{"uuid": changePasswordFormData.UUID}).One(&setPasswordLink)
	if err != nil { 
		changePasswordFormData.Result = changePasswordFormData.Labels["resultUUIDExpiredOrNotFound"]
	}

	// Apply HTML-template
	t := template.New("changepassword.html")
	t, err = t.ParseFiles("templates/changepassword.html")
	if err != nil { panic(err) }
	res.Header().Set("Content-type", "text/html")
	err = t.Execute(res, changePasswordFormData)
	if err != nil { panic(err) }
}