package main

import (
	"net/http" // for HTTP-server
	// to validation todo-list name, for timestamps
	// to use BSON queries format
)

// ===========================================================================================================================
// API: Save today's task list in database.
// POST /SaveTodayTasks
// Checks the current session for validity. If the session is not valid, it returns `"SessionEmptyNotFoundOrExpired"` as a result.
// If today's task-list exist in database, and its timestamp is greater than timestamp of updated task-list, recieved from users-application, return `"TodaysTaskListJustUpdated"` error.
// Update today's task-list of current list in the database.
// Returns an array of updated today's task-list.
// Cookies: User-Session : string (UUID)
// IN: JSON: { List string,
//			   TodayTasks []string (_id task or "" for delimiter),
//			   Timestamp : datetime (updated task timestamp, can't be null or "") }
// OUT: JSON: { Result : string ["InvalidListName", "SessionEmptyNotFoundOrExpired", "TodaysTaskListJustUpdated", "TodaysTaskListUpdated",
//				TodayTasks []string (_id task or "" for delimiter) }

// ===========================================================================================================================

func webSaveTodayTasks(res http.ResponseWriter, req *http.Request) {

}
