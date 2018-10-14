// ===========================================================================
// Manage visibility and availability of buttons "Create new task list" and "Move task to new list"
// ===========================================================================
function manageListsButtons() {
	// Determines the ability to create a new task list
	var dateYYYYMMDD = getCurrentDate();
	if ($('#task-lists-select option[value="'+dateYYYYMMDD+'"]').length == 0)
	{
		$("#new-list-create-button").show();
	} else {
		$("#new-list-create-button").hide();
	};
	// Determines the ability to move a task to another task list
	if ($("#task-lists-select").prop("selectedIndex") == 0) {
		$("#task-move-button").hide();
	} else {
		if ($("#task-id-input").val() != "") {
			if ($("#task-status-select").val() == "created") {
				$("#task-move-button").show();
			} else {
				$("#task-move-button").hide();
			}
		} else	{
			$("#task-move-button").hide();
		}
	};
}


// ===========================================================================
// When the user selects a list from the task-lists
// ===========================================================================
function onTaskListChange() {
    $("#operation-status-label").text("");
    loadTasks();
}

// ===========================================================================
// Fetch names lists of current user from server and put in in dropdown-list
// Called once during page loading and initialization.
// ===========================================================================
function getLists() {
	// Send Ajax POST request
	$("#operation-status-label").text("");
	showSpinner("#task-spinner-div");
	$.ajax( {
		url : "/GetLists",
		cache: false,
		type : "post",
		// if success
		success: function (response) {
			hideSpinner("#task-spinner-div");
			switch(response.Result) {
  				case "SessionEmptyNotFoundOrExpired" :
  					Cookies.remove('User-Session');
  					location.reload();
  					break;
				case "OK" :
					// Collect lists names
					$("#task-lists-select option").remove();
					$.each(response.Lists, function() {
						$("#task-lists-select").append($("<option />").val(this).text(this));
					});
					// If no task-list is found, create a new task-list.
					if ($("#task-lists-select option").length == 0) {
						newList();
						return false;
					 };
					// Select list from the form URL
					if (openList) {
						$("#task-lists-select").val(openList);
						if (!$("#task-lists-select").val()) {
							// List not found - select first list
							$("#task-lists-select").prop("selectedIndex",0);
						}
						openList = "";
					} else {
						// Select first list
						$("#task-lists-select").prop("selectedIndex",0);
					};
					// Load all tasks of current list
					loadTasks();
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
// Create new todo-list named by current date
// ===========================================================================
function newList() {
	// Send Ajax POST request
	$("#operation-status-label").text("");
	showSpinner("#task-spinner-div");
	$.ajax( {
		url : "/CreateList",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( { List: getCurrentDate() } ),
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
				case "CreateListFailed" :
					$("#operation-status-label").html(resultCreateListFailed);
					break;
				case "ListCreated" :
					$("#operation-status-label").html(resultListCreated);
					// Collect lists names
					$("#task-lists-select option").remove();
					$.each(response.Lists, function() {
						$("#task-lists-select").append($("<option />").val(this).text(this));
					});
					// Select first list
					$("#task-lists-select").prop("selectedIndex",0);
					// Load all tasks of current list
					loadTasks();
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