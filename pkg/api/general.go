package api

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/evanphx/json-patch"
	"github.com/gin-gonic/gin"
)

func getMainConfig(c *gin.Context) {
	cfg := datastore.GetConfig()
	c.JSON(http.StatusOK, cfg.General)
}

func updateMainConfig(c *gin.Context) {
	var newGeneralConfig datastore.GeneralConfigStruct
	if err := c.ShouldBindJSON(&newGeneralConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := datastore.GetConfig()
	cfg.General = newGeneralConfig
	if err := datastore.UpdateConfig(*cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}

func getLogConfig(c *gin.Context) {
	cfg := datastore.GetConfig()
	c.JSON(http.StatusOK, cfg.Log)
}

func updateLogConfig(c *gin.Context) {
	originalConfig := datastore.GetConfig().Log

	originalConfigJSON, err := json.Marshal(originalConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal original config"})
		return
	}
	requestBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

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

	var updatedLogConfig datastore.LogAppConfigStruct
	if err := json.Unmarshal(updatedConfigJSON, &updatedLogConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal updated config"})
		return
	}
	datastore.SetLogConfig(updatedLogConfig)

	c.JSON(http.StatusOK, updatedLogConfig)
}

func getAuditConfig(c *gin.Context) {
	cfg := datastore.GetConfig()
	c.JSON(http.StatusOK, cfg.Audit)
}

func updateAuditConfig(c *gin.Context) {
	originalConfig := datastore.GetConfig().Audit

	originalConfigJSON, err := json.Marshal(originalConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal original config"})
		return
	}
	requestBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

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

	var updatedAuditConfig datastore.LogAuditConfigStruct
	if err := json.Unmarshal(updatedConfigJSON, &updatedAuditConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal updated config"})
		return
	}
	datastore.SetAuditConfig(updatedAuditConfig)

	c.JSON(http.StatusOK, updatedAuditConfig)
}

func getWebConfig(c *gin.Context) {
	cfg := datastore.GetConfig()
	c.JSON(http.StatusOK, cfg.WebServer)
}

func updateWebConfig(c *gin.Context) {
	originalConfig := datastore.GetConfig().WebServer

	originalConfigJSON, err := json.Marshal(originalConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal original config"})
		return
	}
	requestBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

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

	var updatedWebConfig datastore.WebServerConfigStruct
	if err := json.Unmarshal(updatedConfigJSON, &updatedWebConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal updated config"})
		return
	}
	datastore.SetWebConfig(updatedWebConfig)

	c.JSON(http.StatusOK, updatedWebConfig)
}

func getSegmentConfig(c *gin.Context) {
	cfg := datastore.GetConfig()
	c.JSON(http.StatusOK, cfg.SegmentConfigs)
}

func getSpecificSegmentConfig(c *gin.Context) {
	segmentName := c.Param("name")
	cfg := datastore.GetConfig()

	for _, segment := range cfg.SegmentConfigs {
		if segment.Name == segmentName {
			c.JSON(http.StatusOK, segment)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
}

func updateSpecificSegmentConfig(c *gin.Context) {
	segmentName := c.Param("name")
	cfg := datastore.GetConfig()

	var segmentUpdate map[string]interface{}
	if err := c.ShouldBindJSON(&segmentUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated := false
	for i, segment := range cfg.SegmentConfigs {
		if segment.Name == segmentName {
			if newName, ok := segmentUpdate["Name"].(string); ok && newName != "" {
				cfg.SegmentConfigs[i].Name = newName
			}
			if newPath, ok := segmentUpdate["Path"].(string); ok {
				cfg.SegmentConfigs[i].Path = newPath
			}
			updated = true
			c.JSON(http.StatusOK, cfg.SegmentConfigs[i])
			return
		}
	}

	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	if err := datastore.UpdateConfig(*cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func createSegmentConfig(c *gin.Context) {
	var newSegment datastore.SegmentConfigsStruct
	if err := c.ShouldBindJSON(&newSegment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := datastore.GetConfig()

	for _, segment := range cfg.SegmentConfigs {
		if segment.Name == newSegment.Name {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Segment already exists"})
			return
		}
	}

	cfg.SegmentConfigs = append(cfg.SegmentConfigs, newSegment)

	if err := datastore.UpdateConfig(*cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newSegment)
}

func deleteSegmentConfig(c *gin.Context) {
	segmentName := c.Param("name")
	cfg := datastore.GetConfig()

	found := false
	for i, segment := range cfg.SegmentConfigs {
		if segment.Name == segmentName {
			cfg.SegmentConfigs = append(cfg.SegmentConfigs[:i], cfg.SegmentConfigs[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	if err := datastore.UpdateConfig(*cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Segment deleted successfully"})
}

func DownloadConfigHandler(c *gin.Context) {
	configPath := datastore.GetConfigFilePath()
	c.FileAttachment(configPath, "config.yaml")
}

func UploadConfigHandler(c *gin.Context) {	
	file, err := c.FormFile("configFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	configPath := datastore.GetConfigFilePath()
    fmt.Printf("configPath: %s", configPath)
	if err := c.SaveUploadedFile(file, configPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the uploaded file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func getApiConfig(c *gin.Context) {
	cfg := datastore.GetConfig()
	c.JSON(http.StatusOK, cfg.Api)
}

func updateApiConfig(c *gin.Context) {
	originalConfig := datastore.GetConfig().Api

	originalConfigJSON, err := json.Marshal(originalConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal original config"})
		return
	}
	requestBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

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

	var updatedWebConfig datastore.APIConfigStruct
	if err := json.Unmarshal(updatedConfigJSON, &updatedWebConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal updated config"})
		return
	}
	datastore.SetApiConfig(updatedWebConfig)

	c.JSON(http.StatusOK, updatedWebConfig)
}