package api

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/evanphx/json-patch"
	"github.com/gin-gonic/gin"
)

func getSegmentNamesHandler(c *gin.Context) {
	segmentNames := datastore.GetAllSegmentNames()
	c.JSON(http.StatusOK, gin.H{"segmentNames": segmentNames})
}

func getSegmentMainConfig(c *gin.Context) {
	segmentName := c.Param("name")
	cfg, _ := datastore.GetSegmentsConfigBySegment(segmentName)
	c.JSON(http.StatusOK, cfg.General)
}

func getSegmentPollingConfig(c *gin.Context) {
	segmentName := c.Param("name")
	cfg, _ := datastore.GetSegmentsConfigBySegment(segmentName)
	c.JSON(http.StatusOK, cfg.Polling)
}

func getSegmentSyncConfig(c *gin.Context) {
	segmentName := c.Param("name")
	cfg, _ := datastore.GetSegmentsConfigBySegment(segmentName)
	c.JSON(http.StatusOK, cfg.Sync)
}

func getSegmentPrometheusConfig(c *gin.Context) {
	segmentName := c.Param("name")
	cfg, _ := datastore.GetSegmentsConfigBySegment(segmentName)
	c.JSON(http.StatusOK, cfg.Prometheus)
}

func getSegmentPrometheusLabelsConfig(c *gin.Context) {
	segmentName := c.Param("name")
	cfg, _ := datastore.GetSegmentsConfigBySegment(segmentName)
	c.JSON(http.StatusOK, cfg.Prometheus.Labels)
}

func getSegmentPollingHostsConfig(c *gin.Context) {
	segmentName := c.Param("name")

	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("perPage", "10")
	page, _ := strconv.Atoi(pageStr)
	perPage, _ := strconv.Atoi(perPageStr)

	allServers, exists := datastore.GetPollingHostsBySegment(segmentName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	query := c.Request.URL.Query()
	filteredServers := filterServers(allServers, query)

	totalRecords := len(filteredServers)
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))
	startIndex := (page - 1) * perPage
	endIndex := startIndex + perPage
	if endIndex > totalRecords {
		endIndex = totalRecords
	}

	paginatedServers := filteredServers[startIndex:endIndex]

	pagination := gin.H{
		"total_records": totalRecords,
		"current_page":  page,
		"total_pages":   totalPages,
		"next_page":     nil,
		"prev_page":     nil,
	}
	if page < totalPages {
		pagination["next_page"] = page + 1
	}
	if page > 1 {
		pagination["prev_page"] = page - 1
	}

	response := gin.H{
		"data":       paginatedServers,
		"pagination": pagination,
	}
	c.JSON(http.StatusOK, response)
}

func updateSegmentMainConfig(c *gin.Context) {
	segmentName := c.Param("name")
	segmentConfig, exists := datastore.GetSegmentsConfigBySegment(segmentName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	originalConfigJSON, err := json.Marshal(segmentConfig.General)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal original config"})
		return
	}

	requestBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var updatedConfig datastore.SegmentGeneralConfigStruct
	if c.Request.Method == "PATCH" {
		patch, err := jsonpatch.CreateMergePatch([]byte("{}"), requestBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create merge patch"})
			return
		}

		updatedConfigJSON, err := jsonpatch.MergePatch(originalConfigJSON, patch)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to apply merge patch"})
			return
		}

		if err := json.Unmarshal(updatedConfigJSON, &updatedConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal updated config"})
			return
		}
	} else {
		if err := json.Unmarshal(requestBody, &updatedConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal request body"})
			return
		}
	}

	segmentConfig.General = updatedConfig
	if err := datastore.UpdateSegmentConfig(segmentConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, segmentConfig.General)
}

