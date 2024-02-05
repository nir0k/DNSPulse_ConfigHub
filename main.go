package main

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	grpcserver "DNSPulse_ConfigHub/pkg/gRPC-server"
	"DNSPulse_ConfigHub/pkg/logger"
	"DNSPulse_ConfigHub/pkg/tools"

	"DNSPulse_ConfigHub/pkg/web"
	"flag"
	"os"
)


func setup() {
	var configFilePath = flag.String("config", "configs/config.yaml", "Path to the configuration file")
	var logPath = flag.String("logPath", "logs/log.json", "Path to the log file")
	var logSeverity = flag.String("logSeverity", "debug", "Min log severity")
	var logMaxSize = flag.Int("logMaxSize", 10, "Max size for log file (Mb)")
	var logMaxFiles = flag.Int("logMaxFiles", 10, "Maximum number of log files")
	var logMaxAge = flag.Int("logMaxAge", 10, "Maximum log file age")
	flag.Parse()

	if len(os.Args) > 1 && os.Args[1] == "--help" {
        flag.PrintDefaults()
        return
    }

	logger.LogSetup(*logPath, *logMaxSize, *logMaxFiles, *logMaxAge, *logSeverity)
	if !tools.FileExists(*configFilePath) {
		logger.Logger.Fatalf("Configuration file '%s' not exist", *configFilePath)
	}

	datastore.SetConfigFilePath(*configFilePath)
	logConf := datastore.LogAppConfigStruct {
		Path: *logPath,
		MinSeverity: *logSeverity,
		MaxAge: *logMaxAge,
		MaxSize: *logMaxSize,
		MaxFiles: *logMaxFiles,
	}
	datastore.SetLogConfig(logConf)
	datastore.LoadConfig()
}

func setupAudit() {
	auditConf := datastore.GetConfig().Audit
	logger.AuditSetup(auditConf.Path, auditConf.MaxSize, auditConf.MaxFiles, auditConf.MaxAge, auditConf.MinSeverity)
}


func main() {
	setup()
	setupAudit()
	datastore.LoadSegmentConfigs()
	datastore.LoadSegmentPollingHosts()
	go grpcserver.StartGRPCServer()
	web.Webserver()
}