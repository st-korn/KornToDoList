package main

import (
	"net/http" // for HTTP-server
	"sort"     // for sortings lists

	// to connect to MongoDB
	"gopkg.in/mgo.v2/bson" // to use BSON queries format
)

// ===========================================================================================================================
// API: Get list of user lists.
// POST /GetLists
// Checks the current session for validity. If the session is not valid, it returns "SessionEmptyNotFoundOrExpired" as a result.
// Returns array of strings with names of saved users todo-lists.
// Cookies: User-Session : string (UUID)
// IN: -
// OUT: JSON: { Result : string ["OK", "SessionEmptyNotFoundOrExpired"], Lists : []string }
// ===========================================================================================================================

// Structure JSON-response for User Lists
type typeGetListsJSONResponse struct {
	Result string
	Lists  []string
}

func webGetLists(res http.ResponseWriter, req *http.Request) {

	// Preparing to response
	var response typeGetListsJSONResponse

	// Check if the current user session is valid
	email, session := TestSession(req)
	if email == "" {
		response.Result = "SessionEmptyNotFoundOrExpired"
		ReturnJSON(res, response)
	}

	// Select user lists
	c := session.DB(DB).C("Tasks")
	c.Find(bson.M{"email": email}).Distinct("list", &response.Lists)
	sort.Sort(sort.Reverse(sort.StringSlice(response.Lists)))
	response.Result = "OK"

	// Return JSON response
	ReturnJSON(res, response)
}
