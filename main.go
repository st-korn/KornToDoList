package main

import (
	"fmt"
	"net/http"
	"strings"
	"os"
	"encoding/json"
	"text/template"
	"sort"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/Shaked/gomobiledetect"
)

var URI string //URI базы данных MongoDB
var DB string //имя базы данных MongoDB

// Настройки расписания времени
const (
	StartHour = 7
	EndHour = 24
)

// Структура записи о задаче
// используется в формате JSON для ajax-обмена с клиентскими страницами
// используется в формате BSON как структура полей коллекции в базе MongoDB
type Task struct {
	Id bson.ObjectId `json:"_id" bson:"_id"`
    Text string  `json:"text" bson:"text"`
    Section string  `json:"section" bson:"section"`
    Status string  `json:"status" bson:"status"`
    List string  `json:"list" bson:"list"`
    Icon string `json:"icon" bson:"icon"`
    Length int `json:"length" bson:"length"`
}

// Структура записи о задаче, которую планируется выполнить сегодня
// Использщуется в HTML-шаблоне для вывода шкалы задач на сегодня
type TodayTask struct {
	Id bson.ObjectId  `json:"_id" bson:"_id"`
	IsRelax bool
	Start int `json:"start" bson:"start"` //номер 15-минутного интервала, в который начинается задача
	Length int `json:"length" bson:"length"` //количество 15-минутных интервалов
	Height int //количество пикселей для отображения задачи
	Text string  `json:"text" bson:"text"` // Текст задачи
	Status string `json:"status" bson:"status"` // Статус задачи
	Icon string `json:"icon" bson:"icon"` // Пиктограмма задачи
}

type TimeLabel struct {
	Text string
	IsRelax bool
}

// Структура, детально описывающая список задач
// Используется в HTML-шаблоне для вывода основной страницы списка задач
type Tasks struct {
	IsMobile bool

	MyLists []string
	CurList string
	ActualListFullName string
	ActualListShortName string

	TimeLabels []TimeLabel
	TodayTasks []TodayTask
	Vs []Task
	Vn []Task
	Ns []Task
	Nn []Task
	Vh []Task
}

// ================================================================================
// Вспомогательная функция: выделяет в тексте задачи text имена исполнителей
// (все символы до первого вхождения " - ")
// и добавляет к ним html-обёртку с тэгами выделения имён исполнителей
// ================================================================================
func MarkTask(text string) string {
	pos := strings.Index(text, " - ")
	var res string
    if pos > 0 {
    	res = "<span class=\"name\">"+text[:pos]+"</span>"+text[pos:]
    } else { res = text}
    return res
}

