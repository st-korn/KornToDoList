package main

import (
	"encoding/json"
	"net/http" // for HTTP-server
	"time"

	"gopkg.in/mgo.v2/bson"
	// to validation todo-list name, for timestamps
	// to use BSON queries format
)

// ===========================================================================================================================
// API: Save today's task list in database.
// POST /SaveTodayTasks
// Checks the current session for validity. If the session is not valid, it returns `"SessionEmptyNotFoundOrExpired"` as a result.
// If today's task-list exist in database, and its timestamp is greater than timestamp of updated task-list, recieved from users-application,
// return `"TodaysTaskListJustUpdated"` error and original today's task list from the database.
// Update today's task-list of current list in the database.
// Returns an array of updated today's task-list.
// Cookies: User-Session : string (UUID)
// IN: JSON: { List string,
//			   TodayTasks []string (_id task or "" for delimiter),
//			   TodayTasksTimestamp : datetime (updated task timestamp, can't be null or "") }
// OUT: JSON: { Result : string ["InvalidListName", "SessionEmptyNotFoundOrExpired", "TodaysTaskListUpdateFailed", "TodaysTaskListJustUpdated", "TodaysTaskListUpdated",
//				TodayTasks []string (_id task or "" for delimiter),
//				TodayTasksTimestamp : datetime }
// ===========================================================================================================================

// Structure JSON-request for getting tasks
type typeSaveTodayTasksJSONRequest struct {
	List                string
	TodayTasks          []string
	TodayTasksTimestamp time.Time
}

// Structure JSON-response for getting tasks
type typeSaveTodayTasksJSONResponse struct {
	Result              string
	TodayTasks          []string
	TodayTasksTimestamp time.Time
}

func webSaveTodayTasks(res http.ResponseWriter, req *http.Request) {

	// Parse request to struct
	var request typeSaveTodayTasksJSONRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	// Preparing to response
	var response typeSaveTodayTasksJSONResponse

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

	// Looking for exists today's tasks list
	var todaysTasks typeTodaysTaks
	c := session.DB(DB).C("TodaysTasks")
	err = c.Find(bson.M{"email": email, "list": request.List}).One(&todaysTasks)
	if err == nil {

		// Compare timestamps
		durationSeconds := todaysTasks.Timestamp.Sub(request.TodayTasksTimestamp).Seconds()
		if durationSeconds > 0 {
			response.Result = "TodaysTaskListJustUpdated"
			response.TodayTasks = todaysTasks.Tasks
			response.TodayTasksTimestamp = todaysTasks.Timestamp
			ReturnJSON(res, response)
			return
		}

		// Update existing today's tasks list
		err = c.Update(bson.M{"email": email, "list": request.List},
			bson.M{"$set": bson.M{"tasks": request.TodayTasks, "timestamp": time.Now().UTC()}})
		if err != nil {
			response.Result = "TodaysTaskListUpdateFailed"
			ReturnJSON(res, response)
			return
		}

		// Obtain an updated today's tasks list
		err = c.Find(bson.M{"email": email, "list": request.List}).One(&todaysTasks)
		if err != nil {
			response.Result = "TodaysTaskListUpdateFailed"
			ReturnJSON(res, response)
			return
		}

		response.Result = "TodaysTaskListUpdated"
	} else {

		// Create new today's tasks list
		todaysTasks.EMail = email
		todaysTasks.List = request.List
		todaysTasks.Tasks = request.TodayTasks
		todaysTasks.Timestamp = time.Now().UTC()

		// Insert new today's tasks list
		err = c.Insert(&todaysTasks)
		if err != nil {
			response.Result = "TodaysTaskListUpdateFailed"
			ReturnJSON(res, response)
			return
		}

		// Obtain an inserted today's tasks list
		err = c.Find(bson.M{"email": email, "list": request.List}).One(&todaysTasks)
		if err != nil {
			response.Result = "TodaysTaskListUpdateFailed"
			ReturnJSON(res, response)
			return
		}

		response.Result = "TodaysTaskListUpdated"
	}

	response.TodayTasks = todaysTasks.Tasks
	response.TodayTasksTimestamp = todaysTasks.Timestamp

	// Return JSON response
	ReturnJSON(res, response)
}
