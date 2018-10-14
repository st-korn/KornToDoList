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
// Returns LastModifiedTimestamp = max(Tasks.Timestamp)
// Also return list of today's tasks from selected list of the current user.
// Cookies: User-Session : string (UUID)
// IN: JSON: {List : string}
// OUT: JSON: { Result : string ["OK", "InvalidListName", "SessionEmptyNotFoundOrExpired"],
//				Tasks : [] { Id : string,
//							 EMail : string,
//							 List : string,
//							 Text : string,
//							 Section : string ["iu","in","nu","nn","ib"],
//							 Status : string ["created", "done", "canceled", "moved"],
//							 Icon : string ["wait","remind","call","force","mail","prepare","manage","meet","visit","make","journey","think"],
//							 Timestamp : datetime },
//				LastModifiedTimestamp : datetime,
//				TodayTasks []string (_id task or "" for delimiter),
//				TodayTasksTimestamp : datetime }
// ===========================================================================================================================

// Structure JSON-request for getting tasks
type typeGetTasksJSONRequest struct {
	List string
}

// Structure JSON-response for getting tasks
type typeGetTasksJSONResponse struct {
	Result                string
	Tasks                 []typeTask
	LastModifiedTimestamp time.Time
	TodayTasks            []string
	TodayTasksTimestamp   time.Time
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

	// Select tasks of user list
	c := session.DB(DB).C("Tasks")
	c.Find(bson.M{"email": email, "list": request.List, "text": bson.M{"$ne": ""}}).All(&response.Tasks)

	// Select max timestamp
	var lastTask typeTask
	c.Find(bson.M{"email": email, "list": request.List, "text": bson.M{"$ne": ""}}).Sort("-timestamp").One(&lastTask)
	response.LastModifiedTimestamp = lastTask.Timestamp

	// Select today's tasks list
	var todaysTasks typeTodaysTaks
	c = session.DB(DB).C("TodaysTasks")
	c.Find(bson.M{"email": email, "list": request.List}).One(&todaysTasks)
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
// return `"TaskJustUpdated"` error.
// Update existing task or generate ID and append new task to the database.
// Returns an ID and Timestamp of created or modified task.
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
//				Id : string,
//				Timestamp : datetime }
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
	Result    string
	Id        string
	Timestamp time.Time
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
		response.Id = request.Id

		// Looking for an updated task
		err = c.Find(bson.M{"email": email, "_id": bson.ObjectIdHex(response.Id)}).One(&task)
		if err != nil {
			response.Result = "UpdatedTaskNotFound"
			ReturnJSON(res, response)
			return
		}

		// Compare timestamps
		durationSeconds := task.Timestamp.Sub(request.Timestamp).Seconds()
		if durationSeconds > 0 {
			response.Result = "TaskJustUpdated"
			response.Timestamp = task.Timestamp
			ReturnJSON(res, response)
			return
		}

		// Update existing task
		response.Timestamp = time.Now().UTC()
		err = c.Update(bson.M{"email": email, "_id": task.Id},
			bson.M{"$set": bson.M{"list": request.List, "text": request.Text, "section": request.Section,
				"icon": request.Icon, "status": request.Status, "timestamp": response.Timestamp}})
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

		response.Id = task.Id.Hex()
		response.Timestamp = task.Timestamp
		response.Result = "TaskInserted"
	}

	// Return JSON response
	ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: Move existing task from one list to another.
// POST /MoveTask
// Checks the current session for validity. If the session is not valid, it returns `"SessionEmptyNotFoundOrExpired"` as a result.
// If updated task exist in database, and its timestamp is greater than timestamp of updated task, recieved from users-application, return `"TaskJustUpdated"` error.
// Generate ID and append new task to destination task-list in the database.
// Update existing task in source task-list: set status = "moved".
// If Today flag is true, then add created task to today's taks list of destination task-list.
// Returns Timestamp of modified task from source task-list.
// Cookies: User-Session : string (UUID)
// IN: JSON: { Id : string (don't may be null or ""),
//			   ToList : string,
//			   Text : string,
//			   Section : string ["iu","in","nu","nn","ib"],
//			   Status : string ["created", "done", "canceled", "moved"],
//			   Icon : string ["wait","remind","call","force","mail","prepare","manage","meet","visit","make","journey","think"],
//			   Timestamp : datetime (updated task timestamp, can't be null or ""),
//			   Today : bool }
// OUT: JSON: { Result : string ["TaskEmpty", "InvalidListName", "SessionEmptyNotFoundOrExpired", "UpdatedTaskNotFound",
//								 "InsertFailed", "UpdateFailed", "TodaysTaskListUpdateFailed", "TaskJustUpdated", "TaskMoved"],
//				Timestamp : datetime }
// ===========================================================================================================================

// Structure JSON-request for getting tasks
type typeMoveTaskJSONRequest struct {
	Id        string
	ToList    string
	Text      string
	Section   string
	Icon      string
	Status    string
	Timestamp time.Time
	Today     bool
}

// Structure JSON-response for getting tasks
type typeMoveTaskJSONResponse struct {
	Result    string
	Id        string
	Timestamp time.Time
}

