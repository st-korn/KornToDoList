// Timer initialization
var timerIdle;

// ===========================================================================
// Generate <p> HTML for the Task structure
// IN: task struct {	ID : string, Text : string, Section : string, Status : string, Icon : string, Timestamp : string }
// ===========================================================================
function htmlPTask(task) {
	var tooltips = {'created': statusCreated, 'moved': statusMoved, 'canceled': statusCanceled, 'done': statusDone};
	var p = '<p id="' + task.Id + '" class="' + task.Status + '" data-tooltip="' + tooltips[task.Status] + '" data-timestamp="'+task.Timestamp+'">'
	if (task.Icon) {
		p = p + '<img class="icon '+task.Status+'" src="/static/icons/'+task.Icon+'.svg" alt="'+task.Icon+'">'
	}
	var idx = task.Text.indexOf(" - ");
	if (idx > 0) {
		p = p + '<span class="employee">' + task.Text.substr(0,idx) + '</span>' + task.Text.substring(idx,task.Text.length);
	} else {
		p = p + task.Text;
	}
	p = p + '</p>';
	return p;
}

// ===========================================================================
// Load from server all tasks of current selected list
// ===========================================================================
function loadTasks() {
	// Purge current tasks
	$("p").remove();
	$("li").remove();
	// If no ToDo-list found - then nothing to do
	if ( !$("#task-lists-select").val() ) { return false };
	// Check selected list: if it isn't first - change current URL
	if ($("#task-lists-select").prop("selectedIndex") == 0) {
		window.history.replaceState("", "", '/');
	} else {
		window.history.replaceState("", "", '/'+$("#task-lists-select").val());
	}
	// Send Ajax POST request
	//$("#operation-status-label").text(""); // don't clear status label: for exception's of updating task or tody's-list
	showSpinner("#task-spinner-div");
	$.ajax( {
		url : "/GetTasks",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( { List: $("#task-lists-select").val()} ),
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
				case "OK" :
					// Collect tasks
					$.each(response.Tasks, function() {
						$("#"+this.Section).append(htmlPTask(this));
					});
					// Collect today's tasks
					$.each(response.TodayTasks, function() {
						if ( this.length > 0 ) {
							// insert task
							$("#today-tasks-ul").append(htmlLiTask(this));
						} else {
							// insert delimiter
							$("#today-tasks-ul").append(htmlLiDelimiter());
						}; 
					} );
					// Update Timestamps
					$("#today-tasks-ul").attr("data-timestamp", response.TodayTasksTimestamp);
					$("#tasks-main").attr("data-timestamp", response.LastModifiedTimestamp);
					// Clear status label
					$("#operation-status-label").text("");
					// Re-apply filter
					applyFilter($("#filter-input").val());
					// Calculate page statistic
					calculateStatistic();
					// Refresh events handlers
					setEvents();
					// Manage visibility and availability of list buttons
					manageListsButtons();
					// Get ready to create a new task
					$("#ib button").click();
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
// Fill field of Task form values of clicked task. Prepare to task editing.
// ===========================================================================
function onTaskEdit() {
	var id
	switch ($(this).prop("tagName")) {
		case "P" :
			id = $(this).attr("id")
			break;
		case "DIV" :
			id = $(this).attr("id").substr(4);
			break;
	}
	$("#task-status-select").val( $("p#"+id).attr("class") );
	$("#task-icon-select").val( $("p#"+id).children("img").attr("alt") );
	$("#task-section-select").val( $("p#"+id).parents("section").attr("id") );
	$("#task-text-input").val( $("p#"+id).text() );
	$("#task-id-input").val( id );
	$("#task-timestamp-input").val( $("p#"+id).attr("data-timestamp") )
	if ($("div#div-"+id).length > 0) {
		$("#checkbox-today-input").prop('checked', true);
	} else {
		$("#checkbox-today-input").prop('checked', false);
	};
	$("#task-checksum-input").val( calculateTaskChecksum() );
	$("#task-submit-button").html(buttonSaveTask);
	$("#operation-status-label").text("");
	manageListsButtons();
	$("#task-text-input").focus();
}

// ===========================================================================
// Reset task form and prepare it to create new task.
// ===========================================================================
function newTask() {
	$("#task-status-select").val("created");
	$("#task-icon-select").val("");
	$("#task-section-select").val( $(this).parents("section").attr("id") );
	$("#task-text-input").val("");
	$("#task-id-input").val("");
	var now = new Date();
	$("#task-timestamp-input").val(now.toJSON());
	$("#checkbox-today-input").prop('checked', false);
	$("#task-checksum-input").val( calculateTaskChecksum() );
	$("#task-submit-button").html(buttonAddTask);
	//$("#operation-status-label").text(""); // not necessary
	manageListsButtons();
	$("#task-text-input").focus();
}

// ===========================================================================
// Calculate checksum of current task in task-form
// ===========================================================================
function calculateTaskChecksum() {
	return md5(
		$("#task-text-input").val()+
		$("#task-section-select").val()+
		$("#task-icon-select").val()+
		$("#task-status-select").val()
	);
}

// ===========================================================================
// Send current edited task (existing or new) to server and database in current list
// ===========================================================================
function submitTaskOnCurrentList() {
	// If no task text - then nothing to do
	if ( !$("#task-text-input").val() ) { 
		$("#operation-status-label").html(resultTaskEmpty); 
		$("#task-text-input").focus();
		return false
	};
	// Is any field of the task-form modified?
	if ( calculateTaskChecksum() == $("#task-checksum-input").val() ) {
		// Nothing changed
		// Update today's todo-list from the current task
		updateTodaysTasksList($("#task-id-input").val());
		// Refresh events handlers
		setEvents();
		// Get ready to create a new task
		$("#ib button").click();
		return false
	}
	// Prepare task struct
	var task = {};
	task.List = $("#task-lists-select").val();
	task.Id = $("#task-id-input").val();
	task.Text = $("#task-text-input").val().trim();
	task.Section = $("#task-section-select").val();
	task.Status = $("#task-status-select").val();
	task.Icon = $("#task-icon-select").val();
	task.Timestamp = $("#task-timestamp-input").val();
	// Send Ajax POST request
	$("#operation-status-label").text("");
	showSpinner("#task-spinner-div");
	$.ajax( {
		url : "/SendTask",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( task ),
		// if success
		success: function (response) {
			hideSpinner("#task-spinner-div");
			switch(response.Result) {
  				case "SessionEmptyNotFoundOrExpired" :
  					Cookies.remove('User-Session');
  					location.reload();
					break;
				case "TaskEmpty" :
					$("#operation-status-label").html(resultTaskEmpty);
					break;
				case "InvalidListName" :
					$("#operation-status-label").html(resultInvalidListName);
					break;
				case "UpdatedTaskNotFound" :
					$("#operation-status-label").html(resultUpdatedTaskNotFound);
					break;
				case "UpdateFailed" :
					$("#operation-status-label").html(resultUpdateFailed);
					break;
				case "InsertFailed" :
					$("#operation-status-label").html(resultInsertFailed);
					break;
				case "TaskJustUpdated" :
					$("#operation-status-label").html(resultTaskJustUpdated);
					loadTasks();
					break;
				case "TaskUpdated" :
				case "TaskInserted" :
					// Update ID and timestamp
					task.Id = response.Id;
					task.Timestamp = response.Timestamp;
					$("#tasks-main").attr("data-timestamp", response.Timestamp);
					// Add or modify task
					if (response.Result == "TaskUpdated") {
						// Update status label
						$("#operation-status-label").html(resultTaskUpdated);
						// Modify existing task
						$("#"+task.Id).replaceWith(htmlPTask(task));
						if (task.Section != $("#"+task.Id).parents("section").attr("id")) {
							// Move task <p> to another section
							$("#"+task.Id).appendTo($("#"+task.Section));
						};
					} else if (response.Result == "TaskInserted") {
						// Update status label
						$("#operation-status-label").html(resultTaskInserted);
						// Add new task to section
						$("#"+task.Section).append(htmlPTask(task));
					};
					// Update today's todo-list from the current task
					updateTodaysTasksList(task.Id);
					// Re-apply filter
					applyFilter($("#filter-input").val());
					// Calculate page statistic
					calculateStatistic();
					// Refresh events handlers
					setEvents();
					// Get ready to create a new task
					$("#ib button").click();
					break
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
// Move selected task to the most recent todo-list
// ===========================================================================
function moveTaskToNewList() {
	// If no task text - then nothing to do
	if ( !$("#task-text-input").val() ) { 
		$("#operation-status-label").html(resultTaskEmpty); 
		$("#task-text-input").focus();
		return false
	};
	// Prepare task struct
	var task = {};
	task.Id = $("#task-id-input").val();
	task.ToList = $("#task-lists-select").find("option:first-child").val();
	task.Text = $("#task-text-input").val().trim();
	task.Section = $("#task-section-select").val();
	task.Status = $("#task-status-select").val();
	task.Icon = $("#task-icon-select").val();
	task.Timestamp = $("#task-timestamp-input").val();
	task.Today = $("#checkbox-today-input").prop('checked');
	// Send Ajax POST request
	$("#operation-status-label").text("");
	showSpinner("#task-spinner-div");
	$.ajax( {
		url : "/MoveTask",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( task ),
		// if success
		success: function (response) {
			hideSpinner("#task-spinner-div");
			switch(response.Result) {
  				case "SessionEmptyNotFoundOrExpired" :
  					Cookies.remove('User-Session');
  					location.reload();
					break;
				case "TaskEmpty" :
					$("#operation-status-label").html(resultTaskEmpty);
					break;
				case "InvalidListName" :
					$("#operation-status-label").html(resultInvalidListName);
					break;
				case "UpdatedTaskNotFound" :
					$("#operation-status-label").html(resultUpdatedTaskNotFound);
					break;
				case "UpdateFailed" :
					$("#operation-status-label").html(resultUpdateFailed);
					break;
				case "InsertFailed" :
					$("#operation-status-label").html(resultInsertFailed);
					break;
				case "TodaysTaskListUpdateFailed" :
					$("#operation-status-label").html(resultTodaysTaskListUpdateFailed);
					break;
				case "TaskJustUpdated" :
					$("#operation-status-label").html(resultTaskJustUpdated);
					loadTasks();
					break;
				case "TaskMoved" :
					// Update timestamp
					task.Timestamp = response.Timestamp;
					$("#tasks-main").attr("data-timestamp", response.Timestamp);
					// Update status label
					$("#operation-status-label").html(resultTaskMoved);
					// Modify existing task
					$("#"+task.Id).attr("class","moved");
					$("#"+task.Id).attr("data-tooltip", statusMoved);
					$("#"+task.Id).attr("data-timestamp", response.Timestamp) ;
					$("#"+task.Id+" img").removeClass("created").removeClass("done").removeClass("canceled").addClass("moved");
					// Update today's todo-list from the current task
					if ( $("div#div-"+task.Id).length > 0 ) {
						$("div#div-"+task.Id).parents("li").replaceWith(htmlLiTask(task.Id));
					}
					// Re-apply filter
					applyFilter($("#filter-input").val());
					// Calculate page statistic
					calculateStatistic();
					// Refresh events handlers
					setEvents();
					// Get ready to create a new task
					$("#ib button").click();
					break
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
// 	Calculate statistics of tasks
// ===========================================================================
function calculateStatistic() {
	$("#total-tasks-count-label").text($("p").length+":");
	if ( $("p").length == 0 )
	{
		$("#done-tasks-count-label").text("0,");
		$("#canceled-tasks-count-label").text("0");
		$("#remaining-tasks-count-label").text("0");
		$("#wait-remind-tasks-count-label").text("0");
		$("#activity-tasks-count-label").text("0");
	}
	else
	{
		var total = $("p").length
		var cnt = $("p.done").length
		$("#done-tasks-count-label").text(cnt+" ("+Math.round(cnt*100/total)+"%),");
		cnt = $("p.canceled").length
		$("#canceled-tasks-count-label").text(cnt+" ("+Math.round(cnt*100/total)+"%),");
		cnt = $("img.wait, img.remind").closest("p.created").length
		$("#wait-remind-tasks-count-label").text(cnt+" ("+Math.round(cnt*100/total)+"%),");
		cnt = $("p.created").length - cnt
		$("#activity-tasks-count-label").text(cnt+" ("+Math.round(cnt*100/total)+"%)");
	}
}

// ===========================================================================
// Refresh events handlers after updating dynamics page objects
// ===========================================================================
function setEvents() {
	$("p").unbind('click');
	$("p").bind('click',onTaskEdit)
	$("div.today-task").unbind('click');
	$("div.today-task").bind('click', onTaskEdit);
	$("img.insert-delimiter").unbind('click');
	$("img.insert-delimiter").bind('click', insertDelimiter);
	$("img.delete-delimiter").unbind('click');
	$("img.delete-delimiter").bind('click', deleteDelimiter);
}

// ===========================================================================
// Restarts the timer, used to check for an update.
// ===========================================================================
function resetTimer() {
	clearTimeout(timerIdle);
	timerIdle = setTimeout(checkNeedUpdate, 5*60*1000);  // time is in milliseconds
}

// ===========================================================================
// Checks if an update tasks or today's tasks list is needed.
// ===========================================================================
function checkNeedUpdate() {
	// Send Ajax POST request
	$.ajax( {
		url : "/NeedUpdate",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( {
			List: $("#task-lists-select").val(),
			LastModifiedTimestamp: $("#tasks-main").attr("data-timestamp"),
			TodayTasksTimestamp: $("#today-tasks-ul").attr("data-timestamp")
		} ),
		// if success
		success: function (response) {
			switch(response.Result) {
  				case "SessionEmptyNotFoundOrExpired" :
  					Cookies.remove('User-Session');
  					location.reload();
					break;
				case "InvalidListName" :
					$("#operation-status-label").html(resultInvalidListName);
					break;
				case "AllActual" :
					resetTimer();
					break;
				case "NeedUpdate" :
					loadTasks();
					resetTimer();
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
