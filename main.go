package main

import (
	"fmt"
	"net/http"
)

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "<h1>Hello, world2!</h1>")
}

func main() {
	http.HandleFunc("/",hello)
	http.ListenAndServe("0.0.0.0:8080",nil)
}