package handlers2

import (
	"DNSPulse_ConfigHub/pkg/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

type GeneralConfig struct {
    Path string `json:"path"`
}

type LogConfig struct {
    Path        string `json:"path"`
    MinSeverity string `json:"minSeverity"`
    MaxAge      int    `json:"maxAge"`
    MaxSize     int    `json:"maxSize"`
    MaxFiles    int    `json:"maxFiles"`
}

type WebConfig struct {
    Port           int    `json:"port"`
    ListenAddress  string `json:"listenAddress"`
    SSLEnabled     bool   `json:"sslIsEnable"`
    SSLCertPath    string `json:"sslCertPath"`
    SSLKeyPath     string `json:"sslKeyPath"`
    SessionTimeout int    `json:"sesionTimeout"`
    Username       string `json:"username"`
    Password       string `json:"password"`
}

type APIConfig struct {
    Port         int    `json:"port"`
    SSLEnabled   bool   `json:"sslEnabled"`
    SSLCertPath  string `json:"sslCertPath"`
    SSLKeyPath   string `json:"sslKeyPath"`
    Username     string `json:"username"`
    Password     string `json:"password"`
    JWTKey       string `json:"jwtKey"`
}

type SegmentConfig struct {
    Name string `json:"name"`
    Path string `json:"path"`
}

type ConfigData struct {
    General  GeneralConfig   `json:"general"`
    Log      LogConfig       `json:"log"`
    Audit    LogConfig       `json:"audit"`
    Web      WebConfig       `json:"web"`
    API      APIConfig       `json:"api"`
    Segments []SegmentConfig `json:"segments"`
}

type TemplateData struct {
    Title string
    Config ConfigData
	ShowNavBar bool
}


func ConfigGeneralHandler(tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
        if err != nil {
            logger.Logger.Errorf("Failed to retrieve session token: %v", err)
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
		authHeaderValue := "Bearer " + token.Value

        configData := ConfigData{}
        baseURL := "http://localhost:8080/api/config/general"

        // Simplify error handling with a loop
        endpoints := map[string]interface{}{
            "/main":    &configData.General,
            "/log":     &configData.Log,
            "/audit":   &configData.Audit,
            "/web":     &configData.Web,
            "/api":     &configData.API,
            "/segment": &configData.Segments,
        }

        for endpoint, config := range endpoints {
            if err := fetchAndUnmarshal(baseURL+endpoint, authHeaderValue, config); err != nil {
                errMsg := fmt.Sprintf("Failed to fetch or unmarshal config for %s: %v", endpoint, err)
                fmt.Println(errMsg) // Log the error
                http.Error(w, errMsg, http.StatusInternalServerError)
                return
            }
        }

		templateData := TemplateData{
			Title: "Configuration Overview",
			Config: configData,
			ShowNavBar: true,
		}
		
		err = tmpl.ExecuteTemplate(w, "config-general.html", templateData)		
		if err != nil {
			http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
		}
    }
}

// fetchAndUnmarshal fetches data from the given URL and unmarshals it into the target.
func fetchAndUnmarshal(url, token string, target interface{}) error {
    client := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return err
    }

    req.Header.Add("Authorization", token)

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return err
    }

    return json.Unmarshal(body, target)
}

// UpdateGeneralConfig handles PATCH requests for the general configuration by forwarding them to an external API.
func UpdateGeneralConfig(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPatch {
        http.Error(w, "Unsupported request method.", http.StatusMethodNotAllowed)
        return
    }

    // Read the body of the incoming request
    reqBody, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body", http.StatusBadRequest)
        return
    }

    // Prepare the request to the external API
    externalAPIURL := "http://external-api.com/config/general/main" // Replace with the actual URL of the external API
    req, err := http.NewRequest(http.MethodPatch, externalAPIURL, bytes.NewBuffer(reqBody))
    if err != nil {
        http.Error(w, "Error creating request to external API", http.StatusInternalServerError)
        return
    }

    // Copy the content type and authorization header from the original request
    req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
    req.Header.Set("Authorization", r.Header.Get("Authorization"))

    // Send the request to the external API
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        http.Error(w, "Error sending request to external API", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    // Check the response from the external API
    if resp.StatusCode != http.StatusOK {
        // If the external API did not return a 200 OK, relay the error back to the client
        http.Error(w, "External API responded with an error", resp.StatusCode)
        return
    }

    // Optionally, read and forward the response body from the external API if needed
    // responseBody, _ := ioutil.ReadAll(resp.Body)
    // w.Write(responseBody)

    // If everything went well, send a success response back to the client
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Configuration updated successfully via external API"))
}