func updateSegmentSyncConfig(c *gin.Context) {
	segmentName := c.Param("name")
	segmentConfig, exists := datastore.GetSegmentsConfigBySegment(segmentName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	originalConfigJSON, err := json.Marshal(segmentConfig.Sync)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal original config"})
		return
	}

	requestBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var updatedConfig datastore.SegmentSyncConfigStruct
	if c.Request.Method == "PATCH" {
		patch, err := jsonpatch.CreateMergePatch([]byte("{}"), requestBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create merge patch"})
			return
		}

		updatedConfigJSON, err := jsonpatch.MergePatch(originalConfigJSON, patch)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to apply merge patch"})
			return
		}

		if err := json.Unmarshal(updatedConfigJSON, &updatedConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal updated config"})
			return
		}
	} else {
		if err := json.Unmarshal(requestBody, &updatedConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal request body"})
			return
		}
	}

	segmentConfig.Sync = updatedConfig
	if err := datastore.UpdateSegmentConfig(segmentConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, segmentConfig.Sync)
}

func updateSegmentPollingConfig(c *gin.Context) {
	segmentName := c.Param("name")
	segmentConfig, exists := datastore.GetSegmentsConfigBySegment(segmentName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	originalConfigJSON, err := json.Marshal(segmentConfig.Polling)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal original config"})
		return
	}

	requestBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var updatedConfig datastore.PollingConfigStruct
	if c.Request.Method == "PATCH" {
		patch, err := jsonpatch.CreateMergePatch([]byte("{}"), requestBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create merge patch"})
			return
		}

		updatedConfigJSON, err := jsonpatch.MergePatch(originalConfigJSON, patch)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to apply merge patch"})
			return
		}

		if err := json.Unmarshal(updatedConfigJSON, &updatedConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal updated config"})
			return
		}
	} else {
		if err := json.Unmarshal(requestBody, &updatedConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal request body"})
			return
		}
	}

	segmentConfig.Polling = updatedConfig
	if err := datastore.UpdateSegmentConfig(segmentConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, segmentConfig.Polling)
}

func updateSegmentPrometheusConfig(c *gin.Context) {
	segmentName := c.Param("name")
	segmentConfig, exists := datastore.GetSegmentsConfigBySegment(segmentName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	originalConfigJSON, err := json.Marshal(segmentConfig.Prometheus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal original config"})
		return
	}

	requestBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var updatedConfig datastore.PrometheusConfStruct
	if c.Request.Method == "PATCH" {
		patch, err := jsonpatch.CreateMergePatch([]byte("{}"), requestBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create merge patch"})
			return
		}

		updatedConfigJSON, err := jsonpatch.MergePatch(originalConfigJSON, patch)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to apply merge patch"})
			return
		}

		if err := json.Unmarshal(updatedConfigJSON, &updatedConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal updated config"})
			return
		}
	} else {
		if err := json.Unmarshal(requestBody, &updatedConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal request body"})
			return
		}
	}

	segmentConfig.Prometheus = updatedConfig
	if err := datastore.UpdateSegmentConfig(segmentConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, segmentConfig.Prometheus)
}

func updateSegmentPrometheusLabelConfig(c *gin.Context) {
	segmentName := c.Param("name")
	segmentConfig, exists := datastore.GetSegmentsConfigBySegment(segmentName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	originalConfigJSON, err := json.Marshal(segmentConfig.Prometheus.Labels)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal original config"})
		return
	}

	requestBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var updatedConfig datastore.PrometheusLabelConfigStruct
	if c.Request.Method == "PATCH" {
		patch, err := jsonpatch.CreateMergePatch([]byte("{}"), requestBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create merge patch"})
			return
		}

		updatedConfigJSON, err := jsonpatch.MergePatch(originalConfigJSON, patch)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to apply merge patch"})
			return
		}

		if err := json.Unmarshal(updatedConfigJSON, &updatedConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal updated config"})
			return
		}
	} else {
		if err := json.Unmarshal(requestBody, &updatedConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal request body"})
			return
		}
	}

	segmentConfig.Prometheus.Labels = updatedConfig
	if err := datastore.UpdateSegmentConfig(segmentConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, segmentConfig.Prometheus.Labels)
}

