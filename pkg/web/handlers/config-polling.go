package handlers

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type SegmentUpdateData struct {
	SegmentName string          `json:"segmentName"`
	Data        []datastore.Csv `json:"data"`
}

func PollingSegmentHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pollingConfigs := datastore.GetPollingHosts()

		data := map[string]interface{}{
			"ShowNavBar": true,
			"Title":      "Polling Configurations",
			"Segments":   *pollingConfigs,
		}

		err := tmpl.ExecuteTemplate(w, "config-polling.html", data)
    if err != nil {
			http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	segmentName := r.URL.Query().Get("segment")
	switch r.Method {
	case "POST":
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusInternalServerError)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving the file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		uploadPath := datastore.GetSegmentsPollingConfigbySegment(segmentName).Path
		if uploadPath == "" {
			http.Error(w, "Segment not found", http.StatusNotFound)
			return
		}

		dst, err := os.Create(uploadPath)
		if err != nil {
			http.Error(w, "Error creating the file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err = io.Copy(dst, file); err != nil {
			http.Error(w, "Error saving the file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		datastore.LoadSegmentPollingHosts()

		fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)

	case "GET":
		filePath := datastore.GetSegmentsPollingConfigbySegment(segmentName).Path
		fileName := filepath.Base(filePath)
		if filePath == "" {
			http.Error(w, "File name is required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(fileName))
		w.Header().Set("Content-Type", "application/octet-stream")

		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Error opening the file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		if _, err = io.Copy(w, file); err != nil {
			http.Error(w, "Error writing file to response: "+err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}

func UpdateSegmentDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var updateData SegmentUpdateData
	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("updateData.SegmentName: %v\n", updateData.SegmentName)
	fmt.Printf("updateData: %v\n", updateData.Data)

	updatedData, err := datastore.UpdatePollingHostsBySegment(updateData.SegmentName, updateData.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("updatedData: %v\n", updatedData)
	responseData := map[string]interface{}{
		"message":     "Data updated successfully",
		"updatedData": updatedData,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
