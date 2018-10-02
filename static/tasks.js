$( init );

// ===========================================================================
// Page initialization
// ===========================================================================
function init() {
	// Bind event handling
	$("#language-select").change(onLanguageChange);
	$("#wellcome-register-user-button").bind('click', clickSignup);
	$("#wellcome-login-user-button").bind('click', showLoginForm);
	$("#wellcome-start-anonymously-button").bind('click', startAnonymously);
	$("#help-label").bind('click', showHelp);
	$("#help-close-button").bind('click', hideHelp);
	$("#submit-signup-button").bind('click', submitSignup);
	$("#submit-restore-button").bind('click', submitSignup);
	$("#cancel-signup-button").bind('click', cancelSignup);
	$("#submit-login-button").bind('click', submitLogin);
	$("#cancel-login-button").bind('click', cancelLogin);
	$("#logout-user-button").bind('click', clickLogout);
	$("#forgot-password-label").bind('click', clickForgotPassword);

	// Analyze the presence of a saved session
	if ( Cookies.get('User-Session') == null )
	{
		// Hide and disable forms before user select one of the welcome-form buttons
		$("#user-form").hide();
		$("#select-list-form").hide();
		$("#task-search-nav *").prop( "disabled", true );
		$("td button").prop( "disabled", true );
		$("#help-label").hide();
		$("#help-close-div").hide();
		$("#signup-header").hide();
		$("#login-header").hide();
		// Fade out main task table
		$("#task-search-nav").fadeTo("slow",0.5);
		$("#tasks-main").fadeTo("slow",0.5);
	}
	else
	{
		// Hide controls lines
		$("#help-header").hide();
		$("#welcome-header").hide();
		$("#signup-header").hide();
		$("#login-header").hide();

		// Hide user-buttons
		$("#user-img").hide();
		$("#anon-img").hide();
		$("#register-user-button").hide();

		// Validate "User-Session" cookie
		getUserInfo();

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
	if ( $("#login-header").is(":visible") ) {
		$("#login-header").animate({height: "hide"}, 100);
	};
	// Show sign-up <div>
	$("#signup-spinner-div").hide();
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
  				case "EmptyEMail" : $("#signup-result-label").html(resultEmptyEMail); break;
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
	$("#login-spinner-div").hide();
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
	$("#login-spinner-div").show();
	$.ajax( {
		url : "/LogIn",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( { EMail: $("#email-login-input").val(), PasswordMD5: hash } ),
		// if success
		success: function (response) {
			$("#login-spinner-div").hide();
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
			$("#login-spinner-div").hide();
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
	$("#task-spinner-div").show();
	$.ajax( {
		url : "/GoAnonymous",
		cache: false,
		type : "post",
		// if success
		success: function (response) {
			$("#task-spinner-div").hide();
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
			$("#task-spinner-div").hide();
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
	$("#task-spinner-div").show();
	$.ajax( {
		url : "/LogOut",
		cache: false,
		type : "post",
		// if success
		success: function (response) {
			$("#task-spinner-div").hide();
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
			$("#task-spinner-div").hide();
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
	$("#task-spinner-div").show();
	$.ajax( {
		url : "/UserInfo",
		cache: false,
		type : "post",
		// if success
		success: function (response) {
			$("#task-spinner-div").hide();
			switch(response.Result) {
  				case "EmptySession" : break;
  				case "SessionNotFoundOrExpired" :
  					Cookies.remove('User-Session');
  					location.reload();
  					break;
  				case "ValidUserSession" :
  					$("#user-img").show();
  					$("#anon-img").hide();
  					$("#user-label").html(response.EMail);
  					$("#register-user-button").hide();
  					break;
  				case "ValidAnonymousSession" : 
  					$("#user-img").hide();
  					$("#anon-img").show();
  					$("#user-label").html(response.EMail);
  					$("#register-user-button").show();
  					break;
				default : $("#operation-status-label").html(resultUnknown);
			}
		},
		// if error returns
		error: function(jqXHR,exception) { 
			$("#task-spinner-div").hide();
			showAjaxError("#operation-status-label",jqXHR,exception);
		}
	} );
	return false;
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

