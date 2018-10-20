// Default cookies lifetime interval (days)
const DefaultCookieLifetimeDays = 365

// Bind event handling
$( function() {$("#language-select").change(onLanguageChange);} );

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

// ===========================================================================
// New jquery selector for case-insesitive comparison
// ===========================================================================
jQuery.expr[':'].Contains = function(a, i, m) {
	return jQuery(a).text().toLowerCase()
		.indexOf(m[3].toLowerCase()) >= 0;
  };

// ===========================================================================
// Function for array unique sorting
// ===========================================================================
function onlyUnique(value, index, self) { 
    return self.indexOf(value) === index;
}

// ===========================================================================
// Function for sorting array of strings alphabetically
// ===========================================================================
function sortStringsAlphabetically(a, b) {
    if(a.toLowerCase() < b.toLowerCase()) return -1;
    if(a.toLowerCase() > b.toLowerCase()) return 1;
	return 0;
}

// ===========================================================================
// Function for hiding and showing spinners by its jquery-selector
// ===========================================================================
function hideSpinner(selector) {
	$(selector).css('display', 'none');
}

function showSpinner(selector) {
	$(selector).css('display', 'inline-block');
}

// ===========================================================================
// Function return current date in format YYYY-MM-DD
// ===========================================================================
function getCurrentDate() {
	var today = new Date();
	var dd = today.getDate();
	var mm = today.getMonth()+1; //January is 0!
	var yyyy = today.getFullYear();
	if(dd<10) { dd = '0'+dd } 
	if(mm<10) { mm = '0'+mm } 
	today = yyyy + '-' + mm + '-' + dd;
	return today;
}

// ===========================================================================
// Extract icon class name from several class names of <img> tag
// ===========================================================================
function extractIcon(classes) {
	if (classes) {
		return classes.replace("icon","").replace("created","").replace("done","").replace("canceled","").replace("moved","").trim();
	} else {
		return "";
	}
}

// ===========================================================================
// When the user selects a language from the list
// ===========================================================================
function onLanguageChange() {
	Cookies.set('User-Language', $("#language-select").val(), { expires: DefaultCookieLifetimeDays });
	location.reload();
}