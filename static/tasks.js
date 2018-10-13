// Initialze
var timerIdle;
$( init );

// ===========================================================================
// Page initialization
// ===========================================================================
function init() {
	// Bind event handling
	$("#language-select").change(onLanguageChange);
	$("#welcome-register-user-button").bind('click', clickSignup);
	$("#welcome-login-user-button").bind('click', showLoginForm);
	$("#welcome-start-anonymously-button").bind('click', startAnonymously);
	$("#help-label").bind('click', showHelp);
	$("#help-close-button").bind('click', hideHelp);
	$("#submit-signup-button").bind('click', submitSignup);
	$("#submit-restore-button").bind('click', submitSignup);
	$("#cancel-signup-button").bind('click', cancelSignup);
	$("#submit-login-button").bind('click', submitLogin);
	$("#cancel-login-button").bind('click', cancelLogin);
	$("#logout-user-button").bind('click', clickLogout);
	$("#forgot-password-label").bind('click', clickForgotPassword);
	$("#filter-select").change( function() { applyFilter($("#filter-input").val()) } );
	$("#filter-clear-button").bind('click', clickClearFilter);
	$("#filter-input").keypress(onEnterFilterInput);
	$("#task-lists-select").change(loadTasks);
	$("#task-submit-button").bind('click', submitTaskOnCurrentList);
	$("button.add").bind('click', newTask);
	$("#new-list-create-button").bind('click', newList);
	$("#task-move-button").bind('click', moveTaskToNewList);
	$("#register-user-button").bind('click', clickSignup);
	$("#clear-done-button").bind('click', clearDoneTasksFromToday)
	$("#clear-all-button").bind('click', clearAllTasksFromToday)

	// Bind handlers for idle timer
    window.onload = resetTimer;
    window.onmousemove = resetTimer;
    window.onmousedown = resetTimer;  // catches touchscreen presses as well      
    window.ontouchstart = resetTimer; // catches touchscreen swipes as well 
    window.onclick = resetTimer;      // catches touchpad clicks as well
    window.onkeypress = resetTimer;   
    window.addEventListener('scroll', resetTimer, true); // improved; see comments

	// Fill autocomplete with names of employees
	$("#filter-input").autocomplete( {
		source: completeEmployeeNames,
		select: function (e, ui) { applyFilter(ui.item.value);  },
		change: function (e, ui) { applyFilter(this.value); },
		open: function(event, ui) {
			$('.ui-autocomplete').off('menufocus hover mouseover mouseenter');
		}
	  } );

	// Initialize sortable today's task-list
	$("#today-tasks-ul").sortable( {
		placeholder: "ui-state-highlight",
		handle: ".handle",
		update: function (e, ui) { saveToday(); }
	});

	// Analyze the presence of a saved session
	if ( Cookies.get('User-Session') == null )
	{
		// Show welcome-headers
		$("#help-header").show();
		$("#welcome-header").show();
		// Hide and disable forms before user select one of the welcome-form buttons
		$("#task-filter-nav *").prop( "disabled", true );
		$("section button").prop( "disabled", true );
		// Fade out main task table
		$("#task-filter-nav").fadeTo("slow",0.5);
		$("#tasks-main").fadeTo("slow",0.5);
		// Hide loading spinner
		hideSpinner("#task-spinner-div");
	}
	else
	{
		// show severals hidden forms
		$("#user-form").show();
		$("#select-list-form").show();
		$("#help-label").show();
		$("#help-close-div").show();
		// Validate "User-Session" cookie
		getUserInfo();
		// Load list of todo-lists for current user
		getLists();
	}
}

// ===========================================================================
// Show sign-up form
// ===========================================================================
function showSignupForm() {
	// Close login <div> if necessary
	if ( $("#login-header").is(":visible") ) {
		$("#login-header").animate({height: "hide"}, 100);
	};
	// Show sign-up <div>
	hideSpinner("#signup-spinner-div");
	$("#signup-result-label").text("");
	$("#signup-header").animate({height: "show"}, 100);
	$("#email-signup-input").focus();
	return false;
}

