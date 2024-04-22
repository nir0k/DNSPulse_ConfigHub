package handlers2

import (

	"DNSPulse_ConfigHub/pkg/logger"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"net/http"
)


func LoginHandler(tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var data map[string]interface{}
        if r.Method == "POST" {
            err := r.ParseForm()
            if err != nil {
                http.Error(w, "Error parsing form", http.StatusBadRequest)
                return
            }
            username := r.FormValue("username")
            password := r.FormValue("password")

            // Encode credentials
            credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
            // Create a new request to the external login API
            req, err := http.NewRequest("POST", "http://localhost:8080/auth/login", nil)
            if err != nil {
                logger.Logger.Errorf("Error creating request to external login API: %v", err)
                http.Error(w, "Error contacting authentication service", http.StatusInternalServerError)
                return
            }

            // Set the Authorization header for Basic Auth
            req.Header.Add("Authorization", "Basic "+credentials)

            // Perform the request
            client := &http.Client{}
            resp, err := client.Do(req)
            if err != nil {
                logger.Logger.Errorf("Error sending request to external login API: %v", err)
                http.Error(w, "Error contacting authentication service", http.StatusInternalServerError)
                return
            }
            defer resp.Body.Close()

            // Check the response
            if resp.StatusCode != http.StatusOK {
                data = map[string]interface{}{
                    "Error": "Invalid username or password",
                }
                tmpl.ExecuteTemplate(w, "login.html", data)
                return
            }

            // Decode the token from the response body
            var respData map[string]string
            err = json.NewDecoder(resp.Body).Decode(&respData)
            if err != nil {
                logger.Logger.Errorf("Error decoding response from external login API: %v", err)
                http.Error(w, "Error decoding response", http.StatusInternalServerError)
                return
            }
            tokenString, ok := respData["token"]
            if !ok {
                logger.Logger.Errorf("Token not found in response")
                http.Error(w, "Authentication failed", http.StatusInternalServerError)
                return
            }

            // Set the token in a cookie
            http.SetCookie(w, &http.Cookie{
                Name:   "session_token",
                Value:  tokenString,
                Path:   "/",
                MaxAge: 3600, // Adjust according to your session timeout
            })

            http.Redirect(w, r, "/", http.StatusFound)
        } else {
            tmpl.ExecuteTemplate(w, "login.html", data)
        }
    }
}