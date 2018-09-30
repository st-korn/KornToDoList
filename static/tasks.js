$( init );

// ===========================================================================
// Page initialization
// ===========================================================================
function init() {
	// Bind event handling
	$("#language-select").change(onLanguageChange);
	$("#wellcome-register-user-button").bind('click', showSignupForm);
	$("#wellcome-login-user-button").bind('click', showLoginForm);
	$("#wellcome-start-anonymously-button").bind('click', startAnonymously);
	$("#help-label").bind('click', showHelp);
	$("#help-close-button").bind('click', hideHelp);
	$("#submit-signup-button").bind('click', submitSignup);
	$("#cancel-signup-button").bind('click', cancelSignup);

	// Analyze the presence of a saved session
	if ( Cookies.get('User-session') == null )
	{
		// Hide and disable forms before user select one of the welcome-form buttons
		$("#user-form").hide();
		$("#select-list-form").hide();
		$("#task-search-div *").prop( "disabled", true );
		$("td button").prop( "disabled", true );
		$("#help-label").hide();
		$("#help-close-div").hide();
		$("#signup-div").hide();
		$("#login-div").hide();
		// Fade out main task table
		$("#task-search-div").fadeTo("slow",0.5);
		$("#tasks-table").fadeTo("slow",0.5);
	}
	else
	{
		// Hide controls lines
		$("#help-div").hide();
		$("#welcome-div").hide();
		$("#signup-div").hide();
		$("#login-div").hide();

		// Validate "User-session" cookie

	}

	// Start working with page
	afterUpdate();
	$("td.vh button").click();
	$("#task-spinner-div").hide();
}

// ===========================================================================
// Show sign-up form
// ===========================================================================
function showSignupForm() {
	// Close login <div> if necessary
	if ( $("#login-div").is(":visible") ) {
		$("#login-div").animate({height: "hide"}, 100);
	};
	// Show sign-up <div>
	$("#signup-spinner-div").hide();
	$("#signup-result-label").text("");
	$("#signup-div").animate({height: "show"}, 100);
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
	$("#signup-spinner-div").show();
	$.ajax( {
		url : "/SignUp",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( { EMail: $("#email-signup-input").val() } ),
		// if success
		success: function (response) {
			$("#signup-spinner-div").hide();
			switch(response.Result) {
  				case "EMailEmpty" : $("#signup-result-label").html(resultEmptyEMail); break;
  				case "UserJustExistsButEmailSent" : $("#signup-result-label").html(resultAllreadyExist); break;
  				case "UserSignedUpAndEmailSent" : $("#signup-result-label").html(resultSignupOK); break;
				default : $("#signup-result-label").html(resultUnknown);
			}
		},
		// if error returns
		error: function(jqXHR,exception) { 
			$("#signup-spinner-div").hide();
			showAjaxError("#signup-result-label",jqXHR,exception);
		}
	} );
	return false;
}

// ===========================================================================
// Universal function to parse Ajax errors
// ===========================================================================
function showAjaxError(labelID, jqXHR, exception) {
	var msg = '';
	if (jqXHR.status === 0) { msg = ajax0; } 
	else if (jqXHR.status == 404) { msg = ajax404; }
	else if (jqXHR.status == 500) { msg = ajax500 + ' ' + jqXHR.responseText; }
	else if (exception === 'parsererror') { msg = ajaxParseError; } 
	else if (exception === 'timeout') { msg = ajaxTimeOut; } 
	else if (exception === 'abort') { msg = ajaxAbort; }
	else { msg = ajaxOther + ' ' + jqXHR.responseText; }
	$(labelID).text(msg);
}

// ===========================================================================
// Cancel sign-up form
// ===========================================================================
function cancelSignup() {
	// Close sign-up <div>
	$("#signup-div").animate({height: "hide"}, 100);
	return false;
}

function showLoginForm() {

}

function startAnonymously() {

}

// ===========================================================================
// Show to user Help <div>
// ===========================================================================
function showHelp() {
	$("#help-close-div").show();
	$("#help-div").animate({height: "show"}, 100);
	$("#help-label").hide();
}

// ===========================================================================
// Show from user Help <div>
// ===========================================================================
function hideHelp() {
	$("#help-div").animate({height: "hide"}, 100);
	$("#help-label").show();
	return false;
}

// ===========================================================================
// When the user selects a language from the list
// ===========================================================================
function onLanguageChange() {
	Cookies.set('User-Language', $("#language-select").val(), { expires: 365 });
	location.reload();
}

// ===========================================================================
// Initializing page objects after updating the task table from the server
// ===========================================================================
function afterUpdate() {
	// Calculate statistics
	$("#total-tasks-count-label").text($("p").length+":");
	if ( $("p").length == 0 )
	{
		$("#done-tasks-count-label").text("0,");
		$("#canceled-tasks-count-label").text("0");
	}
	else
	{
		$("#done-tasks-count-label").text($("p.done").length+" ("+Math.round($("p.done").length*100/$("p").length)+"%),");
		$("#canceled-tasks-count-label").text($("p.canceled").length+" ("+Math.round($("p.canceled").length*100/$("p").length)+"%)");
	}
}

