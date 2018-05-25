package main

import (
	"net/http"
	"log"
	"./setting"
	"./controllers"
	"./views"
)

func main() {
	http.HandleFunc("/", views.Home)
	http.HandleFunc("/save/", controllers.SaveURL)
	http.HandleFunc("/get/", controllers.GetURL)

	log.Println("Web-server is running: http://" + setting.Host + setting.Port)
	err := http.ListenAndServe(setting.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
