// ===========================================================================
// Generate <li> HTML for the Task structure
// IN: taskId string
// ===========================================================================
function htmlLiTask(taskId) {
	var taskIcon = $("#"+taskId).children("img").attr("alt");
	var taskStatus = $("#"+taskId).attr("class");
	var taskTimestamp = $("#"+taskId).attr("data-timestamp");
	var taskText = $("#"+taskId).text();
	var tooltips = {'created': statusCreated, 'moved': statusMoved, 'canceled': statusCanceled, 'done': statusDone};
	var li = '<li>'
	li = li + '<img class="handle" src="/static/icons/updown.svg" alt="move task">'
	if (taskIcon) {
		li = li + '<img class="icon '+taskStatus+'" src="/static/icons/'+taskIcon+'.svg" alt="'+taskIcon+'">'
	}
	li = li + '<div class="today-task '+taskStatus+'" id="div-' + taskId + '" class="' + taskStatus + '" data-tooltip="' + tooltips[taskStatus] + '" data-timestamp="'+taskTimestamp+'">'
	var idx = taskText.indexOf(" - ");
	if (idx > 0) {
		li = li + '<span class="employee">' + taskText.substr(0,idx) + '</span>' + taskText.substring(idx,taskText.length);
	} else {
		li = li + taskText;
	}
	li = li + '</div>'
	li = li + '<img class="insert-delimiter" src="/static/icons/delimiter.svg" title="'+hintInsertDelimiter+'" alt="insert delimiter">'
	li = li + '</li>';
	return li;
}

// ===========================================================================
// Generate <li> HTML for delimiter
// ===========================================================================
function htmlLiDelimiter() {
	var li = '<li class="delimiter">';
	li = li + '<img class="handle" src="/static/icons/updown.svg" alt="move delimiter">';
	li = li + '<div class="delimiter"></div>';
	li = li + '<img class="delete-delimiter" src="/static/icons/delete.svg" title="'+hintDeleteDelimiter+'" alt="remove delimiter">';
	li = li + '</li>';
	return li;	
}

// ===========================================================================
// Update today's todo-list from the current task in task-form
// ===========================================================================
function updateTodaysTasksList(id) {
	if ($("#checkbox-today-input").prop('checked')) {
		if ( $("div#div-"+id).length == 0 ) {
			// Try to include task to the today's task-list
			$("#today-tasks-ul").append(htmlLiTask(id));
			// Save today's tasks
			saveToday();	
		} else {
			// Try to update task in the today's task-list
			$("div#div-"+id).parents("li").replaceWith(htmlLiTask(id));
		}
	} else {
		// Try to exclude task from the today's task-list
		if ($("div#div-"+id).length > 0) {
			$("div#div-"+id).parents("li").remove();
			// Save today's tasks
			saveToday();
		}
	};
}
// ===========================================================================
// Send today's task list on server, to remember it in database
// ===========================================================================
function saveToday() {
	// Collect array of today's tasks
	var todayTasks = []
	$("#today-tasks-ul li").each( function() {
		if ($(this).attr("class") == "delimiter") {
			todayTasks[todayTasks.length] = ""
		} else {
			todayTasks[todayTasks.length] = $(this).children("div.today-task").attr("id").substr(4)
		}
	});	
	// Send Ajax POST request
	$("#operation-status-label").text("");
	showSpinner("#task-spinner-div");
	$.ajax( {
		url : "/SaveTodayTasks",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( {
			List: $("#task-lists-select").val(),
			TodayTasks: todayTasks,
			TodayTasksTimestamp: $("#today-tasks-ul").attr("data-timestamp")
		} ),
		// if success
		success: function (response) {
			hideSpinner("#task-spinner-div");
			switch(response.Result) {
  				case "SessionEmptyNotFoundOrExpired" :
  					Cookies.remove('User-Session');
  					location.reload();
					break;
				case "InvalidListName" :
					$("#operation-status-label").html(resultInvalidListName);
					break;
				case "DateTooFar" :
					$("#operation-status-label").html(resultDateTooFar);
					break;
				case "TodaysTaskListUpdateFailed" :
					$("#operation-status-label").html(resultTodaysTaskListUpdateFailed);
					break;
				case "TodaysTaskListJustUpdated" :
					$("#operation-status-label").html(resultTodaysTaskListJustUpdated);
					loadTasks();
					break;
				case "TodaysTaskListUpdated" :
					$("#operation-status-label").html(resultTodaysTaskListUpdated);
					// Update Timestamp
					$("#today-tasks-ul").attr("data-timestamp", response.TodayTasksTimestamp);
					break;
				default : $("#operation-status-label").html(resultUnknown);
			}
		},
		// if error returns
		error: function(jqXHR,exception) { 
			hideSpinner("#task-spinner-div");
			showAjaxError("#operation-status-label",jqXHR,exception);
		}
	} );
	return false;
}

// ===========================================================================
// Insert delimiter line after selected task in today's task-list
// ===========================================================================
function insertDelimiter() {
	$(this).parents("li").after(htmlLiDelimiter());
	// Save today's tasks
	saveToday();	
	// Refresh events handlers
	setEvents();
}

// ===========================================================================
// Delete delimiter line
// ===========================================================================
function deleteDelimiter() {
	$(this).parents("li").remove();
	// Save today's tasks
	saveToday();
	// Refresh events handlers
	setEvents();
}

// ===========================================================================
// Remove all tasks not in "created" status from today's tasks list
// ===========================================================================
function clearDoneTasksFromToday() {
	$("div.today-task.done").parents("li").remove();
	$("div.today-task.canceled").parents("li").remove();
	$("div.today-task.moved").parents("li").remove();
	saveToday();
}

// ===========================================================================
// Remove all tasks from today's tasks list
// ===========================================================================
function clearAllTasksFromToday() {
	$("li").remove();
	saveToday();
}