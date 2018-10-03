package main

import (
	"net/http" // for HTTP-server
	"encoding/json" // to parse JSON (language strings-tables)
	"net" // to split IP's and PORT's
	"net/url" // to split HTTP/HTTPS prefix
	"strings" // UpperCase and LowerCase functions
	"time" // access to system date and time - to control uuid's expirations
	"github.com/satori/go.uuid" // generate UUIDs
	"gopkg.in/mgo.v2/bson" // to use BSON queries format
)

// ===========================================================================================================================
// API: try to sign-up a new user. 
// POST /SignUp
// In case of success, a link is sent to the user, after which he can set password and complete the registration. 
// Without opening the link, the account is not valid.
// IN: JSON: { EMail : string }
// OUT: JSON: { Result : string ["EmptyEMail", "UserJustExistsButEmailSent", "UserSignedUpAndEmailSent"] }
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
	HttpAddress string // HTTP-address of http-server
	HostName string // HostName of http-server
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
    	response.Result = "EmptyEMail"
	    ReturnJSON(res, response)
    	return
    } 

	// Connect to database
	session := GetMongoDBSession()
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
	eMailBodyData.HttpAddress = SERVERHTTPADDRESS
	serverUrl, err := url.Parse(SERVERHTTPADDRESS)
	if err != nil { panic(err) }
	eMailBodyData.HostName, _, err = net.SplitHostPort(serverUrl.Host)
	eMailBodyData.HostName = strings.ToLower(eMailBodyData.HostName)
	
    // Generate change_password link
	c = session.DB(DB).C("SetPasswordLinks")
	u := uuid.Must(uuid.NewV4())
	c.Insert( bson.M {
		"email" : request.EMail,
		"uuid" : u.String(),
		"expired" : time.Now().UTC().AddDate(0,0,1) } )
	eMailBodyData.SetPasswordLink = SERVERHTTPADDRESS+"/ChangePassword?uuid="+u.String()

 	// Send email
 	SendEmail (labels["mailFrom"], MAILFROM, request.EMail, subject, bodyTemplate, eMailBodyData)

    // Return JSON response
    ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: add new user with password or change password for existing user. 
// POST /SetPassword
// Try to find current set-password-link. If not found - return "UUIDExpiredOrNotFound"
// If access is allowed, insert a document in MongoDB "Users" collection, or update it. 
// If success delete UUID record from MongoDB "SetPasswordLinks" collection (to block access to a link).
// Then return "UserAdded" or "PasswordUpdated".
// IN: JSON: { UUID : string, PasswordMD5 : string }
// OUT: JSON: { Result : string ["UUIDExpiredOrNotFound", "EmptyPassword", "PasswordUpdated"] }
// ===========================================================================================================================

// Structure JSON-request for Set Password
type typeSetPasswordJSONRequest struct {
	UUID string
	PasswordMD5 string
}
// Structure JSON-response for Set Password
type typeSetPasswordJSONResponse struct {
	Result string
}