func getPollingServerConfig(c *gin.Context) {
	segmentName := c.Param("name")
	serverName := c.Param("srvname")

	servers, exists := datastore.GetPollingHostsBySegment(segmentName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	for _, server := range servers {
		if server.Server == serverName {
			c.JSON(http.StatusOK, server)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
}

func updatePollingServerConfig(c *gin.Context) {
	segmentName := c.Param("name")
	serverName := c.Param("srvname")

	servers, exists := datastore.GetPollingHostsBySegment(segmentName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	updatedServerIndex := -1
	for i, server := range servers {
		if server.Server == serverName {
			updatedServerIndex = i
			if c.Request.Method == "PATCH" {
				var changes map[string]interface{}
				if err := c.ShouldBindJSON(&changes); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				serverJSON, _ := json.Marshal(server)
				changesJSON, _ := json.Marshal(changes)
				updatedServerJSON, err := jsonpatch.MergePatch(serverJSON, changesJSON)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply changes"})
					return
				}
				json.Unmarshal(updatedServerJSON, &servers[i])
			} else {
				var newServerConfig datastore.Csv
				if err := c.ShouldBindJSON(&newServerConfig); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				if newServerConfig.Server == "" || newServerConfig.IPAddress == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
					return
				}
				servers[i] = newServerConfig
			}
			break
		}
	}

	if updatedServerIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	if _, err := datastore.UpdatePollingHostsBySegment(segmentName, servers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, servers[updatedServerIndex])
}

func addPollingServerConfig(c *gin.Context) {
	segmentName := c.Param("name")

	servers, exists := datastore.GetPollingHostsBySegment(segmentName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	var newServer datastore.Csv
	if err := c.ShouldBindJSON(&newServer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, server := range servers {
		if server.Server == newServer.Server {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Server already exists"})
			return
		}
	}

	servers = append(servers, newServer)

	_, err := datastore.UpdatePollingHostsBySegment(segmentName, servers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newServer)
}

func deletePollingServerConfig(c *gin.Context) {
	segmentName := c.Param("name")
	serverName := c.Param("srvname")

	servers, exists := datastore.GetPollingHostsBySegment(segmentName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	found := false
	for i, server := range servers {
		if server.Server == serverName {
			servers = append(servers[:i], servers[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	_, err := datastore.UpdatePollingHostsBySegment(segmentName, servers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Server deleted successfully"})
}

func filterServers(servers []datastore.Csv, query map[string][]string) []datastore.Csv {
	filtered := make([]datastore.Csv, 0, len(servers))
	for _, server := range servers {
		match := true
		for key, value := range query {
			if !fieldMatches(server, key, value[0]) {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, server)
		}
	}
	return filtered
}

func fieldMatches(server datastore.Csv, field, queryValue string) bool {
	v := reflect.ValueOf(server)
	f := reflect.Indirect(v).FieldByName(field)
	if !f.IsValid() {
		return false
	}
	fieldValue := strings.ToLower(f.String())
	queryValueLower := strings.ToLower(queryValue)
	return strings.Contains(fieldValue, queryValueLower)
}

func DownloadSegmentConfigHandler(c *gin.Context) {
	segmentName := c.Param("name")
	configPath := datastore.GetSegmentConfigFilePath(segmentName)
	filename := fmt.Sprintf("%s.yaml", segmentName)
	c.FileAttachment(configPath, filename)
}

func UploadSegmentConfigHandler(c *gin.Context) {
	file, err := c.FormFile("configFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	segmentName := c.Param("name")
	configPath := datastore.GetSegmentConfigFilePath(segmentName)
	if err := c.SaveUploadedFile(file, configPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the uploaded file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func DownloadSegmentPollingHandler(c *gin.Context) {
	segmentName := c.Param("name")
	configPath := datastore.GetSegmentsPollingConfigbySegment(segmentName).Path
	filename := fmt.Sprintf("%s.csv", segmentName)
	c.FileAttachment(configPath, filename)
}

func UploadSegmentPollingHandler(c *gin.Context) {
	file, err := c.FormFile("configFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	segmentName := c.Param("name")
	configPath := datastore.GetSegmentsPollingConfigbySegment(segmentName).Path
	if err := c.SaveUploadedFile(file, configPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the uploaded file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}
