package handlers

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	"DNSPulse_ConfigHub/pkg/logger"
	"DNSPulse_ConfigHub/pkg/tools"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"
)

func LoginHandler(tmpl *template.Template) http.HandlerFunc {
    return func (w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
        if r.Method == "POST" {
			r.ParseForm()
			username := r.FormValue("username")
			password := r.FormValue("password")

			if !CheckCredentials(username, password) {
				data = map[string]interface{}{
                    "Error": "Invalid username or password",
                }
				logger.Audit.WithFields(logrus.Fields{
					"username": username,
					"ip":       tools.GetClientIP(r),
					"status":   "failed",
				}).Warn("Authentication attempt failed")
			} else {
				tokenString, err := GenerateToken(username)
				if err != nil {
					logger.Audit.WithFields(logrus.Fields{
						"username": username,
						"ip":       tools.GetClientIP(r),
						"status":   "failed",
					}).Error("Error generating session token")
					data = map[string]interface{}{
						"Error": "Internal Server Error",
					}
				} else {
					http.SetCookie(w, &http.Cookie{
						Name:   "session_token",
						Value:  tokenString,
						Path:   "/",
						MaxAge: datastore.GetConfig().WebServer.SesionTimeout,
					})

					logger.Audit.WithFields(logrus.Fields{
						"username": username,
						"token": tools.TruncateToken(tokenString),
						"ip":       tools.GetClientIP(r),
						"status":   "success",
					}).Info("Authentication success")

					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			}
		}

        err := tmpl.ExecuteTemplate(w, "login.html", data)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
}
