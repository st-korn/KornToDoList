var resetTaskFormAfterUpdate = false;
var oldBackgroundTop = 0;
$( init );

// ===========================================================================
// Инициализация страницы
// ===========================================================================
function init() {
	$("select#ListSelect").change(onListChange);
	$("button#ListCreate").bind('click', createList);
	$("button#TaskSubmit").bind('click', sendTask);
	$("button#TaskMove").bind('click', moveTask);
	$("label#ListHelp").bind('click', startTour);
	afterUpdate();
	$("td.vh button").click();
	$("#spinner").hide();
	$(".tourplace").hide();
	$(".tourplace-background").hide();
}


function startTour() {
	// Скрываем таблицу
	$(".workplace").css("position", "fixed");
	$(".workplace").width($(".TasksTable").width())
	$(".tourplace").show();
	$(".tourplace-background").show();
	oldBackgroundTop = $(".tourplace-background").offset().top;

	$("div#tour5-cat").offset({ top: $("section#tour5").offset().top+$("select#TaskSection").offset().top-7, left: $("select#TaskSection").offset().left-7 });
	$("div#tour5-cat").height( $("select#TaskSection").height() );
	$("div#tour5-cat").width( $("select#TaskSection").width() );
	$("div#tour5-vh").offset({ top: $("section#tour5").offset().top+$("td.vh .ColumnHeader").offset().top+$("td.vh .ColumnHeader").height()+5, left: $("td.vh").offset().left });

	/*// init controller
	var controller = new ScrollMagic.Controller();

	// build scenes
	new ScrollMagic.Scene({triggerElement: "#tour1", duration: window.innerHeight})
					.setClassToggle("#high1", "active") // add class toggle
					.addIndicators() // add indicators (requires plugin)
					.addTo(controller);
	new ScrollMagic.Scene({triggerElement: "#tour2"})
					.setClassToggle("#high2", "active") // add class toggle
					.addIndicators() // add indicators (requires plugin)
					.addTo(controller);*/

	/*		$('.tour').fullpage({
				anchors: ['firstPage', 'secondPage', '3rdPage'],
				sectionsColor: ['#C63D0F', '#1BBC9B', '#7E8F7C'],
				scrollBar: true});*/

	$("div.tour-background").each( function (index, element) {
		var backgroundID = element.id;
		$(this).height( $("section#tour"+backgroundID[backgroundID.length-1]).height()*1.25 );
	} );

	$.scrollify({
            section : ".tour",
            updateHash: false,
    		touchScroll: true
          });

$(window).bind('scroll',function(e){
    parallaxScroll();
});
 

}

function parallaxScroll(){
    var scrolled = $(window).scrollTop();
    //console.log("s=",scrolled);
	//$(".tourplace-background").offset({ top: scrolled*1.25 });
	//console.log("o=", oldBackgroundTop);
	 $('.tourplace-background').css('top',(0-(scrolled*1.25))+'px');
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
	resetTaskFormAfterUpdate = false;
	$("input#TaskId").val( "" );
	$("input#TaskList").val(  $("select#ListSelect").val() );
	$("input#TaskText").val( "" );
	$("select#TaskSection").val( $(this).closest('td').prop('class') );
	$("select#TaskStatus").val( "created" );
	$("select#TaskIcon").val( "" );
	$("button#TaskSubmit").text("Добавить");
	$("button#TaskMove").hide();
	$("input#CheckToday").prop('checked',false);
	$("label#TaskResult").text("");
	$("input#TaskText").focus();
}

// ===========================================================================
// При щелчке на задаче - загружает её для редактирования в форму #TaskForm
// ===========================================================================
function editTask() {
	resetTaskFormAfterUpdate = false;
	$("input#TaskId").val( $(this).prop('id') );
	$("input#TaskList").val(  $("select#ListSelect").val() );
	$("input#TaskText").val( $(this).text() );
	$("select#TaskSection").val( $(this).closest('td').prop('class') );
	$("select#TaskStatus").val( $(this).prop('class') );
	$("select#TaskIcon").val( $(this).find("img").prop('class') );
	$("button#TaskSubmit").text("Изменить");
	if ( ($("button#TaskMove").prop('title') != $("select#ListSelect").val()) && 
		($("select#TaskStatus").val() == "created") ) 
		{ $("button#TaskMove").show() }
	else
		{ $("button#TaskMove").hide() };
	if( ($("select#TaskStatus").val() == "created") &&
		($('li[id="'+$(this).prop('id')+'"]').length == 0) ) {
		$("button#TaskToday").show() }
	else {
		$("button#TaskToday").hide() };
	if ($('li[id="'+$(this).prop('id')+'"]').length > 0) {
		$("input#CheckToday").prop('checked',true) }
	else {
		$("input#CheckToday").prop('checked',false) };
	$("label#TaskResult").text("");
	$("input#TaskText").focus();
}

// ===========================================================================
// При щелчке на блоке запланированной на сегодня задачи - срабатывает
// так же, как если бы щёлкнули на задаче в одной из колонок таблицы
// ===========================================================================
function editTodayTask() {
	var id = $(this).prop('id')
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
	resetTaskFormAfterUpdate = true;
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
		},
		// В случае возвращение сервером ошибки, или недоступности сервера
		error: function(jqXHR,exception) { 
			showAjaxError(jqXHR,exception);
			$("#spinner").hide();
		}
	} );
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
	resetTaskFormAfterUpdate = true;
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
			if (resetTaskFormAfterUpdate) { $("td.vh button").click(); }
			// Выключаем спиннер
			$("#spinner").hide();
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
	$("button#TaskMove").prop('title', username+":"+formated_now);
	$("button#TaskMove").text("Перенести в список "+formated_now);
	if ($("input#TaskId").val() != "") {
		if ( $("button#TaskMove").prop('title') != $("select#ListSelect").val() ) { $("button#TaskMove").show() };
	}
	return false;
}

function moveTask() {
	$("#spinner").show();
	resetTaskFormAfterUpdate = true;
	var old_id = $("input#TaskId").val();

	$("input#TaskId").val( "" );
	$("input#TaskList").val(  $("button#TaskMove").prop('title') );	
	$.ajax( {
		url : "/sendTask",
		cache: false,
		type : "post",
		dataType: "text",
		data : $("#TaskForm").serialize(),
		// В случае успешного получения ответа от сервера
		success: function (response) {
			editTask.call($('p[id="'+old_id+'"]'));
			$("select#TaskStatus").val( "moved" );
			resetTaskFormAfterUpdate = true;
			$.ajax( {
				url : "/sendTask",
				cache: false,
				type : "post",
				dataType: "text",
				data : $("#TaskForm").serialize(),
				// В случае успешного получения ответа от сервера
				success: function (response) {
					// Проверяем, если список только создаётся - обновляем текущую страницу
					if ($('select#ListSelect option[value="'+$("button#TaskMove").prop('title')+'"]').length == 0)
					{
						location.reload();
					}
					else 
					{
						$("label#TaskResult").html("Задача перенесена");
						reloadTasks();
					}
				},
				// В случае возвращение сервером ошибки, или недоступности сервера
				error: function(jqXHR,exception) { 
					showAjaxError(jqXHR,exception);
					$("#spinner").hide();
				}
			} );
		},
		// В случае возвращение сервером ошибки, или недоступности сервера
		error: function(jqXHR,exception) { 
			showAjaxError(jqXHR,exception);
			$("#spinner").hide();
		}
	} );
	return false;
}