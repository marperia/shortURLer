package views

import (
	"net/http"
	"html/template"
	"../setting"
)

func Home(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./templates/index.html")
	data := map[string]string {
		"Name": setting.AppName,
	}
	t.Execute(w, data)
}
