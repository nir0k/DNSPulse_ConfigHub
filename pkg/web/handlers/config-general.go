package handlers

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
)

func toJSON(v interface{}) template.JS {
    b, err := json.Marshal(v)
    if err != nil {
        // Handle error (e.g., by logging or defaulting to a safe value)
        return template.JS("{}")
    }
    return template.JS(b)
}

func createTemplateFuncMap() template.FuncMap {
    return template.FuncMap{
        "toJSON": toJSON,
    }
}

func ConfigGeneralHandler(tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        configData := datastore.GetConfig()

        funcMap := createTemplateFuncMap()
        tmpl = tmpl.Funcs(funcMap)

        // templateData := map[string]interface{}{
        //     "ShowNavBar": true,      // This is existing data you were passing
        //     "Config":     configData, // Pass the entire configuration structure
        // }

        templateData := map[string]interface{}{
            "ShowNavBar": true,
            "Config":     configData,
        }

        err := tmpl.ExecuteTemplate(w, "config-general.html", templateData)
        if err != nil {
            http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
            return
        }
    }
}

func UpdateGeneralConfigHandler(w http.ResponseWriter, r *http.Request) {
    var updatedConfig datastore.UpdateConfigRequest
    err := json.NewDecoder(r.Body).Decode(&updatedConfig)
    if err != nil {
        http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

	newConf := datastore.GetConfig()
	newConf.General = updatedConfig.General
    if err := datastore.UpdateConfig(*newConf); err != nil {
        http.Error(w, "Failed to update general configuration: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
    })
}

func UpdateLogConfigHandler(w http.ResponseWriter, r *http.Request) {
    type EditRequest struct {
        Data datastore.LogAppConfigStruct `json:"data"`
    }
    var updatedConfig EditRequest
    err := json.NewDecoder(r.Body).Decode(&updatedConfig)
    if err != nil {
        http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

	newConf := datastore.GetConfig()
	newConf.Log = updatedConfig.Data
    if err := datastore.UpdateConfig(*newConf); err != nil {
        http.Error(w, "Failed to update log configuration: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
    })
}

func UpdateAuditConfigHandler(w http.ResponseWriter, r *http.Request) {
    type EditRequest struct {
        Data datastore.LogAuditConfigStruct `json:"data"`
    }
    var updatedConfig EditRequest
    err := json.NewDecoder(r.Body).Decode(&updatedConfig)
    if err != nil {
        http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

	newConf := datastore.GetConfig()
	newConf.Audit = updatedConfig.Data
    if err := datastore.UpdateConfig(*newConf); err != nil {
        http.Error(w, "Failed to update audit configuration: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
    })
}

func UpdateWebServerConfigHandler(w http.ResponseWriter, r *http.Request) {
    type EditRequest struct {
        Data datastore.WebServerConfigStruct `json:"data"`
    }
    var updatedConfig EditRequest
    err := json.NewDecoder(r.Body).Decode(&updatedConfig)
    if err != nil {
        http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

	newConf := datastore.GetConfig()
    if updatedConfig.Data.Password == "" {
        updatedConfig.Data.Password = newConf.WebServer.Password
    }
	newConf.WebServer = updatedConfig.Data
    if err := datastore.UpdateConfig(*newConf); err != nil {
        http.Error(w, "Failed to update web-server configuration: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
    })
}

func DownloadConfigHandler(w http.ResponseWriter, r *http.Request) {
    configFile := datastore.GetConfigFilePath()

    file, err := os.Open(configFile)
    if err != nil {
        http.Error(w, "File not found.", 404)
        return
    }
    defer file.Close()
    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Disposition", "attachment; filename=config.yaml")

    if _, err := io.Copy(w, file); err != nil {
        http.Error(w, "Error sending file.", 500)
    }
}

func UploadConfigHandler(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, "Error parsing form.", 400)
        return
    }

    file, _, err := r.FormFile("configFile")
    if err != nil {
        http.Error(w, "Invalid file.", 400)
        return
    }
    defer file.Close()

    configFile := datastore.GetConfigFilePath()

    savedFile, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        http.Error(w, "Error saving file.", 500)
        return
    }
    defer savedFile.Close()
    if _, err := io.Copy(savedFile, file); err != nil {
        http.Error(w, "Error saving file.", 500)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
