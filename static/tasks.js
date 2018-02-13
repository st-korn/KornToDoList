var resetTaskFromAfterUpdate = false;
$( init );

// ===========================================================================
// Инициализация страницы
// ===========================================================================
function init() {
	$("select#ListSelect").change(onListChange);
	$("button#ListCreate").bind('click', createList);
	$("button#TaskSubmit").bind('click', sendTask);
	$("button#TaskMove").bind('click', moveTask);
	$("button#TaskToday").bind('click', todayTask);
	$("button#TaskFromday").bind('click', fromdayTask);
	afterUpdate();
	$("td.vh button").click();
	$("#spinner").hide();
}

// ===========================================================================
// Инициализация объектов страницы после обновления таблицы задач с сервера
// ===========================================================================
function afterUpdate() {
	$('button.add').bind('click', addNewTask);
	$("button#ClearDone").bind('click', clearDoneTodayTasks);
	$("button#ClearAll").bind('click', clearAllTodayTasks);
	$('p').bind('click', editTask);
	$("ul#TodayTasks").sortable({
		cursor: "move",
    	placeholder: "ui-state-highlight",
    	cancel: ".TodayRelax",
    	//scroll: true,
    	update: function(event,ui) { sendToday() }
    });
    $("ul#TodayTasks").disableSelection();
    $("li.TodayTask").resizable({
    	grid: 20,
    	handles: 's',
    	distance: 3,
    	start: function(event,ui) { $("li.TodayTask").css("cursor", "s-resize") }, 
    	stop: function(event,ui) { $("li.TodayTask").css("cursor", ""); sendToday() }
    });
	$("li.TodayTask").bind('click', editTodayTask);
}

// ===========================================================================
// При щелчке на любую из кнопок "+", расположенных в колонках таблицы
// Подготавливает форму ввода #TaskForm для добавления новой задачи
// ===========================================================================
function addNewTask() {
	resetTaskFromAfterUpdate = false;
	$("input#TaskId").val( "" );
	$("input#TaskList").val(  $("select#ListSelect").val() );
	$("input#TaskText").val( "" );
	$("select#TaskSection").val( $(this).closest('td').attr('class') );
	$("select#TaskStatus").val( "created" );
	$("select#TaskIcon").val( "" );
	$("button#TaskSubmit").text("Добавить");
	$("button#TaskMove").hide();
	$("button#TaskToday").hide();
	$("button#TaskFromday").hide();
	$("label#TaskResult").text("");
	$("input#TaskText").focus();
}

// ===========================================================================
// При щелчке на задаче - загружает её для редактирования в форму #TaskForm
// ===========================================================================
function editTask() {
	resetTaskFromAfterUpdate = false;
	$("input#TaskId").val( $(this).attr('id') );
	$("input#TaskList").val(  $("select#ListSelect").val() );
	$("input#TaskText").val( $(this).text() );
	$("select#TaskSection").val( $(this).closest('td').attr('class') );
	$("select#TaskStatus").val( $(this).attr('class') );
	$("select#TaskIcon").val( $(this).find("img").attr('class') );
	$("button#TaskSubmit").text("Изменить");
	if ( ($("button#TaskMove").attr('title') != $("select#ListSelect").val()) && 
		($("select#TaskStatus").val() == "created") ) 
		{ $("button#TaskMove").show() }
	else
		{ $("button#TaskMove").hide() };
	if( ($("select#TaskStatus").val() == "created") &&
		($('li[id="'+$(this).attr('id')+'"]').length == 0) ) {
		$("button#TaskToday").show() }
	else {
		$("button#TaskToday").hide() };
	if ($('li[id="'+$(this).attr('id')+'"]').length > 0) {
		$("button#TaskFromday").show() }
	else {
		$("button#TaskFromday").hide() };
	$("label#TaskResult").text("");
	$("input#TaskText").focus();
}

// ===========================================================================
// При щелчке на блоке запланированной на сегодня задачи - срабатывает
// так же, как если бы щёлкнули на задаче в одной из колонок таблицы
// ===========================================================================
function editTodayTask() {
	var id = $(this).attr('id')
	var pp = $('p[id="'+id+'"]') // поиск с # не срабатывает, оказывается на странице нельзя иметь элементы с одинаковым id, пусть даже разных типов
	editTask.call(pp)
}

// ===========================================================================
// При щелчке на кнопке #TaskSubmit в форме создания/редактирования задач
// Отправляет на сервер команду, обновляющую задачу (если её id задан),
// или создающую новую задачу (если id пустой)
// ===========================================================================
function sendTask() {
	$("#spinner").show();
	resetTaskFromAfterUpdate = true;
	if (!$("input#TaskText").val())
	{
		alert('Введите задачу!');
		$("input#TaskText").focus();
		return false;
	}
	$.ajax( {
		url : "/sendTask",
		cache: false,
		type : "post",
		dataType: "text",
		data : $("#TaskForm").serialize(),
		// В случае успешного получения ответа от сервера
		success: function (response) {
			// После того, как задача будет успешно отправлена на сервер
			$("label#TaskResult").html("Готово");
			reloadTasks();
			$("#spinner").hide();
		},
		// В случае возвращение сервером ошибки, или недоступности сервера
		error: function(jqXHR,exception) { 
			showAjaxError(jqXHR,exception);
			$("#spinner").hide();
		}
	} );
	return false;
}

