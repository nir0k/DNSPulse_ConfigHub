package handlers

import (
	"ConfigHub/pkg/datastore"
	"encoding/json"
	"html/template"
	"net/http"
)

func ConfigGeneralHandler(tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        configData := datastore.GetConfig()

        templateData := map[string]interface{}{
            "ShowNavBar": true,      // This is existing data you were passing
            "Config":     configData, // Pass the entire configuration structure
        }

        err := tmpl.ExecuteTemplate(w, "config-general.html", templateData)
        if err != nil {
            http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
            return
        }
    }
}

func UpdateGeneralConfigHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the JSON request body into a struct representing the updated configuration
    var updatedConfig datastore.UpdateConfigRequest // Define a struct that matches your configuration structure
    err := json.NewDecoder(r.Body).Decode(&updatedConfig)
    if err != nil {
        http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Update the configuration in your datastore (datastore.UpdateGeneralConfig)
	newConf := datastore.GetConfig()
	newConf.General = updatedConfig.General
    if err := datastore.UpdateConfig(*newConf); err != nil {
        http.Error(w, "Failed to update general configuration: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Respond with a success message
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
    })
}

func UpdateLogConfigHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the JSON request body into a struct representing the updated configuration
    type EditRequest struct {
        Data datastore.LogAppConfigStruct `json:"data"`
    }
    var updatedConfig EditRequest // Define a struct that matches your configuration structure
    err := json.NewDecoder(r.Body).Decode(&updatedConfig)
    if err != nil {
        http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Update the configuration in your datastore (datastore.UpdateGeneralConfig)
	newConf := datastore.GetConfig()
	newConf.Log = updatedConfig.Data
    if err := datastore.UpdateConfig(*newConf); err != nil {
        http.Error(w, "Failed to update log configuration: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Respond with a success message
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
    })
}

func UpdateAuditConfigHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the JSON request body into a struct representing the updated configuration
    type EditRequest struct {
        Data datastore.LogAuditConfigStruct `json:"data"`
    }
    var updatedConfig EditRequest // Define a struct that matches your configuration structure
    err := json.NewDecoder(r.Body).Decode(&updatedConfig)
    if err != nil {
        http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Update the configuration in your datastore (datastore.UpdateGeneralConfig)
	newConf := datastore.GetConfig()
	newConf.Audit = updatedConfig.Data
    if err := datastore.UpdateConfig(*newConf); err != nil {
        http.Error(w, "Failed to update audit configuration: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Respond with a success message
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
    })
}

func UpdateWebServerConfigHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the JSON request body into a struct representing the updated configuration
    type EditRequest struct {
        Data datastore.WebServerConfigStruct `json:"data"`
    }
    var updatedConfig EditRequest // Define a struct that matches your configuration structure
    err := json.NewDecoder(r.Body).Decode(&updatedConfig)
    if err != nil {
        http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Update the configuration in your datastore (datastore.UpdateGeneralConfig)
	newConf := datastore.GetConfig()
    if updatedConfig.Data.Password == "" {
        updatedConfig.Data.Password = newConf.WebServer.Password
    }
	newConf.WebServer = updatedConfig.Data
    if err := datastore.UpdateConfig(*newConf); err != nil {
        http.Error(w, "Failed to update web-server configuration: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Respond with a success message
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
    })
}