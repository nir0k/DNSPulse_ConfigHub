package datastore

import (
	"DNSPulse_ConfigHub/pkg/logger"
	"DNSPulse_ConfigHub/pkg/tools"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/gocarina/gocsv"
)


type Csv struct {
    Server                  string `csv:"server" json:"Server"`
    IPAddress               string `csv:"server_ip" json:"IPAddress"`
    Domain                  string `csv:"domain" json:"Domain"`
    Location                string `csv:"location" json:"Location"`
    Site                    string `csv:"site" json:"Site"`
    ServerSecurityZone      string `csv:"server_security_zone" json:"ServerSecurityZone"`
    Prefix                  string `csv:"prefix" json:"Prefix"`
    Protocol                string `csv:"protocol" json:"Protocol"`
    Zonename                string `csv:"zonename" json:"Zonename"`
    QueryCount              string `csv:"query_count_rps" json:"QueryCount"`
    ZonenameWithRecursion   string `csv:"zonename_with_recursion" json:"ZonenameWithRecursion"`
    QueryCountWithRecursion string `csv:"query_count_with_recursion_rps" json:"QueryCountWithRecursion"`
    ServiceMode             string `csv:"service_mode" json:"ServiceMode"`
}

type pollingCSVMap map[string][]Csv

var (
    pollingCSV  pollingCSVMap
	pollingHostsMutex sync.RWMutex
)

func init() {
    pollingCSV = make(map[string][]Csv)
}

func LoadSegmentPollingHosts() error {
	segments := GetConfig().SegmentConfigs
	configs := *GetSegmentsConfig()
	for _, s := range segments {
		readResolversFromCSV(s, configs[s.Name].Polling)
	}
	return nil
}

func readResolversFromCSV(segment SegmentConfigsStruct, conf PollingConfigStruct) error {
    var (
		delimeter rune
	)
	if len(conf.Delimeter) > 0 {
		delimeter = rune(conf.Delimeter[0])
	} else {
		return errors.New("string is empty, cannot parse to rune")
	}

    fileHash, err := tools.CalculateHash(string(conf.Path))
    if err != nil {
        logger.Logger.Errorf("Error Calculate hash to file '%s' err: %v\n", configFile, err)
        return err
    }

    if conf.Hash == fileHash {
        logger.Logger.Debug("CSV file has not been changed")
        return nil
    }
    logger.Logger.Info("CSV file has been changed")

	resolversFromCsv := []Csv{}
    clientsFile, err := os.OpenFile(conf.Path, os.O_RDWR, os.ModePerm)
	if err != nil {
        return err
	}
	defer clientsFile.Close()
    
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
        r := csv.NewReader(in)
        r.LazyQuotes = true
        r.Comma = delimeter
        return r
    })
	if err := gocsv.UnmarshalFile(clientsFile, &resolversFromCsv); err != nil {
        return fmt.Errorf("error Unmarshal file %s : %v", conf.Path, err)
	}

	uniqueChecker := make(map[string]bool)
    var duplicates []string
    for _, record := range resolversFromCsv {
        key := record.Server + "-" + record.IPAddress + "-" + record.Domain
        if _, exists := uniqueChecker[key]; exists {
            duplicates = append(duplicates, key)
            continue
        }
        uniqueChecker[key] = true
    }

    if len(duplicates) > 0 {
        return fmt.Errorf("duplicates found in CSV file: %s", strings.Join(duplicates, ", "))
    }
    
	pollingHostsMutex.Lock()
    defer pollingHostsMutex.Unlock()
    pollingCSV[segment.Name] = resolversFromCsv
	UpdatePollingHash(segment.Name, fileHash)
    return nil
}

func GetPollingHostsBySegment(segmentName string) ([]Csv, bool) {
    pollingHostsMutex.RLock()
    defer pollingHostsMutex.RUnlock()
    hosts, ok := pollingCSV[segmentName]
    return hosts, ok
}

func GetPollingHosts() *pollingCSVMap {
    pollingHostsMutex.RLock()
    defer pollingHostsMutex.RUnlock()
    fmt.Printf("pollingCSV: %v\n", pollingCSV)
    return &pollingCSV
}

func UpdatePollingHostsBySegment(segmentName string, hosts []Csv) ([]Csv, error) {
    pollingHostsMutex.RLock()
    defer pollingHostsMutex.RUnlock()
    pollingCSV[segmentName] = hosts
    err := WriteResolversToCSV(segmentName, GetSegmentsPollingConfigbySegment(segmentName))
    if err != nil {
        return []Csv{}, fmt.Errorf("error writing updated polling host into file for %s, err: %v", segmentName, err)
    }
    return pollingCSV[segmentName], nil
}

func WriteResolversToCSV(segmentName string, conf PollingConfigStruct) error {
    pollingHostsMutex.RLock()
    hosts, ok := pollingCSV[segmentName]
    pollingHostsMutex.RUnlock()

    if !ok {
        return fmt.Errorf("no hosts found for segment %s", segmentName)
    }

    clientsFile, err := os.OpenFile(conf.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
    if err != nil {
        return fmt.Errorf("error opening file %s for writing: %v", conf.Path, err)
    }
    defer clientsFile.Close()

    gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
        writer := csv.NewWriter(out)
        if len(conf.Delimeter) > 0 {
            writer.Comma = rune(conf.Delimeter[0])
        }
        return gocsv.NewSafeCSVWriter(writer)
    })

    if err := gocsv.MarshalFile(&hosts, clientsFile); err != nil {
        return fmt.Errorf("error marshalling data to file %s : %v", conf.Path, err)
    }

    fileHash, err := tools.CalculateHash(string(conf.Path))
    if err != nil {
        logger.Logger.Errorf("Error Calculate hash to file '%s' err: %v\n", conf.Path, err)
        return err
    }
    UpdatePollingHash(segmentName, fileHash)

    return nil
}
