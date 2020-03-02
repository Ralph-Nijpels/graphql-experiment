package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Context describes the environment of the application including
// permanent connections and defaults
type Context struct {
	DBClient       *mongo.Client
	DBContext      context.Context
	LogFolder      string
	MaxResults     int64
	CountriesCSV   string
	RegionsCSV     string
	AirportsCSV    string
	RunwaysCSV     string
	FrequenciesCSV string
}

// Optionfile descibes the content of the options file
type optionFile struct {
	Database       string `json:"database"`
	LogFolder      string `json:"log-folder"`
	CountriesCSV   string `json:"countries-csv"`
	RegionsCSV     string `json:"regions-csv"`
	AirportsCSV    string `json:"airports-csv"`
	RunwaysCSV     string `json:"runways-csv"`
	FrequenciesCSV string `json:"frequencies-csv"`
	MaxResults     int64  `json:"max-results"`
}

func readOptions() (*optionFile, error) {
	var options optionFile

	optionFile, err := os.Open("options.json")
	if err != nil {
		return nil, err
	}

	defer optionFile.Close()
	decoder := json.NewDecoder(optionFile)
	err = decoder.Decode(&options)
	if err != nil {
		return nil, err
	}

	return &options, nil
}

// GetContext reads the application options and initializes permanent connections and defaults
func GetContext() (*Context, error) {

	applicationOptions, err := readOptions()
	if err != nil {
		return nil, err
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(applicationOptions.Database)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	// Compose result
	context := Context{
		DBClient:       client,
		DBContext:      context.TODO(),
		LogFolder:      applicationOptions.LogFolder,
		MaxResults:     applicationOptions.MaxResults,
		CountriesCSV:   applicationOptions.CountriesCSV,
		RegionsCSV:     applicationOptions.RegionsCSV,
		AirportsCSV:    applicationOptions.AirportsCSV,
		RunwaysCSV:     applicationOptions.RunwaysCSV,
		FrequenciesCSV: applicationOptions.FrequenciesCSV}

	return &context, nil

}

func (context *Context) openLogFolder() (*os.File, error) {
	logDir, err := os.Open(context.LogFolder)
	if err != nil {
		return nil, err
	}

	logDirStat, err := logDir.Stat()
	if err != nil {
		return nil, err
	}
	if !logDirStat.IsDir() {
		return nil, fmt.Errorf("Not a directory")
	}

	return logDir, nil
}

func (context *Context) lastLogNumber(logDir *os.File, topic string) (int64, error) {
	logFileNames, err := logDir.Readdirnames(-1)
	if err != nil {
		return 0, err
	}

	highestLogNumber := int64(0)
	for _, logFileName := range logFileNames {
		if strings.HasPrefix(logFileName, topic) && strings.HasSuffix(logFileName, ".log") {
			logFileParts := strings.Split(strings.TrimSuffix(logFileName, ".log"), "-")
			if len(logFileParts) == 3 {
				logFileNumber, _ := strconv.ParseInt(logFileParts[2], 10, 64)
				if logFileNumber > highestLogNumber {
					highestLogNumber = logFileNumber
				}
			}
		}
	}

	return highestLogNumber, nil
}

// LogFile creates a new logfile for the given topic in the logfolder
func (context *Context) LogFile(topic string) (*os.File, error) {

	logDate := time.Now().Format("20060102")

	logDir, err := context.openLogFolder()
	if err != nil {
		log.Panicf("Could not open log folder: %v\n", err)
	}

	logFileNumber, err := context.lastLogNumber(logDir, topic)
	if err != nil {
		log.Panicf("Could not open log folder: %v\n", err)
	}

	logFile, err := os.Create(fmt.Sprintf("%s/%s-%s-%04d.log", context.LogFolder, topic, logDate, logFileNumber+1))
	if err != nil {
		return nil, err
	}

	log.SetOutput(logFile)
	return logFile, err
}

// LogPrintln inserts a message in the logfile
func (context *Context) LogPrintln(s string) {
	if len(s) != 0 {
		log.Println(s)
	}
}

// LogError inserts an error in the logfile if there is one
func (context *Context) LogError(err error) {
	if err != nil {
		log.Println(err)
	}
}
