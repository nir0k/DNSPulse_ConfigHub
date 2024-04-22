package handlers2

import (
	"DNSPulse_ConfigHub/pkg/logger"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

// Structure to hold segment configuration details
type SegmentConfigStruct struct {
	Name          string
	ConfCheckInterval int `json:"confCheckInterval"`
	Hash           string
	SyncIsEnable   bool
	SyncToken      string
	PollingPath    string
	PollingHash    string
	Delimeter      string
	ExtraDelimeter string
	PullTimeout    int
	PrometheusURL  string
	PrometheusAuth bool
	PrometheusUsername string
	PrometheusPassword string
	MetricName     string
	RetriesCount   int
	BuferSize      int
	Labels         map[string]bool
}

type PageData struct {
    Title    string
    Segments []SegmentConfigStruct
    ShowNavBar bool
}

func ConfigSegmentHandler(tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tokenString, err := getTokenFromRequest(r)
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        client := &http.Client{}

        // Fetch the list of segments
        segmentNames, err := fetchSegments(client, tokenString) // Corrected: `segmentNames` is now a slice of strings
        if err != nil {
            http.Error(w, "Failed to fetch segment list", http.StatusInternalServerError)
            return
        }

        var configs []SegmentConfigStruct // Corrected: Using the right type for slice
        for _, segmentName := range segmentNames { // Corrected: Iterating directly over the slice of strings
            config, err := fetchSegmentConfig(client, tokenString, segmentName)
            if err != nil {
                http.Error(w, "Failed to fetch segment configuration for "+segmentName, http.StatusInternalServerError)
                return
            }
            configs = append(configs, config)
        }

        pageData := PageData{
            Title:      "Segment Configurations",
            Segments:   configs, // Assuming 'configs' is your slice of SegmentConfigStruct
            ShowNavBar: true,    // Set this based on your application logic
        }

        err = tmpl.ExecuteTemplate(w, "config-segment.html", pageData)
        if err != nil {
            logger.Logger.Errorf("Error executing template: %v", err)
            http.Error(w, "Error rendering page", http.StatusInternalServerError)
            return
        }
    }
}

func getTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err // Token not found in cookie
	}
	return cookie.Value, nil
}


func fetchSegments(client *http.Client, token string) ([]string, error) {
	req, err := http.NewRequest("GET", "http://localhost:8080/api/config/segment", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch segments, status code: %d", resp.StatusCode)
	}

	var result struct {
		SegmentNames []string `json:"segmentNames"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.SegmentNames, nil
}


func fetchSegmentConfig(client *http.Client, token string, segmentName string) (SegmentConfigStruct, error) {
    var config SegmentConfigStruct
    config.Name = segmentName

    // Define endpoints for fetching segment configuration details
    endpoints := []string{
        "/api/config/segment/" + segmentName + "/main",
        "/api/config/segment/" + segmentName + "/sync",
        "/api/config/segment/" + segmentName + "/polling",
        "/api/config/segment/" + segmentName + "/prometheus",
    }

    // Iterate through endpoints to fetch and update config
    for _, endpoint := range endpoints {
        req, err := http.NewRequest("GET", "http://localhost:8080"+endpoint, nil)
        if err != nil {
            return SegmentConfigStruct{}, err
        }

        req.Header.Add("Authorization", "Bearer "+token)
        resp, err := client.Do(req)
        if err != nil {
            return SegmentConfigStruct{}, err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            return SegmentConfigStruct{}, fmt.Errorf("failed to fetch config for segment %s from endpoint %s, status code: %d", segmentName, endpoint, resp.StatusCode)
        }

        switch endpoint {
        case "/api/config/segment/" + segmentName + "/main":
            var mainConfig struct {
                ConfCheckInterval int    `json:"confCheckInterval"`
                Hash              string `json:"Hash"`
            }
            if err := json.NewDecoder(resp.Body).Decode(&mainConfig); err != nil {
                return SegmentConfigStruct{}, err
            }
            config.ConfCheckInterval = mainConfig.ConfCheckInterval
            config.Hash = mainConfig.Hash
        case "/api/config/segment/" + segmentName + "/sync":
            var syncConfig struct {
                IsEnable bool   `json:"isEnable"`
                Token    string `json:"token"`
            }
            if err := json.NewDecoder(resp.Body).Decode(&syncConfig); err != nil {
                return SegmentConfigStruct{}, err
            }
            config.SyncIsEnable = syncConfig.IsEnable
            config.SyncToken = syncConfig.Token
        case "/api/config/segment/" + segmentName + "/polling":
            var pollingConfig struct {
                Path            string `json:"path"`
                Hash            string `json:"Hash"`
                Delimeter       string `json:"delimeter"`
                ExtraDelimeter  string `json:"extraDelimeter"`
                PullTimeout     int    `json:"pullTimeout"`
            }
            if err := json.NewDecoder(resp.Body).Decode(&pollingConfig); err != nil {
                return SegmentConfigStruct{}, err
            }
            config.PollingPath = pollingConfig.Path
            config.PollingHash = pollingConfig.Hash
            config.Delimeter = pollingConfig.Delimeter
            config.ExtraDelimeter = pollingConfig.ExtraDelimeter
            config.PullTimeout = pollingConfig.PullTimeout
        case "/api/config/segment/" + segmentName + "/prometheus":
            var prometheusConfig struct {
                URL            string            `json:"url"`
                Auth           bool              `json:"auth"`
                Username       string            `json:"username"`
                Password       string            `json:"password"`
                MetricName     string            `json:"metricName"`
                RetriesCount   int               `json:"retriesCount"`
                BuferSize      int               `json:"buferSize"`
                Labels         map[string]bool   `json:"labels"`
            }
            if err := json.NewDecoder(resp.Body).Decode(&prometheusConfig); err != nil {
                return SegmentConfigStruct{}, err
            }
            config.PrometheusURL = prometheusConfig.URL
            config.PrometheusAuth = prometheusConfig.Auth
            config.PrometheusUsername = prometheusConfig.Username
            config.PrometheusPassword = prometheusConfig.Password
            config.MetricName = prometheusConfig.MetricName
            config.RetriesCount = prometheusConfig.RetriesCount
            config.BuferSize = prometheusConfig.BuferSize
            config.Labels = prometheusConfig.Labels
        }
    }

    return config, nil
}


