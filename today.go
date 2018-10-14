package main

import (
	"encoding/json" // to parse and generate JSON input and output parameters
	"net/http"      // for HTTP-server
	"time"          // for timestamps

	"gopkg.in/mgo.v2/bson" // to use BSON queries format
)

// ===========================================================================================================================
// API: Save today's task list in database.
// POST /SaveTodayTasks
// Checks the current session for validity. If the session is not valid, it returns `"SessionEmptyNotFoundOrExpired"` as a result.
// If today's task-list exist in database, and its timestamp is greater than timestamp of updated task-list, recieved from users-application,
// return `"TodaysTaskListJustUpdated"` error.
// Update today's task-list of current list in the database.
// Returns timestamp of updated today's task-list.
// Cookies: User-Session : string (UUID)
// IN: JSON: { List string,
//			   TodayTasks []string (_id task or "" for delimiter),
//			   TodayTasksTimestamp : datetime (updated task timestamp, can't be null or "") }
// OUT: JSON: { Result : string ["InvalidListName", "SessionEmptyNotFoundOrExpired", "TodaysTaskListUpdateFailed", "TodaysTaskListJustUpdated", "TodaysTaskListUpdated"],
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
			response.TodayTasksTimestamp = todaysTasks.Timestamp
			ReturnJSON(res, response)
			return
		}

		// Update existing today's tasks list
		response.TodayTasksTimestamp = time.Now().UTC()
		err = c.Update(bson.M{"email": email, "list": request.List},
			bson.M{"$set": bson.M{"tasks": request.TodayTasks, "timestamp": response.TodayTasksTimestamp}})
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

		response.TodayTasksTimestamp = todaysTasks.Timestamp
		response.Result = "TodaysTaskListUpdated"
	}

	// Return JSON response
	ReturnJSON(res, response)
}
