package handlers

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	"fmt"
	// "DNSPulse_ConfigHub/pkg/logger"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
)

func ConfigSegmentHandler(tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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

// func UpdateSegmentConfigHandler(w http.ResponseWriter, r *http.Request) {
//     if r.Method != http.MethodPost {
//         http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
//         return
//     }

//     var updateReq datastore.SegmentConfStruct
//     err := json.NewDecoder(r.Body).Decode(&updateReq)
//     if err != nil {
//         http.Error(w, "Error decoding request body: "+err.Error(), http.StatusBadRequest)
//         return
//     }

//     err = datastore.UpdateSegmentConfig(updateReq)
//     if err != nil {
//         logger.Logger.Errorf("Failed to update segment config: %v", err)
//         http.Error(w, "Failed to update segment config", http.StatusInternalServerError)
//         return
//     }

//     w.WriteHeader(http.StatusOK)
//     json.NewEncoder(w).Encode(map[string]string{"status": "success"})
// }

func DownloadSegmentConfigHandler(w http.ResponseWriter, r *http.Request) {
    segmentName := r.URL.Query().Get("segment")

    filePath := datastore.GetSegmentConfigFilePath(segmentName)
    if filePath == "" {
        http.Error(w, "Segment not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Disposition", "attachment; filename="+segmentName+".yaml")
    w.Header().Set("Content-Type", "application/octet-stream")
    http.ServeFile(w, r, filePath)
}

func UploadSegmentConfigHandler(w http.ResponseWriter, r *http.Request) {
    segmentName := r.URL.Query().Get("segment")

    err := r.ParseMultipartForm(10 << 20)
    if err != nil {
        http.Error(w, "File too large", http.StatusBadRequest)
        return
    }

    file, _, err := r.FormFile("configFile")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    path := datastore.GetSegmentConfigFilePath(segmentName)
    if path == "" {
        http.Error(w, "Segment not found", http.StatusNotFound)
        return
    }
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

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func UpdateSegmentConfigHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Segments []datastore.SegmentConfStruct `json:"segments"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Iterate over the provided segments
    for _, updatedSegment := range req.Segments {
        // Validate the segment, for example, checking for a non-empty name
        if updatedSegment.SegmentName == "" {
            http.Error(w, "Segment name is required", http.StatusBadRequest)
            return
        }

        // Use datastore.UpdateSegmentConfig to update the segment configuration
        err := datastore.UpdateSegmentConfig(updatedSegment)
        if err != nil {
            // Handle errors, such as logging and returning an HTTP error response
            http.Error(w, fmt.Sprintf("Failed to update segment '%s': %v", updatedSegment.SegmentName, err), http.StatusInternalServerError)
            return
        }
    }

    // Respond to the request indicating success
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
    })
}

