package datastore

import (
	"ConfigHub/pkg/logger"
	"ConfigHub/pkg/tools"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)


type SegmentGeneralConfigStruct struct {
	CheckInterval 	int `yaml:"confCheckInterval" json:"confCheckInterval"`
	Hash			string `yaml:"-"`
}

type SegmentSyncConfigStruct struct {
	IsEnable 	bool 	`yaml:"isEnable" json:"isEnable"`
	Token 		string 	`yaml:"token" json:"token"` 
}

type PrometheusConfStruct struct {
	URL 			string						`yaml:"url" json:"url"`
	AuthEnabled 	bool						`yaml:"auth" json:"auth"`
	Username 		string						`yaml:"username" json:"username"`
	Password 		string						`yaml:"password" json:"password"`
	MetricName 		string						`yaml:"metricName" json:"metricName"`
	RetriesCount 	int							`yaml:"retriesCount" json:"retriesCount"`
	BufferSize 		int							`yaml:"buferSize" json:"buferSize"`
	Labels 			PrometheusLabelConfigStruct `yaml:"labels" json:"labels"`
}

type PrometheusLabelConfigStruct struct {
	Opcode             bool `yaml:"opcode" json:"opcode"`
    Authoritative      bool `yaml:"authoritative" json:"authoritative"`
    Truncated          bool `yaml:"truncated" json:"truncated"`
    Rcode              bool `yaml:"rcode" json:"rcode"`
    RecursionDesired   bool `yaml:"recursionDesired" json:"recursionDesired"`
    RecursionAvailable bool `yaml:"recursionAvailable" json:"recursionAvailable"`
    AuthenticatedData  bool `yaml:"authenticatedData" json:"authenticatedData"`
    CheckingDisabled   bool `yaml:"checkingDisabled" json:"checkingDisabled"`
    PollingRate        bool `yaml:"pollingRate" json:"pollingRate"`
    Recursion          bool `yaml:"recursion" json:"recursion"`
}

type PollingConfigStruct struct {
	Path 			string	`yaml:"path" json:"path"`
	Hash			string	`yaml:"-"`
	Delimeter 		string	`yaml:"delimeter" json:"delimeter"`
	ExtraDelimeter 	string	`yaml:"extraDelimeter" json:"extraDelimeter"`
	PollTimeout 	int		`yaml:"pullTimeout" json:"pullTimeout"`
}


type SegmentConfStruct struct {
	SegmentName string    					`json:"SegmentName"`
	General		SegmentGeneralConfigStruct	`yaml:"General" json:"General"`
	Sync		SegmentSyncConfigStruct		`yaml:"Sync" json:"Sync"`
	Prometheus 	PrometheusConfStruct		`yaml:"Prometheus" json:"Prometheus"`
	Polling 	PollingConfigStruct			`yaml:"Resolvers" json:"Resolvers"`
}

type SegmentsMap map[string]SegmentConfStruct

var (
	// segmentConfigFile		string
	segmentsConfig			SegmentsMap
	segmentConfigMutex 		sync.RWMutex
	// segmentLastConfigHash 	HashStruct
)

func init() {
	segmentsConfig = make(map[string]SegmentConfStruct)
}

func LoadSegmentConfigs() error {
	segments := GetConfig().SegmentConfigs
	for _, s := range segments {
		loadSegmentConfig(s)
	}
	return nil
}

