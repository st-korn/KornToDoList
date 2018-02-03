$( init );
function init() {
	$('td button').bind('click', addNewTask);
	$('p').bind('click', editTask);
	$("form button").bind('click', sendTask);
	$("td.vh button").click();
}

function addNewTask() {
	$("form[name='task'] label.action").text("Добавить задачу: ");
	$("input[name='id']").val( "" );
	$("input[name='text']").val( "" );
	$("select[name='section']").val( $(this).closest('td').attr('class') );
	$("select[name='status']").val( "created" );
	$("form[name='task'] button").text("Добавить");
	$("form[name='task'] label.result").text("");
	$("input[name='text']").focus();
}

function editTask() {
	$("form[name='task'] label.action").text("Редактировать задачу: ");
	$("input[name='id']").val( $(this).attr('id') );
	$("input[name='text']").val( $(this).text() );
	$("select[name='section']").val( $(this).closest('td').attr('class') );
	$("select[name='status']").val( $(this).attr('class') );
	$("form[name='task'] button").text("Изменить");
	$("form[name='task'] label.result").text("");
	$("input[name='text']").focus();
}

function sendTask() {
	$.ajax( {
		url : "/sendTask",
		cache: false,
		type : "post",
		dataType: "text",
		data : $("form").serialize(),
		success: function (response) {
			// После того, как задача будет успешно отправлена на сервер
			//console.log(response);
			//var json_obj = $.parseJSON(response);
			//$("form[name='task'] label.result").html(json_obj._id);
			$("form[name='task'] label.result").html("Готово");
			// Обновляем таблицу задач
			$("table").load(location.href+" #data",
				"",
				function() { 
					// Восстанавливаем функциональность кнопок "+" и кликабельность задач
					$('td button').bind('click', addNewTask);
					$('p').bind('click', editTask);
				}
			);
		},
		error: function (jqXHR, exception) {
	        var msg = '';
			if (jqXHR.status === 0) { msg = 'Not connect.\n Verify Network.'; } 
			else if (jqXHR.status == 404) { msg = 'Requested page not found. [404]'; }
			else if (jqXHR.status == 500) { msg = 'Internal Server Error [500].'+jqXHR.responseText; }
			else if (exception === 'parsererror') { msg = 'Requested JSON parse failed.'; } 
			else if (exception === 'timeout') { msg = 'Time out error.'; } 
			else if (exception === 'abort') { msg = 'Ajax request aborted.'; }
			else { msg = 'Uncaught Error.\n' + jqXHR.responseText; }
			$("form[name='task'] label.result").html(msg);
		},
	} );
	return false;
}
