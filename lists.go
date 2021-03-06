package main

import (
	"encoding/json" // to parse and generate JSON input and output parameters
	"net/http"      // for HTTP-server
	"sort"          // for sortings lists
	"time"          // to validation todo-list name

	"gopkg.in/mgo.v2/bson" // to use BSON queries format
)

// ===========================================================================================================================
// API: Get list of user lists.
// POST /GetLists
// Checks the current session for validity. If the session is not valid, it returns "SessionEmptyNotFoundOrExpired" as a result.
// Returns array of strings with names of saved users todo-lists.
// Cookies: User-Session : string (UUID)
// IN: -
// OUT: JSON: { Result : string ["OK", "SessionEmptyNotFoundOrExpired"],
//				Lists : []string }
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
		return
	}

	// Select user lists
	c := session.DB(DB).C("Tasks")
	c.Find(bson.M{"email": email}).Distinct("list", &response.Lists)
	sort.Sort(sort.Reverse(sort.StringSlice(response.Lists)))
	response.Result = "OK"

	// Return JSON response
	ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: Create new todo-list for current user.
// POST /CreateList
// Checks the current session for validity. If the session is not valid, it returns "SessionEmptyNotFoundOrExpired" as a result.
// Verifies that the desired list name does not differ by more than 24 hours from the current server date.
// Add new task with empty Text field, assigned to created list name.
// Such tasks with empty Text field are sevice tasks is not returned by the GetTasks API.
// Returns array of strings with names of saved users todo-lists.
// Cookies: User-Session : string (UUID)
// IN: JSON: {List : string "YYYY-MM-DD"}
// OUT: JSON: { Result : string ["ListCreated", "InvalidListName", "DateTooFar", "CreateListFailed", "SessionEmptyNotFoundOrExpired"],
//				Lists : []string }
// ===========================================================================================================================

// Structure JSON-request for getting tasks
type typeCreateListJSONRequest struct {
	List string
}

// Structure JSON-response for getting tasks
type typeCreateListJSONResponse struct {
	Result string
	Lists  []string
}

func webCreateList(res http.ResponseWriter, req *http.Request) {
	// Parse request to struct
	var request typeCreateListJSONRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	// Preparing to response
	var response typeCreateListJSONResponse

	// Check incoming parameters
	passed, dateList := TestTaskListName(request.List)
	if !passed {
		response.Result = "InvalidListName"
		ReturnJSON(res, response)
		return
	}
	durationHours := time.Now().UTC().Sub(dateList).Hours()
	if (durationHours > 24) || (durationHours < -24) {
		response.Result = "DateTooFar"
		ReturnJSON(res, response)
		return
	}

	// Check if the current user session is valid
	email, session := TestSession(req)
	if email == "" {
		response.Result = "SessionEmptyNotFoundOrExpired"
		ReturnJSON(res, response)
		return
	}

	// Try to insert new empty task
	var task typeTask
	task.Id = bson.NewObjectId()
	task.EMail = email
	task.List = request.List
	task.Text = ""
	task.Section = ""
	task.Icon = ""
	task.Status = ""

	// Insert new task
	c := session.DB(DB).C("Tasks")
	err = c.Insert(&task)
	if err != nil {
		response.Result = "CreateListFailed"
		ReturnJSON(res, response)
		return
	}

	// Select user lists
	c.Find(bson.M{"email": email}).Distinct("list", &response.Lists)
	sort.Sort(sort.Reverse(sort.StringSlice(response.Lists)))

	// Return JSON response
	response.Result = "ListCreated"
	ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: Remove empty todo-list.
// POST /RemoveEmptyList
// Checks the current session for validity. If the session is not valid, it returns "SessionEmptyNotFoundOrExpired" as a result.
// If there are no other tasks with the filled Text field in the list, it will automatically cease to exist,
// since there will be no tasks that belong to it.
// Returns array of strings with names of saved users todo-lists.
// Cookies: User-Session : string (UUID)
// IN: JSON: {List : string "YYYY-MM-DD"}
// OUT: JSON: { Result : string ["ListRemoved", "InvalidListName", "SessionEmptyNotFoundOrExpired"],
//				Lists : []string }
// ===========================================================================================================================

// Structure JSON-request for getting tasks
type typeRemoveEmptyListJSONRequest struct {
	List string
}

// Structure JSON-response for getting tasks
type typeRemoveEmptyListJSONResponse struct {
	Result string
	Lists  []string
}

func webRemoveEmptyList(res http.ResponseWriter, req *http.Request) {
	// Parse request to struct
	var request typeRemoveEmptyListJSONRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	// Preparing to response
	var response typeRemoveEmptyListJSONResponse

	// Check incoming parameters
	passed, _ := TestTaskListName(request.List)
	if !passed {
		response.Result = "InvalidListName"
		ReturnJSON(res, response)
		return
	}

	// Check if the current user session is valid
	email, session := TestSession(req)
	if email == "" {
		response.Result = "SessionEmptyNotFoundOrExpired"
		ReturnJSON(res, response)
		return
	}

	// Purge all empty task of this user
	c := session.DB(DB).C("Tasks")
	c.RemoveAll(bson.M{"email": email, "list": request.List, "text": ""})

	// Select user lists
	c.Find(bson.M{"email": email}).Distinct("list", &response.Lists)
	sort.Sort(sort.Reverse(sort.StringSlice(response.Lists)))

	// Return JSON response
	response.Result = "ListRemoved"
	ReturnJSON(res, response)
}
