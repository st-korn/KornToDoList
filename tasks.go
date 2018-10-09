package main

import (
	"encoding/json" // to parse and generate JSON input and output parameters
	"net/http"      // for HTTP-server
	"strings"       // TrimSpace function
	"time"          // for timestamps

	"gopkg.in/mgo.v2/bson" // to use BSON queries format
)

// ===========================================================================================================================
// API: Get tasks of selected user lists.
// POST /GetTasks
// Checks the current session for validity. If the session is not valid, it returns "SessionEmptyNotFoundOrExpired" as a result.
// Returns an array of structures that identify tasks from a selected list of the current user.
// Also return list of today's tasks from selected list of the current user.
// Cookies: User-Session : string (UUID)
// IN: JSON: {List : string}
// OUT: JSON: { Result : string ["OK", "SessionEmptyNotFoundOrExpired"],
//				Tasks : [] { Id : string,
//							 EMail : string,
//							 List : string,
//							 Text : string,
//							 Section : string ["iu","in","nu","nn","ib"],
//							 Status : string ["created", "done", "canceled", "moved"],
//							 Icon : string ["wait","remind","call","force","mail","prepare","manage","meet","visit","make","journey","think"],
//							 Timestamp : datetime },
//				TodayTasks []string (_id task or "" for delimiter),
//				TodayTasksTimestamp : datetime }
// ===========================================================================================================================

// Structure JSON-request for getting tasks
type typeGetTasksJSONRequest struct {
	List string
}

// Structure JSON-response for getting tasks
type typeGetTasksJSONResponse struct {
	Result              string
	Tasks               []typeTask
	TodayTasks          []string
	TodayTasksTimestamp time.Time
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

	// Select today's tasks list
	var todaysTasks typeTodaysTaks
	c = session.DB(DB).C("TodaysTasks")
	err = c.Find(bson.M{"email": email, "list": request.List}).One(&todaysTasks)
	response.TodayTasks = todaysTasks.Tasks
	response.TodayTasksTimestamp = todaysTasks.Timestamp

	response.Result = "OK"

	// Return JSON response
	ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: Update existing task from the list or append new task to the list.
// POST /SendTask
// If updated task exist in database, and its timestamp is greater than timestamp of updated task, recieved from users-application,
// return `"TaskJustUpdated"` error and array of a single element - original task from the database.
// Update existing task or generate ID and append new task to the database.
// Returns an array of a single element - an added or updated task with its ID.
// Cookies: User-Session : string (UUID)
// IN: JSON: { List : string,
//			   Id : string (may be null or ""),
//			   Text : string,
// 			   Section : string ["iu","in","nu","nn","ib"],
//			   Status : string ["created", "done", "canceled", "moved"],
// 			   Icon : string ["wait","remind","call","force","mail","prepare","manage","meet","visit","make","journey","think"],
//			   Timestamp : datetime (can't be null or "") }
// OUT: JSON: { Result : string ["TaskEmpty", "InvalidListName", "SessionEmptyNotFoundOrExpired", "UpdatedTaskNotFound",
//								 "UpdateFailed", "TaskJustUpdated", "TaskUpdated", "InsertFailed", "TaskInserted"],
//				Tasks : [] { Id : string,
//							 EMail : string,
//							 List : string,
//							 Text : string,
//							 Section : string,
//							 Status : string,
//							 Icon : string,
//							 Timestamp : datetime } }
// ===========================================================================================================================

// Structure JSON-request for getting tasks
type typeSendTaskJSONRequest struct {
	List      string
	Id        string
	Text      string
	Section   string
	Icon      string
	Status    string
	Timestamp time.Time
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
