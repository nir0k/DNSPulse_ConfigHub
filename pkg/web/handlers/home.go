package handlers

import (
	"html/template"
	"net/http"
)

func HomeHandler(tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        err := tmpl.ExecuteTemplate(w, "index.html", map[string]interface{}{
            "ShowNavBar": true,
        })
        if err != nil {
            http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
            return
        }
    }
}