// ================================================================================
// Web-страница: Выводит страницу редактирования списка
// ================================================================================
func hello(res http.ResponseWriter, req *http.Request) {
	// Соединяемся с базой
	session, err := mgo.Dial(URI)
    if err != nil { panic(err) }
    defer session.Close()
   	c := session.DB(DB).C("Tasks")


   	// Подготавливаем структуру для инициализации шаблона
	var tasks Tasks

	// Определяем, с какого устройства пришёл запрос
	detect := mobiledetect.NewMobileDetect(req, nil)
	tasks.IsMobile = detect.IsMobile()

	// Составляем перечень списов, доступных текущему пользователю
	err = c.Find(nil).Distinct("list",&tasks.MyLists)
	if err != nil { panic(err) }
	sort.Sort(sort.Reverse(sort.StringSlice(tasks.MyLists)))

	// Определяем текущий запрошенный список
	tasks.CurList = req.URL.Query().Get("list")
	// Проверяем его корректность:
	// проверяем, что он не пустой
	if len(tasks.CurList) == 0 { tasks.CurList=tasks.MyLists[0] }
	// проверяем, что такой список есть в базе
	listfound := false;
	for _, listname := range tasks.MyLists {
		if listname == tasks.CurList { listfound = true}
	}
	// если запросили отсутствующий список
	if listfound == false { 
		http.Redirect(res,req,"/",302)
		return
	}
	// Определяем актуальный список
	tasks.ActualListFullName = tasks.MyLists[0]
	slices := strings.SplitAfter(tasks.ActualListFullName,":")
	tasks.ActualListShortName = slices[len(slices)-1]
	
	// Заполняем временную шкалу
	var timelabel TimeLabel
	for hour:=StartHour; hour<EndHour; hour++ {
		for min:=0; min<60; min+=15 {
			timelabel.Text = fmt.Sprintf("%02d:%02d",hour,min)
			tasks.TimeLabels = append(tasks.TimeLabels, timelabel )
		}
	}

	// Выбираем сегодняшние задачи
	var todaytasks []TodayTask
	var relaxtask TodayTask
	err = c.Find(bson.M{"length": bson.M{"$gt": 0}, "list": tasks.CurList}).Sort("start").All(&todaytasks)
	if err != nil { panic(err) }
	// Заполняем пустые и недостающие интервалы промежутками отдыха
	relaxtask.IsRelax = true
	relaxtask.Length = 1
	relaxtask.Height = 20
	curInterval := 0
	for _, todaytask := range todaytasks {
		// Добавляем необходимое количество промежутков отдыха перед данной задачей
		for i:=curInterval; i<todaytask.Start; i++ {
			relaxtask.Start = i
			tasks.TodayTasks = append(tasks.TodayTasks, relaxtask)
			tasks.TimeLabels[i].IsRelax = true
		}
		// Добавляем текущую задачу, заполняя недостающие поля в информации о задаче
		todaytask.IsRelax = false
		todaytask.Height = todaytask.Length*20
		if curInterval > todaytask.Start { todaytask.Start = curInterval }
		tasks.TodayTasks = append(tasks.TodayTasks, todaytask)
		tasks.TimeLabels[todaytask.Start].IsRelax = false
		// Переходим на следующий интервал
		curInterval = todaytask.Start + todaytask.Length
	}
	// Добавляем необходимое количество промежутков отдыха после последней задачи
	for i:=curInterval; i<(EndHour-StartHour)*4; i++ {
		relaxtask.Start = i
		tasks.TodayTasks = append(tasks.TodayTasks, relaxtask)
		tasks.TimeLabels[i].IsRelax = true
	}

	// Выбираем задачи из базы
	err = c.Find(bson.M{"section": "vs", "list": tasks.CurList}).Sort("-timestamp").All(&tasks.Vs)
	if err != nil { panic(err) }
	err = c.Find(bson.M{"section": "vn", "list": tasks.CurList}).Sort("-timestamp").All(&tasks.Vn)
	if err != nil { panic(err) }
	err = c.Find(bson.M{"section": "ns", "list": tasks.CurList}).Sort("-timestamp").All(&tasks.Ns)
	if err != nil { panic(err) }
	err = c.Find(bson.M{"section": "nn", "list": tasks.CurList}).Sort("-timestamp").All(&tasks.Nn)
	if err != nil { panic(err) }
	err = c.Find(bson.M{"section": "vh", "list": tasks.CurList}).Sort("-timestamp").All(&tasks.Vh)
	if err != nil { panic(err) }

	funcMap := template.FuncMap{"marktask": MarkTask}

	// Применяем HTML-шаблон
	res.Header().Set("Content-type", "text/html")
	t := template.New("tasks.html")
	t = t.Funcs(funcMap)
	t, err = t.ParseFiles("templates/tasks.html")
	if err != nil { panic(err) }
	err = t.Execute(res, tasks)
	if err != nil { panic(err) }
}

// ================================================================================
// Web-API: сохраняет задачу в базе данных
// ================================================================================
func sendTask(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	//fmt.Println(req.Form)

	// Соединяемся с базой
	session, err := mgo.Dial(URI)
    if err != nil { panic(err) }
    defer session.Close()
   	c := session.DB(DB).C("Tasks")

    // Заполняем поля задчи
	task := &Task {
		List : strings.Join(req.Form["list"],""),
		Text : strings.Join(req.Form["text"],""),
		Section : strings.Join(req.Form["section"],""),
		Status : strings.Join(req.Form["status"],""),
		Icon : strings.Join(req.Form["icon"],"") }
	var checkToday = strings.Join(req.Form["today"],"")

	// Проверяем на корректность задачи
	task.Text = strings.TrimSpace(task.Text)
	if task.Text == "" { 
		http.Error(res, "Задача не может быть пустой!", http.StatusInternalServerError)
		return
	}
	if task.List == "" { 
		http.Error(res, "Задача должна принадлежать списку!", http.StatusInternalServerError)
		return
	}

	if len(req.Form["id"])==0 || req.Form["id"][0] == ""  {
		// Добавляем задачу
		task.Id = bson.NewObjectId() // заменяем пустой id на новый
		err = c.Insert(task)
		if err != nil { panic(err) }
	} else {
		// Исправляем задачу
		task.Id = bson.ObjectIdHex(strings.Join(req.Form["id"],""))
		condition := bson.M{"_id": task.Id}
		change := bson.M { "$set": bson.M {
			"list": task.List,
			"text": task.Text,
			"section": task.Section,
			"status": task.Status,
			"icon": task.Icon } }
		err = c.Update(condition, change)
		if err != nil { panic(err) }
	}

	// Проверяем, требуется ли задачу поместить в сегодняшнее расписание / удалить их него
	err = c.Find(bson.M{"_id": task.Id}).One(&task)
	if err != nil { panic(err) }
	if !(task.Length > 0) && (checkToday == "1") {
		// Выбираем сегодняшние задачи
		var todaytasks []TodayTask
		err = c.Find(bson.M{"length": bson.M{"$gt": 0}, "list": task.List}).Sort("start").All(&todaytasks)
		if err != nil { panic(err) }
		// Ищем среди сегодняшних задач первый пустой интервал
		curInterval := 0
		for _, todaytask := range todaytasks {
			if curInterval<todaytask.Start { break }
			curInterval = todaytask.Start + todaytask.Length
		}
		// Добавляем задачу в сегодняшнее расписание
		condition := bson.M{"_id": task.Id}
		change := bson.M { "$set": bson.M {"length": 1, "start": curInterval} }
		err = c.Update(condition, change)
		if err != nil { panic(err) }

	} else if (task.Length > 0) && !(checkToday == "1") {
		// Удаляем задачу из сегодняшнего расписания
		condition := bson.M{"_id": task.Id}
		change := bson.M { "$set": bson.M {"length": 0} }
		err = c.Update(condition, change)
		if err != nil { panic(err) }
	}

	// Возвращаем полную информацию о задаче, включая ID, в формате JSON
	js, err := json.Marshal(task)
  	if err != nil { http.Error(res, err.Error(), http.StatusInternalServerError)
  		return  }

  	res.Header().Set("Content-Type", "application/json; charset=utf-8")
  	//res.Header().Set("Access-Control-Allow-Origin", "*")
  	res.Write(js)
  	//fmt.Println(string(js))
}

