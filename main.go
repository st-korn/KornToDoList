package main

import (
	"net/http" // для создания http-сервера
	"os" // для получения переменных среды
	"text/template" // для использования шаблона html-страницы
	"encoding/json" // для чтения
	"golang.org/x/text/language" // для определения предпочтительного языка пользователя
	"golang.org/x/text/language/display" // для вывода национальных наименования различных языков
	"github.com/Shaked/gomobiledetect" // для определения мобильных браузеров
)

var URI string //URI базы данных MongoDB
var DB string //имя базы данных MongoDB

// Все известные языки системы
var SupportedLangs = []language.Tag{
    language.English,   // The first language is used as fallback.
    language.German,
    language.Russian}

// Структура, описывающая язык: английское_название_языка и национальное_название_языка
type typeLang struct {
	EnglishName string
	NationalName string
}

// Структура, детально описывающая список задач
// Используется в HTML-шаблоне для вывода основной страницы списка задач
type typeWebFormData struct {
	IsMobile bool // признак того, что страница открыта из мобильного браузера
	UserLang string // английское_названия_языка, на котором следует показывать страницу
	Langs []typeLang // общий перечень языков, на которые переведена система
	Labels [string]string // текстовки текущего выбранного языка
}

// ================================================================================
// Web-страница: Выводит основную страницу web-приложения
// ================================================================================
func webFormShow(res http.ResponseWriter, req *http.Request) {

   	// Подготавливаем структуру для инициализации шаблона
	var webFormData typeWebFormData

	// Определяем, с какого устройства пришёл запрос
	detect := mobiledetect.NewMobileDetect(req, nil)
	webFormData.IsMobile = detect.IsMobile() && !detect.IsTablet()

    // Загружаем полный список поддерживаемых языков
    webFormData.Langs = make([]typeLang,len(SupportedLangs))
	for i, tag := range SupportedLangs {
		webFormData.Langs[i].EnglishName = display.English.Tags().Name(tag)
		webFormData.Langs[i].NationalName = display.Self.Name(tag)
	}

	// Определяем язык, который следует использовать
	lang, _ := req.Cookie("User-Language")
    accept := req.Header.Get("Accept-Language")
    matcher := language.NewMatcher(SupportedLangs)
    tag, _ := language.MatchStrings(matcher, lang.String(), accept)
    webFormData.UserLang = display.English.Tags().Name(tag);

	// Загружаем текстовки выбранного языка
	webFormData.

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