function todayTask() {
	var li_html = '<li id="'+$("input#TaskId").val()+'" class="TodayTask" style="height:20px" title="'+$("input#TaskText").val()+'">'+$("input#TaskText").val()+'</li>';
	if ($("li.TodayTask:last").length>0) { 
		$("li.TodayTask:last").after(li_html) }
	else { $("ul#TodayTasks").prepend(li_html) };
	sendToday();
	return false;
}

function fromdayTask() {
	var li = $('li[id="'+$("input#TaskId").val()+'"]') // поиск с # не срабатывает, оказывается на странице нельзя иметь элементы с одинаковым id, пусть даже разных типов
	console.log(li);
	li.remove();
	sendToday();
	return false;
}

function clearDoneTodayTasks() {
	$("li.done").remove();
	sendToday();
	return false;
}

function clearAllTodayTasks() {
	$("li.TodayTask").remove();
	sendToday();
	return false;
}

function sendToday() {
	$("#spinner").show();
	resetTaskFromAfterUpdate = true;
	var TodayTasksOffset = $("ul#TodayTasks").offset().top;
	var TodayTasks = [];
	$("li.TodayTask").each( function (index, element) {
		var TodayTask = {};
		TodayTask.id = element.id;
		TodayTask.start = Math.round(($(this).offset().top-TodayTasksOffset)/20);
		TodayTask.length = Math.round($(this).outerHeight()/20);
		TodayTasks.push( TodayTask );
	} )
	$.ajax( {
		url : "/arrangeTodayTasks",
		cache: false,
		type : "post",
		dataType: "json",
		data : JSON.stringify( { TodayTasks: TodayTasks, List: $("select#ListSelect").val() } ),
		// В случае успешного получения ответа от сервера
		success: function (response) {
			// После того, как задача будет успешно отправлена на сервер
			$("label#TaskResult").html("Расписание обновлено");
			reloadTasks();
			$("#spinner").hide();
		},
		// В случае возвращение сервером ошибки, или недоступности сервера
		error: function(jqXHR,exception) { 
			showAjaxError(jqXHR,exception);
			$("#spinner").hide();
		}
	} );
}

function reloadTasks() {
	// Обновляем таблицу задач
	$("#TasksTable").load(location.href+" #TasksTable",
		"", function() { 
			// Восстанавливаем функциональность кнопок "+" и кликабельность задач
			afterUpdate();
			// Подготавливаемся ко вводу новой задачи
			if (resetTaskFromAfterUpdate) { $("td.vh button").click(); }
		}
	);

}

function showAjaxError(jqXHR, exception) {
	var msg = '';
	if (jqXHR.status === 0) { msg = 'Not connect.\n Verify Network.'; } 
	else if (jqXHR.status == 404) { msg = 'Requested page not found. [404]'; }
	else if (jqXHR.status == 500) { msg = 'Internal Server Error [500].'+jqXHR.responseText; }
	else if (exception === 'parsererror') { msg = 'Requested JSON parse failed.'; } 
	else if (exception === 'timeout') { msg = 'Time out error.'; } 
	else if (exception === 'abort') { msg = 'Ajax request aborted.'; }
	else { msg = 'Uncaught Error.\n' + jqXHR.responseText; }
	$("label#TaskResult").html(msg);
}

function onListChange() {
	if ($("select#ListSelect").prop('selectedIndex') == 0) {
		window.location.href = "/";
	}
	else {
		window.location.href = "/?list="+$("select#ListSelect").val();	
	}
}

Date.prototype.yyyymmdd = function() {
  var mm = this.getMonth() + 1; // январь=0
  var dd = this.getDate();
  return [this.getFullYear(),
          (mm>9 ? '' : '0') + mm,
          (dd>9 ? '' : '0') + dd
         ].join('-');
};

function createList() {
	var now = new Date();
	var formated_now = now.yyyymmdd();
	var str = $("select#ListSelect").val();
	var username = str.substring(0, str.indexOf(":"));
	$("button#TaskMove").attr('title', username+":"+formated_now);
	$("button#TaskMove").text("Перенести в список "+formated_now);
	if ($("input#TaskId").val() != "") {
		if ( $("button#TaskMove").attr('title') != $("select#ListSelect").val() ) { $("button#TaskMove").show() };
	}
	return false;
}

function moveTask() {
	var old_id = $("input#TaskId").val();
	$("input#TaskId").val( "" );
	$("input#TaskList").val(  $("button#TaskMove").attr('title') );	
	sendTask();

	editTask.call($('p[id="'+old_id+'"]'));
	$("select#TaskStatus").val( "moved" );
	sendTask();
	// Проверяем, если список только создаётся - обновляем текущую страницу
	$('select#ListSelect option[value="'+$("button#TaskMove").attr('title')+'"]').length == 0
	{
		location.reload();
	}
	return false;
}