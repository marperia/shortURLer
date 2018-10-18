package views

import (
	"net/http"
	"html/template"
	"github.com/marperia/shortURLer/setting"
)

func Home(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/index.html")
	if err != nil {

	}
	data := map[string]string {
		"Name": setting.AppName,
	}
	t.Execute(w, data)
}
