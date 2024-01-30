package datastore

import (
	"ConfigHub/pkg/logger"
	"ConfigHub/pkg/tools"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)


type GeneralConfigStruct struct {
	Path	string `json:"path"`
}

type UpdateConfigRequest struct {
    General GeneralConfigStruct `json:"general"`
}

type LogAppConfigStruct struct {
	Path 		string	`json:"path"`
	MinSeverity string	`json:"minSeverity"`
	MaxAge 		int		`json:"maxAge"`
	MaxSize 	int		`json:"maxSize"`
	MaxFiles 	int		`json:"maxFiles"`
}

type LogAuditConfigStruct struct {
	Path 		string	`yaml:"path" json:"path"`
	MinSeverity string	`yaml:"minSeverity" json:"minSeverity"`
	MaxAge 		int		`yaml:"maxAge" json:"maxAge"`
	MaxSize 	int		`yaml:"maxSize" json:"maxSize"`
	MaxFiles 	int		`yaml:"maxFiles" json:"maxFiles"`
}

type WebServerConfigStruct struct {
	Port 			int		`yaml:"port" json:"port"`
	ListenAddress 	string	`yaml:"listenAddress" json:"listenAddress"`
	SSLEnabled 		bool	`yaml:"sslIsEnable" json:"sslIsEnable"`
	SSLCertPath 	string	`yaml:"sslCertPath" json:"sslCertPath"`
	SSLKeyPath 		string	`yaml:"sslKeyPath" json:"sslKeyPath"`
	SesionTimeout 	int		`yaml:"sesionTimeout" json:"sesionTimeout"`
	Username 		string	`yaml:"username" json:"username"`
	Password 		string	`yaml:"password" json:"password"`
}

type SegmentConfigsStruct struct {
	Name	string	`yaml:"name"`
	Path	string	`yaml:"path"`
}

type ConfigStruct struct {
	General			GeneralConfigStruct
	Log				LogAppConfigStruct 
	Audit			LogAuditConfigStruct	`yaml:"Audit"`
	WebServer		WebServerConfigStruct	`yaml:"WebServer"`
	SegmentConfigs	[]SegmentConfigsStruct	`yaml:"Configs"`
}

type HashStruct struct {
    LastHash    string  `json:"Hash"`
    LastUpdate  int64   `json:"LastUpdate"`
}


var (
	configFile	string
	config		ConfigStruct
	configMutex sync.RWMutex
	lastConfigHash HashStruct
)


func SetConfigFilePath(path string){
	configMutex.Lock()
    configFile = path
	logger.Logger.Debugf("Setup new path to configuration file: %s", configFile)
    configMutex.Unlock()
}

func GetConfigFilePath() string{
	configMutex.RLock()
    defer configMutex.RUnlock()
    return configFile
}

func SetLogConfig(logconfig LogAppConfigStruct){
	configMutex.Lock()
    config.Log = logconfig
	logger.Logger.Debugf("Setup new Log configuration: %v", config.Log)
    configMutex.Unlock()
}

func LoadConfig() (bool, error) {
	fileData, err := os.ReadFile(configFile)
    if err != nil {
        return false, err
    }
    newHash, err := tools.CalculateHash(string(configFile))
    if err != nil {
        logger.Logger.Errorf("Error Calculate hash to file '%s' err: %v\n", configFile, err)
        return false, err
    }
    if lastConfigHash.LastHash == newHash {
        logger.Logger.Debug("Configuration file has not been changed")
        return false, nil
    }
    logger.Logger.Infof("Configuration file has been changed")

	var newConfig ConfigStruct
    if err := yaml.Unmarshal(fileData, &newConfig); err != nil {
        return false, err
    }

    newConfig.Log = config.Log
	newConfig.General.Path = configFile
	configMutex.Lock()
    config = newConfig
    configMutex.Unlock()
	lastConfigHash.LastHash = newHash
    lastConfigHash.LastUpdate = time.Now().Unix()
	logger.Logger.Debugf("Configurations: %v\n", config)

    return true, nil
}

func GetConfig() *ConfigStruct {
    configMutex.RLock()
    defer configMutex.RUnlock()
    return &config
}

func UpdateConfig(newConf ConfigStruct) error {
	configMutex.RLock()
    defer configMutex.RUnlock()
	config = newConf
	err := SaveConfigToFile()
	return err
}

func SaveConfigToFile() error {
    configMutex.RLock()
    defer configMutex.RUnlock()

    fileData, err := yaml.Marshal(config)
    if err != nil {
        return err
    }

    // Write to a temporary file first
    tempFile := configFile + ".tmp"
    if err := os.WriteFile(tempFile, fileData, 0644); err != nil {
        return err
    }

    // Rename temporary file to the actual config file
    return os.Rename(tempFile, configFile)
}