// ================================================================================
// Web-API: сохраняет информацию о расписании сегодняшнего дня, обновляя задачи в базе данных
// ================================================================================
func arrangeTodayTasks(res http.ResponseWriter, req *http.Request) {
	// Определяем структуры данных для входящих аргументов API
	type jsonTodayTask struct {
		Id string
		Start int
		Length int
	}
	type jsonTodayTasks struct {
		TodayTasks []jsonTodayTask
		List string
	}
	
	// Разбираем поступивший JSON
	decoder := json.NewDecoder(req.Body)
    var tt jsonTodayTasks
    err := decoder.Decode(&tt)
    if err != nil { panic(err) }
    defer req.Body.Close()

	// Соединяемся с базой
	session, err := mgo.Dial(URI)
    if err != nil { panic(err) }
    defer session.Close()
   	c := session.DB(DB).C("Tasks")

   	// Формируем перечень задач текущего списка, у которых есть длительность 
   	// (которые должны выполниться сегодня)
   	var OldTodayTasks []Task
   	err = c.Find(bson.M{"length": bson.M{"$gt": 0}, "list": tt.List}).All(&OldTodayTasks)
	if err != nil { panic(err) }

    // Перебираем задачи и внсим изменения а базу данных
    for _, t := range tt.TodayTasks {
		// Исправляем задачу
		condition := bson.M{"_id": bson.ObjectIdHex(t.Id)}
		change := bson.M { "$set": bson.M {
			"start": t.Start,
			"length": t.Length} }
		err = c.Update(condition, change)
		if err != nil { panic(err) }
		// Удаляем обновлённую задачу из списка всех сегодняшних задач
		for i := range OldTodayTasks {
			if OldTodayTasks[i].Id.Hex() == t.Id {
				OldTodayTasks = append(OldTodayTasks[:i], OldTodayTasks[i+1:]...)
				break
			}
		}
    }

    // Задачи, которые имели длительность, но не попали в параметры API-запроса - 
    // подлежат исключению из текщуих дел
	for _, t := range OldTodayTasks {
		// обнуляем у таких задач длительность
		condition := bson.M{"_id": t.Id}
		change := bson.M { "$set": bson.M {"length": 0} }
		err = c.Update(condition, change)
		if err != nil { panic(err) }
	}

    // Возвращаем тот же JSON
	js, err := json.Marshal(tt)
  	if err != nil { http.Error(res, err.Error(), http.StatusInternalServerError)
  		return  }

  	res.Header().Set("Content-Type", "application/json; charset=utf-8")
  	//res.Header().Set("Access-Control-Allow-Origin", "*")
  	res.Write(js)
  	//fmt.Println(string(js))
}


// ================================================================================
// Основная программа: запуск web-сервера
// ================================================================================
func main() {

	// Считываем переменные окружения
	URI = os.Getenv("MONGODB_ADDON_URI")
	DB = os.Getenv("MONGODB_ADDON_DB")
	PORT := os.Getenv("PORT")

	// Назначаем обработчики для пользовательских web-запросов
	http.HandleFunc("/",hello)
	
	// Назначаем обработчики для web-запросов API
	http.HandleFunc("/sendTask",sendTask)
	http.HandleFunc("/arrangeTodayTasks",arrangeTodayTasks)

	// Регистрируем файловый сервер для отдачи статичных файлов из папки static
	fs := http.FileServer(http.Dir("./static"))
 	http.Handle("/static/", http.StripPrefix("/static/", fs))

 	// Запускаем web-сервер на всех интерфейсах на порту PORT
	http.ListenAndServe(":"+PORT,nil)
}