// ===========================================================================
// Submit sign-up form
// ===========================================================================
function submitSignup() {
	// Validate fields
	if ( $("#email-signup-input").val() == "" ) {
		$("#signup-result-label").text(resultEmptyEMail);
		return false;
	}
	// Send Ajax POST request
	$("#signup-result-label").text("");
	showSpinner("#signup-spinner-div");
	$.ajax( {
		url : "/SignUp",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( { EMail: $("#email-signup-input").val() } ),
		// if success
		success: function (response) {
			hideSpinner("#signup-spinner-div");
			switch(response.Result) {
  				case "EmptyEMail" : $("#signup-result-label").html(resultEmptyEMail); break;
  				case "UserJustExistsButEmailSent" : $("#signup-result-label").html(resultAllreadyExist); break;
  				case "UserSignedUpAndEmailSent" : $("#signup-result-label").html(resultSignupOK); break;
				default : $("#signup-result-label").html(resultUnknown);
			}
		},
		// if error returns
		error: function(jqXHR,exception) { 
			hideSpinner("#signup-spinner-div");
			showAjaxError("#signup-result-label",jqXHR,exception);
		}
	} );
	return false;
}

// ===========================================================================
// Cancel signup form
// ===========================================================================
function cancelSignup() {
	// Close sign-up <div>
	$("#signup-header").animate({height: "hide"}, 100);
	return false;
}

// ===========================================================================
// Let user try to sign up
// ===========================================================================
function clickSignup() {
	$("#submit-signup-button").show();
	$("#submit-restore-button").hide();
	showSignupForm();
	return false;
}

// ===========================================================================
// Show login form
// ===========================================================================
function showLoginForm() {
	// Close sign-up <div> if necessary
	if ( $("#signup-header").is(":visible") ) {
		$("#signup-header").animate({height: "hide"}, 100);
	};
	// Show login <div>
	hideSpinner("#login-spinner-div");
	$("#login-result-label").text("");
	$("#login-header").animate({height: "show"}, 100);
	$("#email-login-input").focus();
	return false;
}

// ===========================================================================
// Submit login form
// ===========================================================================
function submitLogin() {
	// Validate fields
	if ( $("#email-login-input").val() == "" ) {
		$("#login-result-label").text(resultEmptyEMail);
		return false;
	}
	if ( $("#password-login-input").val() == "" ) {
		$("#login-result-label").text(resultEmptyPassword);
		return false;
	}
	// Calculate password's hash
	var hash = md5($("#password-login-input").val());
	// Send Ajax POST request
	$("#login-result-label").text("");
	showSpinner("#login-spinner-div");
	$.ajax( {
		url : "/LogIn",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( { EMail: $("#email-login-input").val(), PasswordMD5: hash } ),
		// if success
		success: function (response) {
			hideSpinner("#login-spinner-div");
			switch(response.Result) {
  				case "EmptyEMail" : $("#login-result-label").html(resultEmptyEMail); break;
  				case "EmptyPassword" : $("#login-result-label").html(resultEmptyPassword); break;
  				case "UserAndPasswordPairNotFound" : $("#login-result-label").html(resultUserAndPasswordPairNotFound); break;
  				case "LoggedIn" : 
  					$("#login-result-label").html(resultLoggedIn); 
  					Cookies.set('User-Session', response.UUID, { expires: DefaultCookieLifetimeDays });
  					location.reload();
  					break;
				default : $("#login-result-label").html(resultUnknown);
			}
		},
		// if error returns
		error: function(jqXHR,exception) { 
			hideSpinner("#login-spinner-div");
			showAjaxError("#login-result-label",jqXHR,exception);
		}
	} );
	return false;
}

// ===========================================================================
// Cancel login form
// ===========================================================================
function cancelLogin() {
	// Close sign-up <div>
	$("#login-header").animate({height: "hide"}, 100);
	return false;
}

// ===========================================================================
// Restore forgotten password
// ===========================================================================
function clickForgotPassword() {
	$("#submit-signup-button").hide();
	$("#submit-restore-button").show();
	showSignupForm();
}

