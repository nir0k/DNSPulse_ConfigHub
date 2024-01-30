package handlers

import (
	"ConfigHub/pkg/datastore"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
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

func DownloadConfigHandler(w http.ResponseWriter, r *http.Request) {
    // Specify the path to your configuration file
    
    // configFile := "path/to/your/config.file"
    configFile := datastore.GetConfigFilePath()

    // Open the file
    file, err := os.Open(configFile)
    if err != nil {
        http.Error(w, "File not found.", 404)
        return
    }
    defer file.Close()
    // Set the Content-Type header
    w.Header().Set("Content-Type", "application/octet-stream")
    // Suggest a filename
    w.Header().Set("Content-Disposition", "attachment; filename=config.yaml")

    // Copy the file content to the response writer
    if _, err := io.Copy(w, file); err != nil {
        http.Error(w, "Error sending file.", 500)
    }
}

func UploadConfigHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the multipart form, 32 << 20 specifies the maximum upload size
    if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, "Error parsing form.", 400)
        return
    }

    // Retrieve the file from the form data
    file, _, err := r.FormFile("configFile")
    if err != nil {
        http.Error(w, "Invalid file.", 400)
        return
    }
    defer file.Close()

    // Save the file to the server
    // Create a new file in the desired directory
    configFile := datastore.GetConfigFilePath()

    savedFile, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        http.Error(w, "Error saving file.", 500)
        return
    }
    defer savedFile.Close()
    // Copy the uploaded file to the new file
    if _, err := io.Copy(savedFile, file); err != nil {
        http.Error(w, "Error saving file.", 500)
        return
    }

    // Optionally, respond to the client that the upload was successful
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
