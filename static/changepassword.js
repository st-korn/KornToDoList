$( init );

// ===========================================================================
// Page initialization
// ===========================================================================
function init() {
	// Bind event handling
	$("#submit-button").bind('click', submitPasswords);
	$("#return-button").bind('click', returnBack);

	// If Result is already known
	if ( textResult != "") {
		$("#password1-div").hide();
		$("#password2-div").hide();
		$("#submit-div").hide();
	}
	else
	{
		$("#password1-input").focus();
	}
}

// ===========================================================================
// Submit SetPassword form
// ===========================================================================
function submitPasswords() {
	// Validate fields
	if ( $("#password1-input").val() != $("#password2-input").val() ) {
		$("#result-label").text(resultPasswordsNotEquals);
		return false;
	}
	if ( $("#password1-input").val() == "" ) {
		$("#result-label").text(resultEmptyPassword);
		return false;
	}

	// Calculate password's hash
	var hash = md5($("#password1-input").val());

	// Send Ajax POST request
	$("#result-label").text("");
	showSpinner("#spinner-div");
	$.ajax( {
		url : "/SetPassword",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( { UUID: uuid, PasswordMD5: hash } ),
		// if success
		success: function (response) {
			hideSpinner("#spinner-div");
			switch(response.Result) {
				case "UUIDExpiredOrNotFound" : $("#result-label").html(resultUUIDExpiredOrNotFound); break;
  				case "EmptyPassword" : $("#result-label").html(resultEmptyPassword); break;
  				case "UserCreated" : $("#result-label").html(resultUserCreated); window.location.href = "/"; break;
  				case "PasswordUpdated" : $("#result-label").html(resultPasswordUpdated); window.location.href = "/"; break;
				default : $("#result-label").html(resultUnknown);
			}
		},
		// if error returns
		error: function(jqXHR,exception) { 
			hideSpinner("#spinner-div");
			showAjaxError("#result-label",jqXHR,exception);
		}
	} );
	return false;
}

// ===========================================================================
// Return back to main page
// ===========================================================================
function returnBack() {
	window.location.href = "/";
	return false;
}