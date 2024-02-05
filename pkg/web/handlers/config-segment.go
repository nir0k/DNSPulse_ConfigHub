package handlers

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	"DNSPulse_ConfigHub/pkg/logger"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
)

func ConfigSegmentHandler(tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Fetch segment configs
        segmentConfigs := datastore.GetSegmentsConfig()

        err := tmpl.ExecuteTemplate(w, "config-segment.html", map[string]interface{}{
            "ShowNavBar": true,
            "Title":    "Segment Configurations",
            "Segments": segmentConfigs,
        })
        if err != nil {
            http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
            return
        }
    }
}

func UpdateSegmentConfigHandler(w http.ResponseWriter, r *http.Request) {
    // Only accept POST requests
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var updateReq datastore.SegmentConfStruct
    err := json.NewDecoder(r.Body).Decode(&updateReq)
    if err != nil {
        http.Error(w, "Error decoding request body: "+err.Error(), http.StatusBadRequest)
        return
    }

    fmt.Printf("updateReq: %v\n", updateReq)



    // Perform validation on updateReq as necessary
    // Update the datastore with new configurations
    err = datastore.UpdateSegmentConfig(updateReq) // Implement this function in your datastore package
    if err != nil {
        logger.Logger.Errorf("Failed to update segment config: %v", err)
        http.Error(w, "Failed to update segment config", http.StatusInternalServerError)
        return
    }

    // Send a success response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func DownloadSegmentConfigHandler(w http.ResponseWriter, r *http.Request) {
    segmentName := r.URL.Query().Get("segment")

    // Find the configuration file path based on the segment name
    filePath := datastore.GetSegmentConfigFilePath(segmentName)
    if filePath == "" {
        http.Error(w, "Segment not found", http.StatusNotFound)
        return
    }

    // Serve the file for download
    w.Header().Set("Content-Disposition", "attachment; filename="+segmentName+".yaml")
    w.Header().Set("Content-Type", "application/octet-stream")
    http.ServeFile(w, r, filePath)
}

func UploadSegmentConfigHandler(w http.ResponseWriter, r *http.Request) {
    segmentName := r.URL.Query().Get("segment")

    // Parse the multipart form
    err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
    if err != nil {
        http.Error(w, "File too large", http.StatusBadRequest)
        return
    }

    // Retrieve the file from form data
    file, _, err := r.FormFile("configFile")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Process file (e.g., save to disk, validate, etc.)
    // Define the path where the file will be saved
    // path := "uploads/" + segmentName + "-" + header.Filename
    path := datastore.GetSegmentConfigFilePath(segmentName)
    if path == "" {
        http.Error(w, "Segment not found", http.StatusNotFound)
        return
    }
    // Create the file
    dst, err := os.Create(path)
    if err != nil {
        http.Error(w, "Unable to create the file for writing", http.StatusInternalServerError)
        return
    }
    defer dst.Close()
    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, "Unable to write the file to disk", http.StatusInternalServerError)
        return
    }

    // Respond with success
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}