// ===========================================================================
// Start to work anonymously
// ===========================================================================
function startAnonymously() {
	// Send Ajax POST request
	$("#operation-status-label").text("");
	showSpinner("#task-spinner-div");
	$.ajax( {
		url : "/GoAnonymous",
		cache: false,
		type : "post",
		// if success
		success: function (response) {
			hideSpinner("#task-spinner-div");
			switch(response.Result) {
  				case "SuccessAnonymous" : 
  					Cookies.set('User-Session', response.UUID, { expires: DefaultCookieLifetimeDays });
  					location.reload();
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
// Show to user Help <div>
// ===========================================================================
function showHelp() {
	$("#help-close-div").show();
	$("#help-header").animate({height: "show"}, 100);
	$("#help-label").hide();
}

// ===========================================================================
// Show from user Help <div>
// ===========================================================================
function hideHelp() {
	$("#help-header").animate({height: "hide"}, 100);
	$("#help-label").show();
	return false;
}

// ===========================================================================
// When the user selects a language from the list
// ===========================================================================
function onLanguageChange() {
	Cookies.set('User-Language', $("#language-select").val(), { expires: DefaultCookieLifetimeDays });
	location.reload();
}

// ===========================================================================
// Click logout button
// ===========================================================================
function clickLogout() {
	// Send Ajax POST request
	$("#operation-status-label").text("");
	showSpinner("#task-spinner-div");
	$.ajax( {
		url : "/LogOut",
		cache: false,
		type : "post",
		// if success
		success: function (response) {
			hideSpinner("#task-spinner-div");
			switch(response.Result) {
  				case "EmptySession" : break;
  				case "LoggedOut" : 
  					Cookies.remove('User-Session');
  					location.reload();
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
// Fetch user info from server and put in on page
// ===========================================================================
function getUserInfo() {
	// Send Ajax POST request
	$("#operation-status-label").text("");
	showSpinner("#task-spinner-div");
	$.ajax( {
		url : "/UserInfo",
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
  				case "ValidUserSession" :
  					$("#user-img").show();
  					$("#anon-img").hide();
  					$("#user-label").html(response.EMail);
  					$("#register-user-button").hide();
  					Cookies.set('User-Session', Cookies.get('User-Session'), { expires: DefaultCookieLifetimeDays });
  					break;
  				case "ValidAnonymousSession" : 
  					$("#user-img").hide();
  					$("#anon-img").show();
  					$("#user-label").html(response.EMail);
  					$("#register-user-button").show();
  					Cookies.set('User-Session', Cookies.get('User-Session'), { expires: DefaultCookieLifetimeDays });
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
// Generate <p> HTML for the Task structure
// IN: task struct {	ID : string, Text : string, Section : string, Status : string, Icon : string, Timestamp : string }
// ===========================================================================
function htmlPTask(task) {
	var tooltips = {'created': statusCreated, 'moved': statusMoved, 'canceled': statusCanceled, 'done': statusDone};
	var p = '<p id="' + task.Id + '" class="' + task.Status + '" data-tooltip="' + tooltips[task.Status] + '" data-timestamp="'+task.Timestamp+'">'
	if (task.Icon != "") {
		p = p + '<img class="icon '+task.Icon+' '+task.Status+'" src="/static/icons/'+task.Icon+'.svg">'
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
// Generate <li> HTML for the Task structure
// IN: taskId string
// ===========================================================================
function htmlLiTask(taskId) {
	var taskIcon = extractIcon( $("#"+taskId).children("img").attr("class") );
	var taskStatus = $("#"+taskId).attr("class");
	var taskTimestamp = $("#"+taskId).attr("data-timestamp");
	var taskText = $("#"+taskId).text();
	var tooltips = {'created': statusCreated, 'moved': statusMoved, 'canceled': statusCanceled, 'done': statusDone};
	var li = '<li>'
	li = li + '<img class="handle" src="/static/icons/updown.svg">'
	if (taskIcon) {
		li = li + '<img class="icon '+taskIcon+' '+taskStatus+'" src="/static/icons/'+taskIcon+'.svg">'
	}
	li = li + '<div class="today-task '+taskStatus+'" id="div-' + taskId + '" class="' + taskStatus + '" data-tooltip="' + tooltips[taskStatus] + '" data-timestamp="'+taskTimestamp+'">'
	var idx = taskText.indexOf(" - ");
	if (idx > 0) {
		li = li + '<span class="employee">' + taskText.substr(0,idx) + '</span>' + taskText.substring(idx,taskText.length);
	} else {
		li = li + taskText;
	}
	li = li + '</div>'
	li = li + '<img class="insert-delimiter" src="/static/icons/delimiter.svg" title="'+hintInsertDelimiter+'">'
	li = li + '</li>';
	return li;
}

// ===========================================================================
// Generate <li> HTML for delimiter
// ===========================================================================
function htmlLiDelimiter() {
	var li = '<li class="delimiter">';
	li = li + '<img class="handle" src="/static/icons/updown.svg">';
	li = li + '<div class="delimiter"></div>';
	li = li + '<img class="delete-delimiter" src="/static/icons/delete.svg" title="'+hintDeleteDelimiter+'">';
	li = li + '</li>';
	return li;	
}

// ===========================================================================
// Load from server all tasks of current selected list
// ===========================================================================
function loadTasks() {
	// Manage visibility and availability of list buttons
	manageListsButtons();
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
					// Calculate page statistic
					calculateStatistic();
					// Refresh events handlers
					setEvents();
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
// When the user selects a tasks display mode or 
// when the user enter any filter letters
// ===========================================================================
function applyFilter(filter) {
	switch($("#filter-select").val()) {
		case "all" :
			$("p").show();
			break;
		case "created-only" :
			$("p.created").show();
			$("p.done").hide();
			$("p.canceled").hide();
			$("p.moved").hide();
			break;
		case "created-not-wait-not-remind" :
			$("p.done").hide();
			$("p.canceled").hide();
			$("p.moved").hide();
			$("p.created").show();
			$("img.wait, img.remind").closest("p.created").hide();
			break;
	}
	if ( filter != "" ) {
		$("p:not(:Contains("+ filter +"))").hide(); 
	}
}

// ===========================================================================
// When the user enter any filter letters
// ===========================================================================
function clickClearFilter() {
	$('#filter-input').autocomplete('close');
	$("#filter-input").val("");
	applyFilter("");
	return false;
}

// ===========================================================================
// Keyboard event handler on Filter Input field
// ===========================================================================
function onEnterFilterInput(event) {
	if (event.keyCode == 13) {
		event.preventDefault();
		applyFilter();
		$("#filter-clear-button").focus();
	}
}

// ===========================================================================
// Autocompletion function of search filter text input box - searches among employees names
// ===========================================================================
function completeEmployeeNames(request, response) {

	function onlyTerm(value) { 
		return value.toLowerCase().indexOf(request.term.toLowerCase()) > -1;
	}

	var employees = $("span.employee:Contains("+request.term+")").toArray().map( 
						function(elem) { return elem.textContent.split(",").map(
							function(item) { return item.trim() } )
						} );
	var ac = [].concat.apply([], employees);
	ac = ac.filter(onlyTerm).filter(onlyUnique).sort(sortStringsAlphabetically);
	response(ac);
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
	$("#task-icon-select").val( extractIcon( $("p#"+id).children("img").attr("class") ) );
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
// Send current edited task (existing or new) to server and database in the specified list
// ===========================================================================
function submitTask(list) {
	// If no task text - then nothing to do
	if ( !$("#task-text-input").val() ) { 
		$("#operation-status-label").html(resultTaskEmpty); 
		$("#task-text-input").focus();
		return false
	};
	// Is any field of the task-form modified?
	if ( ( calculateTaskChecksum() == $("#task-checksum-input").val() ) && ( list == $("#task-lists-select").val() ) ) {
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
	task.List = list;
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
					if (list == $("#task-lists-select").val()) {
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
					}
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
// Send current edited task (existing or new) to server and database in current list
// ===========================================================================
function submitTaskOnCurrentList() {
	submitTask($("#task-lists-select").val());
	return false;
}

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

// ===========================================================================
// Move selected task to the most recent todo-list
// ===========================================================================
function moveTaskToNewList() {
	var oldID = $("#task-id-input").val();
	$("#task-id-input").val("");
	submitTask($("#task-lists-select").find("option:first-child").val())
	$("#"+oldID).click();
	$("#task-status-select").val("moved");
	submitTask($("#task-lists-select").val());
	return false;
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