func webMoveTask(res http.ResponseWriter, req *http.Request) {

	// Parse request to struct
	var request typeMoveTaskJSONRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	// Preparing to response
	var response typeMoveTaskJSONResponse

	// Check incoming parameters
	request.Text = strings.TrimSpace(request.Text)
	if request.Section == "" {
		request.Section = "ib"
	}
	if request.Status == "" {
		request.Section = "created"
	}
	if (request.Text == "") || (request.Id == "") {
		response.Result = "TaskEmpty"
		ReturnJSON(res, response)
		return
	}
	passed, _ := TestTaskListName(request.ToList)
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

	// Looking for an updated task
	var task typeTask
	c := session.DB(DB).C("Tasks")
	err = c.Find(bson.M{"email": email, "_id": bson.ObjectIdHex(request.Id)}).One(&task)
	if err != nil {
		response.Result = "UpdatedTaskNotFound"
		ReturnJSON(res, response)
		return
	}

	// Compare timestamps
	durationSeconds := task.Timestamp.Sub(request.Timestamp).Seconds()
	if durationSeconds > 0 {
		response.Result = "TaskJustUpdated"
		response.Timestamp = task.Timestamp
		ReturnJSON(res, response)
		return
	}

	// Create new task
	// Generate new task _id
	task.Id = bson.NewObjectId()
	task.EMail = email
	task.List = request.ToList
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

	// Update existing task
	response.Timestamp = time.Now().UTC()
	err = c.Update(bson.M{"email": email, "_id": bson.ObjectIdHex(request.Id)},
		bson.M{"$set": bson.M{"status": "moved", "timestamp": response.Timestamp}})
	if err != nil {
		response.Result = "UpdateFailed"
		ReturnJSON(res, response)
		return
	}

	// Add the task to today's task list if needed
	if request.Today {
		// Looking for exists today's tasks list
		var todaysTasks typeTodaysTaks
		c := session.DB(DB).C("TodaysTasks")
		err = c.Find(bson.M{"email": email, "list": request.ToList}).One(&todaysTasks)
		if err == nil {
			// Update existing today's tasks list
			todaysTasks.Tasks = append(todaysTasks.Tasks, task.Id.Hex())
			todaysTasks.Timestamp = time.Now().UTC()
			err = c.Update(bson.M{"email": email, "list": request.ToList},
				bson.M{"$set": bson.M{"tasks": todaysTasks.Tasks, "timestamp": todaysTasks.Timestamp}})
			if err != nil {
				response.Result = "TodaysTaskListUpdateFailed"
				ReturnJSON(res, response)
				return
			}
			response.Result = "TodaysTaskListUpdated"
		} else {
			// Create new today's tasks list
			todaysTasks.EMail = email
			todaysTasks.List = request.ToList
			todaysTasks.Tasks = append(todaysTasks.Tasks, task.Id.Hex())
			todaysTasks.Timestamp = time.Now().UTC()
			// Insert new today's tasks list
			err = c.Insert(&todaysTasks)
			if err != nil {
				response.Result = "TodaysTaskListUpdateFailed"
				ReturnJSON(res, response)
				return
			}
		}
	}

	// Return JSON response
	response.Result = "TaskMoved"
	ReturnJSON(res, response)
}

// ===========================================================================================================================
// API: Checks need to update the task list.
// POST /NeedUpdate
// Checks the current session for validity. If the session is not valid, it returns `"SessionEmptyNotFoundOrExpired"` as a result.
// Compare LastModifiedTimestamp with max timestamp of list's tasks. If any of tasks timestamps is greater - return "NeedUpdate".
// Compare TodayTasksTimestamp with timestamp of today's task list. If today's taks list's timestamp is greater - return "NeedUpdate".
// Otherwise return "AllActual".
// Cookies: User-Session : string (UUID)
// IN: JSON: { List string,
//			   LastModifiedTimestamp : datetime,
//			   TodayTasksTimestamp : datetime }
// OUT: JSON: { Result : string ["InvalidListName", "SessionEmptyNotFoundOrExpired", "AllActual", "NeedUpdate"] }
// ===========================================================================================================================

// Structure JSON-request for getting tasks
type typeNeedUpdateJSONRequest struct {
	List                  string
	LastModifiedTimestamp time.Time
	TodayTasksTimestamp   time.Time
}

// Structure JSON-response for getting tasks
type typeNeedUpdateJSONResponse struct {
	Result string
}

func webNeedUpdate(res http.ResponseWriter, req *http.Request) {

	// Parse request to struct
	var request typeNeedUpdateJSONRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	// Preparing to response
	var response typeNeedUpdateJSONResponse

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

	// Select max timestamp
	var lastTask typeTask
	c := session.DB(DB).C("Tasks")
	c.Find(bson.M{"email": email, "list": request.List, "text": bson.M{"$ne": ""}}).Sort("-timestamp").One(&lastTask)

	// Compare timestamps
	durationSeconds := lastTask.Timestamp.Sub(request.LastModifiedTimestamp).Seconds()
	if durationSeconds > 0 {
		response.Result = "NeedUpdate"
		ReturnJSON(res, response)
		return
	}

	// Select today's tasks list
	var todaysTasks typeTodaysTaks
	c = session.DB(DB).C("TodaysTasks")
	c.Find(bson.M{"email": email, "list": request.List}).One(&todaysTasks)

	// Compare timestamps
	durationSeconds = todaysTasks.Timestamp.Sub(request.TodayTasksTimestamp).Seconds()
	if durationSeconds > 0 {
		response.Result = "NeedUpdate"
		ReturnJSON(res, response)
		return
	}

	// Return JSON response
	response.Result = "AllActual"
	ReturnJSON(res, response)
}
