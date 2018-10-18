package main

import (
	"net/http"
	"log"
	"github.com/marperia/shortURLer/setting"
	"github.com/marperia/shortURLer/controllers"
	"github.com/marperia/shortURLer/views"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	http.HandleFunc("/", views.Home)
	http.HandleFunc("/save/", controllers.SaveURL)
	http.HandleFunc("/get/", controllers.GetURL)

	log.Println("Web-server is running: http://" + setting.Host + setting.Port)
	err := http.ListenAndServe(setting.Port, http.DefaultServeMux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
