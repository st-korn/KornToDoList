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
		$("p:not(:Contains("+ filter.trim() +"))").hide(); 
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
		applyFilter($("#filter-input").val());
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