package main

import (
	"encoding/json"
	"net/http" // for HTTP-server

	"gopkg.in/mgo.v2/bson" // to use BSON queries format
)

// ===========================================================================================================================
// API: Get tasks of selected user lists.
// POST /GetTasks
// Checks the current session for validity. If the session is not valid, it returns "SessionEmptyNotFoundOrExpired" as a result.
// Returns an array of structures that identify tasks from a selected list of the current user.
// Cookies: User-Session : string (UUID)
// IN: JSON: {List : string}
// OUT: JSON: { Result : string ["OK", "SessionEmptyNotFoundOrExpired"], Tasks : [] { Id : string, Text : string,
// section : string ["iu","in","nu","nn","ib"], status : string ["created", "done", "canceled", "moved"],
// icon : string ["wait","remind","call","force","mail","prepare","manage","meet","visit","make","journey","think"] } }
// ===========================================================================================================================

// Structure JSON-request for getting tasks
type typeGetTasksJSONRequest struct {
	List string
}

// Structure JSON-response for getting tasks
type typeGetTasksJSONResponse struct {
	Result string
	Tasks  []typeTask
}

func webGetTasks(res http.ResponseWriter, req *http.Request) {

	// Parse request to struct
	var request typeGetTasksJSONRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	// Preparing to response
	var response typeGetTasksJSONResponse

	// Check if the current user session is valid
	email, session := TestSession(req)
	if email == "" {
		response.Result = "SessionEmptyNotFoundOrExpired"
		ReturnJSON(res, response)
	}

	// Select tasks of user list
	c := session.DB(DB).C("Tasks")
	c.Find(bson.M{"email": email, "list": request.List}).All(&response.Tasks)
	response.Result = "OK"

	// Return JSON response
	ReturnJSON(res, response)
}
