package main

import (
	"net/http"
	"os"
	"text/template"
	"golang.org/x/text/language"
	"github.com/Shaked/gomobiledetect"
)

var URI string //URI базы данных MongoDB
var DB string //имя базы данных MongoDB

// Все известные языки системы
var matcher = language.NewMatcher([]language.Tag{
    language.English,   // The first language is used as fallback.
    language.Russian})

// Структура, детально описывающая список задач
// Используется в HTML-шаблоне для вывода основной страницы списка задач
type typeWebFormData struct {
	IsMobile bool
	UserLanguage string
}

// ================================================================================
// Web-страница: Выводит страницу редактирования списка
// ================================================================================
func webFormShow(res http.ResponseWriter, req *http.Request) {

   	// Подготавливаем структуру для инициализации шаблона
	var webFormData typeWebFormData

	// Определяем, с какого устройства пришёл запрос
	detect := mobiledetect.NewMobileDetect(req, nil)
	webFormData.IsMobile = detect.IsMobile() && !detect.IsTablet()

	// Определяем язык, который следует использовать
	lang, _ := req.Cookie("User-Language")
    accept := req.Header.Get("Accept-Language")
    tag, _ := language.MatchStrings(matcher, lang.String(), accept)
    webFormData.UserLanguage = tag.String();

	// Применяем HTML-шаблон
	res.Header().Set("Content-type", "text/html")
	t := template.New("tasks.html")
	t, err := t.ParseFiles("templates/tasks.html")
	if err != nil { panic(err) }
	err = t.Execute(res, webFormData)
	if err != nil { panic(err) }
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
	http.HandleFunc("/",webFormShow)
	
	// Регистрируем файловый сервер для отдачи статичных файлов из папки static
	fs := http.FileServer(http.Dir("./static"))
 	http.Handle("/static/", http.StripPrefix("/static/", fs))

 	// Запускаем web-сервер на всех интерфейсах на порту PORT
	http.ListenAndServe(":"+PORT,nil)
}