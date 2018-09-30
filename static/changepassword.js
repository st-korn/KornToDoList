$( init );

// ===========================================================================
// Page initialization
// ===========================================================================
function init() {
	// Bind event handling
	$("#submit-button").bind('click', submitPasswords);
	$("#return-button").bind('click', returnBack);

	// Hide spinner
	$("#spinner-div").hide();

	// If Result is already known
	if ( textResult != "") {
		$("#password1-div").hide();
		$("#password2-div").hide();
		$("#submit-div").hide();
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
	$("#spinner-div").show();
	$.ajax( {
		url : "/SetPassword",
		cache: false,
		type : "post",
		dataType: "json",
		contentType: "application/json; charset=utf-8",
		data : JSON.stringify( { uuid: uuid, password: hash } ),
		// if success
		success: function (response) {
			$("#spinner-div").hide();
			switch(response.Result) {
  				case "PasswordsNotEquals" : $("#result-label").html(resultPasswordsNotEquals); break;
  				case "EmptyPassword" : $("#result-label").html(resultEmptyPassword); break;
  				case "PasswordUpdated" : $("#result-label").html(resultPasswordUpdated); break;
				default : $("#result-label").html(resultUnknown);
			}
		},
		// if error returns
		error: function(jqXHR,exception) { 
			$("#spinner-div").hide();
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