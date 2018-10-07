package main

import (
	"encoding/json" // to parse and generate JSON input and output parameters
	"net/http"      // for HTTP-server
	"regexp"        // to validation todo-list name
	"strings"       // TrimSpace function
	"time"          // to validation todo-list name, for timestamps

	"gopkg.in/mgo.v2/bson" // to use BSON queries format
)

// ===========================================================================================================================
// API: Get tasks of selected user lists.
// POST /GetTasks
// Checks the current session for validity. If the session is not valid, it returns "SessionEmptyNotFoundOrExpired" as a result.
// Returns an array of structures that identify tasks from a selected list of the current user.
// Cookies: User-Session : string (UUID)
// IN: JSON: {List : string}
// OUT: JSON: { Result : string ["OK", "SessionEmptyNotFoundOrExpired"], Tasks : [] { Id : string, EMail : string, List : string,
//		Text : string, Section : string ["iu","in","nu","nn","ib"], status : string ["created", "done", "canceled", "moved"],
//		Icon : string ["wait","remind","call","force","mail","prepare","manage","meet","visit","make","journey","think"] } }
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
		return
	}

	// Select tasks of user list
	c := session.DB(DB).C("Tasks")
	c.Find(bson.M{"email": email, "list": request.List, "text": bson.M{"$ne": ""}}).All(&response.Tasks)
	response.Result = "OK"

	// Return JSON response
	ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: Update existing task from the list or append new task to the list.
// POST /SendTask
// Checks the current session for validity. If the session is not valid, it returns "SessionEmptyNotFoundOrExpired" as a result.
// Update existing task or generate ID and ppend new task to the database.
// Returns an array of a single element - an added or updated task with its ID.
// Cookies: User-Session : string (UUID)
// IN: JSON: { List : string, Id : string (may be null or ""), Text : string,
// 			Section : string ["iu","in","nu","nn","ib"], Status : string ["created", "done", "canceled", "moved"],
// 			Icon : string ["wait","remind","call","force","mail","prepare","manage","meet","visit","make","journey","think"]}
// OUT: JSON: { Result : string ["TaskEmpty", "InvalidListName", "SessionEmptyNotFoundOrExpired", "UpdatedTaskNotFound",
//			"UpdateFailed", "TaskJustUpdated", "TaskUpdated", "InsertFailed", "TaskInserted"],
//			Tasks : [] { Id : string, EMail : string, List : string, Text : string, Section : string, Status : string, Icon : string } }
// ===========================================================================================================================

// Structure JSON-request for getting tasks
type typeSendTaskJSONRequest struct {
	List      string
	Id        string
	Text      string
	Section   string
	Icon      string
	Status    string
	Timestamp time.Time `json:",omitempty"`
}

// Structure JSON-response for getting tasks
type typeSendTaskJSONResponse struct {
	Result string
	Tasks  []typeTask
}

func webSendTask(res http.ResponseWriter, req *http.Request) {

	// Parse request to struct
	var request typeSendTaskJSONRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	// Preparing to response
	var response typeSendTaskJSONResponse

	// Check incoming parameters
	request.Text = strings.TrimSpace(request.Text)
	if request.Section == "" {
		request.Section = "ib"
	}
	if request.Status == "" {
		request.Section = "created"
	}
	if request.Text == "" {
		response.Result = "TaskEmpty"
		ReturnJSON(res, response)
		return
	}
	regexp, _ := regexp.Compile(`\d\d\d\d-\d\d-\d\d`)
	if !regexp.MatchString(request.List) {
		response.Result = "InvalidListName"
		ReturnJSON(res, response)
		return
	}
	_, err = time.Parse("2006-01-02", request.List)
	if err != nil {
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

	// Try to update or insert task
	var task typeTask
	c := session.DB(DB).C("Tasks")
	if len(request.Id) > 0 {

		// Convert _id from string to MongoDB BSON format
		task.Id = bson.ObjectIdHex(request.Id)

		// Looking for an updated task
		err = c.Find(bson.M{"email": email, "_id": task.Id}).One(&task)
		if err != nil {
			response.Result = "UpdatedTaskNotFound"
			ReturnJSON(res, response)
			return
		}

		// Compare timestamps
		durationSeconds := task.Timestamp.Sub(request.Timestamp).Seconds()
		if durationSeconds > 0 {
			response.Result = "TaskJustUpdated"
			response.Tasks = append(response.Tasks, task)
			ReturnJSON(res, response)
			return
		}

		// Update existing task
		err = c.Update(bson.M{"email": email, "_id": task.Id},
			bson.M{"$set": bson.M{"list": request.List, "text": request.Text, "section": request.Section,
				"icon": request.Icon, "status": request.Status, "timestamp": time.Now().UTC()}})
		if err != nil {
			response.Result = "UpdateFailed"
			ReturnJSON(res, response)
			return
		}

		// Obtain an updated task
		err = c.Find(bson.M{"email": email, "_id": task.Id}).One(&task)
		if err != nil {
			response.Result = "UpdateFailed"
			ReturnJSON(res, response)
			return
		}

		response.Result = "TaskUpdated"

	} else {

		// Create new task
		// Generate new task _id
		task.Id = bson.NewObjectId()
		task.EMail = email
		task.List = request.List
		task.Text = request.Text
		task.Section = request.Section
		task.Icon = request.Icon
		task.Status = request.Status
		task.Timestamp = time.Now().UTC()

		// Insert new task
		err = c.Insert(&task)
		if err != nil {
			response.Result = "InsertFailed"
			ReturnJSON(res, response)
			return
		}

		// Obtain an inserted task
		err = c.Find(bson.M{"email": email, "list": request.List, "_id": task.Id}).One(&task)
		if err != nil {
			response.Result = "InsertFailed"
			ReturnJSON(res, response)
			return
		}

		response.Result = "TaskInserted"
	}

	response.Tasks = append(response.Tasks, task)

	// Return JSON response
	ReturnJSON(res, response)
}
