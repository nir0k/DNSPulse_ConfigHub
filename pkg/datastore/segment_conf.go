package datastore

import (
	"ConfigHub/pkg/logger"
	"ConfigHub/pkg/tools"
	"os"
	"sync"
	// "time"

	"gopkg.in/yaml.v2"
)


type SegmentGeneralConfigStruct struct {
	CheckInterval 	int `yaml:"confCheckInterval"`
	Hash			string
}

type SegmentSyncConfigStruct struct {
	IsEnable 	bool 	`yaml:"isEnable"`
	Token 		string 	`yaml:"token"`
}

type PrometheusConfStruct struct {
	URL 			string						`yaml:"url"`
	AuthEnabled 	bool						`yaml:"auth"`
	Username 		string						`yaml:"username"`
	Password 		string						`yaml:"password"`
	MetricName 		string						`yaml:"metricName"`
	RetriesCount 	int							`yaml:"retriesCount"`
	BufferSize 		int							`yaml:"buferSize"`
	Labels 			PrometheusLabelConfigStruct `yaml:"labels"`
}

type PrometheusLabelConfigStruct struct {
	Opcode             bool `yaml:"opcode"`
    Authoritative      bool `yaml:"authoritative"`
    Truncated          bool `yaml:"truncated"`
    Rcode              bool `yaml:"rcode"`
    RecursionDesired   bool `yaml:"recursionDesired"`
    RecursionAvailable bool `yaml:"recursionAvailable"`
    AuthenticatedData  bool `yaml:"authenticatedData"`
    CheckingDisabled   bool `yaml:"checkingDisabled"`
    PollingRate        bool `yaml:"pollingRate"`
    Recursion          bool `yaml:"recursion"`
}

type PollingConfigStruct struct {
	Path 			string	`yaml:"path"`
	Hash			string
	Delimeter 		string	`yaml:"delimeter"`
	ExtraDelimeter 	string	`yaml:"extraDelimeter"`
	PollTimeout 	int		`yaml:"pullTimeout"`
}


type SegmentConfStruct struct {
	General		SegmentGeneralConfigStruct	`yaml:"General"`
	Sync		SegmentSyncConfigStruct		`yaml:"Sync"`
	Prometheus 	PrometheusConfStruct		`yaml:"Prometheus"`
	Polling 	PollingConfigStruct			`yaml:"Resolvers"`
}

type SegmentsMap map[string]SegmentConfStruct

var (
	// segmentConfigFile		string
	segmentConfig			SegmentsMap
	segmentConfigMutex 		sync.RWMutex
	// segmentLastConfigHash 	HashStruct
)

func init() {
	segmentConfig = make(map[string]SegmentConfStruct)
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
    if segmentConfig[s.Name].General.Hash == newHash {
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
    segmentConfig[s.Name] = newConfig
    segmentConfigMutex.Unlock()
	// segmentLastConfigHash.LastHash = newHash
    // segmentLastConfigHash.LastUpdate = time.Now().Unix()
	logger.Logger.Debugf("Configurations for segment %s: %v\n", s.Name, segmentConfig[s.Name])

    return true, nil
}

func GetSegmentsConfig() *SegmentsMap {
    segmentConfigMutex.RLock()
    defer segmentConfigMutex.RUnlock()
    return &segmentConfig
}