func loadSegmentConfig(s SegmentConfigsStruct) (bool, error) {
	fileData, err := os.ReadFile(s.Path)
    if err != nil {
        return false, err
    }
    newHash, err := tools.CalculateHash(string(s.Path))
    if err != nil {
        logger.Logger.Errorf("Error Calculate hash to file '%s' (segment: %s) err: %v\n", s.Path, s.Name, err)
        return false, err
    }
    if segmentsConfig[s.Name].General.Hash == newHash {
        logger.Logger.Debugf("Configuration for %s has not been changed", s.Name)
        return false, nil
    }
    logger.Logger.Infof("Configuration file %s (segment: %s) has been changed", s.Path, s.Name)

	var newConfig SegmentConfStruct
    if err := yaml.Unmarshal(fileData, &newConfig); err != nil {
        return false, err
    }

    // newConfig.Log = config.Log
	// newConfig.General.Path = configFile
	newConfig.General.Hash = newHash
	segmentConfigMutex.Lock()
    segmentsConfig[s.Name] = newConfig
    segmentConfigMutex.Unlock()
	// segmentLastConfigHash.LastHash = newHash
    // segmentLastConfigHash.LastUpdate = time.Now().Unix()
	logger.Logger.Debugf("Configurations for segment %s: %v\n", s.Name, segmentsConfig[s.Name])

    return true, nil
}

func GetSegmentsConfig() *SegmentsMap {
    segmentConfigMutex.RLock()
    defer segmentConfigMutex.RUnlock()
    return &segmentsConfig
}

func GetSegmentsConfigBySegment(segmentName string) (SegmentConfStruct, bool) {
    segmentConfigMutex.RLock()
    defer segmentConfigMutex.RUnlock()
    config, ok := segmentsConfig[segmentName]
    return config, ok
}

func GetSegmentsPollingConfigbySegment(segmentName string) PollingConfigStruct{
    segmentConfigMutex.RLock()
    defer segmentConfigMutex.RUnlock()
	conf := segmentsConfig[segmentName]
	return conf.Polling
}

func GetSegmentsSyncConf(segmentName string) SegmentSyncConfigStruct{
    segmentConfigMutex.RLock()
    defer segmentConfigMutex.RUnlock()
	conf := segmentsConfig[segmentName]
	return conf.Sync
}

func UpdateSegmentConfig(newConf SegmentConfStruct) error {
	segmentConfigMutex.RLock()
    defer segmentConfigMutex.RUnlock()
	filePath, err := SaveSegmentConfigToFile(newConf)
	if err != nil {
		logger.Logger.Errorf("failure to save segment '%s' config into file, err: %v", newConf.SegmentName, err)
		return err
	}
	newHash, err := tools.CalculateHash(filePath)
	if err != nil {
		logger.Logger.Errorf("failure to calculate hash for segment '%s' config into file, err: %v", newConf.SegmentName, err)
		return err
	}
	newConf.General.Hash = newHash
	segmentsConfig[newConf.SegmentName] = newConf
	logger.Logger.Debugf("Succesfuly to calculate hash for segment '%s' config into file, err: %v", newConf.SegmentName, err)
	return nil
}

func SaveSegmentConfigToFile(segmentConfig SegmentConfStruct) (string, error) {
	filePath := GetSegmentConfigFilePath(segmentConfig.SegmentName)
	if filePath != "" {
		configMutex.RLock()
		defer configMutex.RUnlock()
		fileData, err := yaml.Marshal(segmentConfig)
		if err != nil {
			return "", err
		}
		// Write to a temporary file first
		tempFile := filePath + ".tmp"
		if err := os.WriteFile(tempFile, fileData, 0644); err != nil {
			return "", err
		}
		// Rename temporary file to the actual config file
		return filePath, os.Rename(tempFile, filePath)
	}
	return "", fmt.Errorf("failed to search segment '%s' config file path", segmentConfig.SegmentName)
}

func GetSegmentConfigFilePath(segmentName string) string {
	for _, s := range GetConfig().SegmentConfigs {
		if s.Name == segmentName {
			return s.Path
		}
	}
	return ""
}

func UpdatePollingHash(segmentName string, hash string) {
	segmentConfigMutex.RLock()
    defer segmentConfigMutex.RUnlock()
	newConf := segmentsConfig[segmentName]
	newConf.Polling.Hash = hash
	segmentsConfig[segmentName] = newConf
}