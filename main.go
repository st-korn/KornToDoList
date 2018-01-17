package main

import (
	"fmt"
	"net/http"
	"strings"
	"os"
	"encoding/json"
	"text/template"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var URI string
var DB string

type Task struct {
	Id bson.ObjectId `json:"_id" bson:"_id"`
    Text string  `json:"text" bson:"text"`
    Section string  `json:"section" bson:"section"`
    Status string  `json:"status" bson:"status"`
}

type Tasks struct {
	Vs []Task
	Vn []Task
	Ns []Task
	Nn []Task
	Vh []Task
}

func MarkTask(text string) string {
	pos := strings.Index(text, " - ")
	var res string
    if pos > 0 {
    	res = "<span class=\"name\">"+text[:pos]+"</span>"+text[pos:]
    } else { res = text}
    return res
}

func hello(res http.ResponseWriter, req *http.Request) {
	// Соединяемся с базой
	session, err := mgo.Dial(URI)
    if err != nil { panic(err) }
    defer session.Close()
   	c := session.DB(DB).C("Tasks")

	// Выбираем задачи из базы
	var tasks Tasks
	err = c.Find(bson.M{"section": "vs"}).Sort("-timestamp").All(&tasks.Vs)
	if err != nil { panic(err) }
	err = c.Find(bson.M{"section": "vn"}).Sort("-timestamp").All(&tasks.Vn)
	if err != nil { panic(err) }
	err = c.Find(bson.M{"section": "ns"}).Sort("-timestamp").All(&tasks.Ns)
	if err != nil { panic(err) }
	err = c.Find(bson.M{"section": "nn"}).Sort("-timestamp").All(&tasks.Nn)
	if err != nil { panic(err) }
	err = c.Find(bson.M{"section": "vh"}).Sort("-timestamp").All(&tasks.Vh)
	if err != nil { panic(err) }

	funcMap := template.FuncMap{"marktask": MarkTask}

	// Применяем HTML-шаблон
	res.Header().Set("Content-type", "text/html")
	t := template.New("tasks.html")
	t = t.Funcs(funcMap)
	t, err = t.ParseFiles("tasks.html")
	if err != nil { panic(err) }
	err = t.Execute(res, tasks)
	if err != nil { panic(err) }
}

func sendTask(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	fmt.Println(req.Form)

	// Соединяемся с базой
	session, err := mgo.Dial(URI)
    if err != nil { panic(err) }
    defer session.Close()
   	c := session.DB(DB).C("Tasks")

    // Заполняем поля задчи
	task := &Task {
		Text : strings.Join(req.Form["text"],""),
		Section : strings.Join(req.Form["section"],""),
		Status : strings.Join(req.Form["status"],"") }

	// Проверяем на корректность задачи
	task.Text = strings.TrimSpace(task.Text)
	if task.Text == "" { 
		http.Error(res, "Задача не может быть пустой!", http.StatusInternalServerError)
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
			"text": task.Text,
			"section": task.Section,
			"status": task.Status } }
		err = c.Update(condition, change)
		if err != nil { panic(err) }
	}

	// Возвращаем полную информацию о задаче, включая ID, в формате JSON
	js, err := json.Marshal(task)
  	if err != nil { http.Error(res, err.Error(), http.StatusInternalServerError)
  		return  }

  	res.Header().Set("Content-Type", "application/json; charset=utf-8")
  	res.Header().Set("Access-Control-Allow-Origin", "*")
  	res.Write(js)
  	fmt.Println(string(js))
}

func main() {

	URI = os.Getenv("MONGODB_ADDON_URI")
	DB = os.Getenv("MONGODB_ADDON_DB")

	http.HandleFunc("/",hello)
	http.HandleFunc("/sendTask",sendTask)

	// Регистрируем файловый сервер для отдачи статичных файлов из папки static
	fs := http.FileServer(http.Dir("./static"))
 	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(":8080",nil)
}