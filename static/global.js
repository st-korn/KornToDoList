// Default cookies lifetime interval (days)
const DefaultCookieLifetimeDays = 365

// ===========================================================================
// Universal function to parse Ajax errors
// ===========================================================================
function showAjaxError(labelID, jqXHR, exception) {
	var msg = '';
	if (jqXHR.status == 0) { msg = ajax0; } 
	else if (jqXHR.status == 404) { msg = ajax404; }
	else if (jqXHR.status == 500) { msg = ajax500 + ' ' + jqXHR.responseText; }
	else if (exception == 'parsererror') { msg = ajaxParseError; } 
	else if (exception == 'timeout') { msg = ajaxTimeOut; } 
	else if (exception == 'abort') { msg = ajaxAbort; }
	else { msg = ajaxOther + ' ' + jqXHR.responseText; }
	$(labelID).text(msg);
}