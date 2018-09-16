$( init );

// ===========================================================================
// Page initialization
// ===========================================================================
function init() {
	// Bind event handling
	$("select#language-select").change(onLanguageChange);
	afterUpdate();
	$("td.vh button").click();
	$("#spinner").hide();
}

// ===========================================================================
// When the user selects a language from the list
// ===========================================================================
function onLanguageChange() {
	Cookies.set('User-Language', $("select#language-select").val());
	location.reload();
}

// ===========================================================================
// Initializing page objects after updating the task table from the server
// ===========================================================================
function afterUpdate() {
	//$('button.add').bind('click', addNewTask);
	//$("button#ClearDone").bind('click', clearDoneTodayTasks);
	//$("button#ClearAll").bind('click', clearAllTodayTasks);
	//$('p').bind('click', editTask);

	// Calculate statistics
	$("label#total-tasks-count-label").text($("p").length+":");
	if ( $("p").length == 0 )
	{
		$("label#done-tasks-count-label").text("0,");
		$("label#canceled-tasks-count-label").text("0");
	}
	else
	{
		$("label#done-tasks-count-label").text($("p.done").length+" ("+Math.round($("p.done").length*100/$("p").length)+"%),");
		$("label#canceled-tasks-count-label").text($("p.canceled").length+" ("+Math.round($("p.canceled").length*100/$("p").length)+"%)");
	}
}

