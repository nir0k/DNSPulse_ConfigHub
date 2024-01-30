package handlers

import (
	"ConfigHub/pkg/datastore"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
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
    // if err != nil {
    //     logger.Logger.Errorf("Failed to update segment config: %v", err)
    //     http.Error(w, "Failed to update segment config", http.StatusInternalServerError)
    //     return
    // }

    // Send a success response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}