func webSetPassword(res http.ResponseWriter, req *http.Request) {

    // Parse request to struct
    var request typeSetPasswordJSONRequest
    err := json.NewDecoder(req.Body).Decode(&request)
    if err != nil { panic(err) }

    // Preparing to response
    var response typeSignUpJSONResponse

	// Check request fields
    if request.PasswordMD5 == "" { 
    	response.Result = "EmptyPassword"
	    ReturnJSON(res, response)
    	return
    } 

	// Connect to database
	session := GetMongoDBSession()
	defer session.Close()
	c := session.DB(DB).C("SetPasswordLinks")

	// Try to find current set-password-link
	var setPasswordLink typeSetPasswordLink
	err = c.Find(bson.M{"uuid": request.UUID}).One(&setPasswordLink)
	if err != nil { 
		response.Result = "UUIDExpiredOrNotFound"
	    ReturnJSON(res, response)
		return
	}
	setPasswordLink.EMail = strings.ToLower(setPasswordLink.EMail)

	// Try to find user record in database
	c = session.DB(DB).C("Users")
	var user typeUser
	err = c.Find(bson.M{"email": setPasswordLink.EMail}).One(&user)
	if err != nil { 
		user.EMail = setPasswordLink.EMail
		user.PasswordMD5 = request.PasswordMD5
		c.Insert(user)
		response.Result = "UserCreated"
	} else
	{
		c.Update( bson.M{"email": setPasswordLink.EMail}, bson.M{"$set": bson.M{"passwordmd5": request.PasswordMD5}} )
		response.Result = "PasswordUpdated"
	}

	// Remove old set-password-links of this user
	c = session.DB(DB).C("SetPasswordLinks")
	c.RemoveAll(bson.M{"email" : setPasswordLink.EMail })

    // Return JSON response
    ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: Try to login an user, returns session-UUID.
// POST /LogIn
// If current session exist - then logout.
// If the pair (EMail and PasswordMD5) is not present in the collection `Users`, return "UserAndPasswordPairNotFound".
// Otherwise register new session in the `Sessions` database collection, and return its UUID to set cookie in browser.
// Cookies: User-Session : string (UUID)
// IN: JSON: { EMail : string, PasswordMD5 : string }
// OUT: JSON: { Result : string ["EmptyEMail", "EmptyPassword", "UserAndPasswordPairNotFound", "LoggedIn"], UUID : string }
// ===========================================================================================================================

// Structure JSON-request for Log In
type typeLogInJSONRequest struct {
	EMail string
	PasswordMD5 string
}
// Structure JSON-response for Log In
type typeLogInJSONResponse struct {
	Result string
	UUID string
}

func webLogIn(res http.ResponseWriter, req *http.Request) {

    // Parse request to struct
    var request typeLogInJSONRequest
    err := json.NewDecoder(req.Body).Decode(&request)
    if err != nil { panic(err) }

    // We store email-addresses only in lowecase
    request.EMail = strings.ToLower(request.EMail)

    // Preparing to response
    var response typeLogInJSONResponse

	// Check request fields
    if request.EMail == "" { 
    	response.Result = "EmptyEMail"
	    ReturnJSON(res, response)
    	return
    } 
    if request.PasswordMD5 == "" { 
    	response.Result = "EmptyPassword"
	    ReturnJSON(res, response)
    	return
    }

	// Connect to database
	session := GetMongoDBSession()
	defer session.Close()
	c := session.DB(DB).C("Sessions")

	// Try to find user
	c = session.DB(DB).C("Users")
	var user typeUser
	err = c.Find(bson.M{"email": request.EMail, "passwordmd5": request.PasswordMD5}).One(&user)
	if err != nil {
    	response.Result = "UserAndPasswordPairNotFound"
	    ReturnJSON(res, response)
    	return
    }

    // Try to detect previous user session
	sessionCookie, err := req.Cookie("User-Session")
	if err == nil { 
		// Remove previous session from the database
		c := session.DB(DB).C("Sessions")
		c.RemoveAll(bson.M{"uuid" : sessionCookie.Value })
	}

    // Generate new UUID
	u := uuid.Must(uuid.NewV4())
	response.Result = "LoggedIn"
	response.UUID = u.String()

    // Save new Session
	c = session.DB(DB).C("Sessions")
	c.Insert( bson.M{
		"uuid" : u.String(),
		"email" : request.EMail,
		"expired" : time.Now().UTC().AddDate(0,0,DefaultSessionLifetimeDays) } )

    // Return JSON response
    ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: Logout user and erase information in database about his active session.
// POST /LogOut
// Cookies: User-Session : string (UUID)
// IN: -
// OUT: JSON: { Result : string ["EmptySession", "LoggedOut"] }
// ===========================================================================================================================

// Structure JSON-response for Log Out
type typeLogOutJSONResponse struct {
	Result string
}

func webLogOut(res http.ResponseWriter, req *http.Request) {

    // Preparing to response
    var response typeLogInJSONResponse

    // Try to detect previous user session
	sessionCookie, err := req.Cookie("User-Session")
	if err != nil { 
		response.Result = "EmptySession"
	    ReturnJSON(res, response)
    	return
	}

	// Connect to database
	session := GetMongoDBSession()
	defer session.Close()
	c := session.DB(DB).C("Sessions")

	// Remove session from the database
	c.RemoveAll(bson.M{"uuid" : sessionCookie.Value })
	response.Result = "LoggedOut"

    // Return JSON response
    ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: Start an anonymous-session, returns session-UUID.
// POST /GoAnonymous
// If current session exist - then logout.
// Register new anonymous session in the `Sessions` database collection, and return its UUID to set cookie in browser.
// Cookies: User-Session : string (UUID)
// IN: -
// OUT: JSON: { Result : string ["SuccessAnonymous"], UUID : string }
// ===========================================================================================================================

// Structure JSON-response for Go Anonymous
type typeGoAnonymousJSONResponse struct {
	Result string
	UUID string
}

func webGoAnonymous(res http.ResponseWriter, req *http.Request) {

    // Preparing to response
    var response typeGoAnonymousJSONResponse

	// Connect to database
	session := GetMongoDBSession()
	defer session.Close()
	c := session.DB(DB).C("Sessions")

    // Try to detect previous user session
	sessionCookie, err := req.Cookie("User-Session")
	if err == nil { 
		// Remove previous session from the database
		c.RemoveAll(bson.M{"uuid" : sessionCookie.Value })
	}

    // Generate new UUID
	u := uuid.Must(uuid.NewV4())
	response.Result = "SuccessAnonymous"
	response.UUID = u.String()

	// Generate anonymous email-style name
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil { panic(err) }
	timeStr := time.Now().UTC().Format("2006-01-02-15-04-05")

    // Save new Session
	c = session.DB(DB).C("Sessions")
	c.Insert( bson.M{
		"uuid" : u.String(),
		"email" : ip+"@"+timeStr,
		"expired" : time.Now().UTC().AddDate(0,0,DefaultSessionLifetimeDays) } )

    // Return JSON response
    ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: Check session's UUID and get information of current user.
// POST /UserInfo
// Checks the current session for validity.
// If the session is valid, returns the Email (real or imaginary) of the current user. Returns flag: is the current user anonymous or not.
// If the session is not valid, it returns "SessionNotFoundOrExpired" as a result.
// Cookies: User-Session : string (UUID)
// IN: -
// OUT: JSON: { Result : string ["EmptySession", "ValidUserSession", "ValidAnonymousSession", "SessionNotFoundOrExpired"], EMail : string }
// ===========================================================================================================================

// Structure JSON-response for User Info
type typeUserInfoJSONResponse struct {
	Result string
	EMail string
}

func webUserInfo(res http.ResponseWriter, req *http.Request) {

    // Preparing to response
    var response typeUserInfoJSONResponse

	// Connect to database
	session := GetMongoDBSession()
	defer session.Close()
	c := session.DB(DB).C("Sessions")

    // Try to detect previous user session
	sessionCookie, err := req.Cookie("User-Session")
	if err != nil { 
		response.Result = "EmptySession"
	    ReturnJSON(res, response)
    	return
	}

	// Try to find active session
	var sid typeSession
	err = c.Find(bson.M{"uuid": sessionCookie.Value}).One(&sid)
	if err != nil {
		response.Result = "SessionNotFoundOrExpired"
	    ReturnJSON(res, response)
    	return
	}

	response.EMail = sid.EMail

	// Try to find registered user
	var user typeUser
	c = session.DB(DB).C("Users")
	err = c.Find(bson.M{"email": sid.EMail}).One(&user)
	if err != nil {
		response.Result = "ValidAnonymousSession"
	    ReturnJSON(res, response)
    	return
	}

	response.Result = "ValidUserSession"

    // Return JSON response
    ReturnJSON(res, response)
}