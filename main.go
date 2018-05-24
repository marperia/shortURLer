package main

import (
	"net/http"
	"log"
	"./setting"
	"./controllers" // import controllers such as Save and Get
	"./views" // import views such as home
	//	well now we have not to import models
)

func main() {
	http.HandleFunc("/", views.Home)
	http.HandleFunc("/save/", controllers.SaveURL)
	http.HandleFunc("/get/", controllers.GetURL)

	log.Println("Web-server is running: http://127.0.0.1" + setting.Port)
	err := http.ListenAndServe(setting.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
