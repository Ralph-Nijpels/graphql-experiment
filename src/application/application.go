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

	"github.com/minio/minio-go"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Context describes the environment of the application including
// permanent connections and defaults
type Context struct {
	S3Client       *minio.Client
	DBClient       *mongo.Client
	DBContext      context.Context
	LogFolder      string
	MaxResults     int64
	CountriesURL   string
	RegionsURL     string
	AirportsURL    string
	RunwaysURL     string
	FrequenciesURL string
}

// Optionfile descibes the content of the options file
type sourceOptions struct {
	CountriesURL   string `json:"countries-url"`
	RegionsURL     string `json:"regions-url"`
	AirportsURL    string `json:"airports-url"`
	RunwaysURL     string `json:"runways-url"`
	FrequenciesURL string `json:"frequencies-url"`
}

type storageOptions struct {
	Server string `json:"server"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type optionFile struct {
	Source     sourceOptions  `json:"source"`
	Storage    storageOptions `json:"storage"`
	Database   string         `json:"database"`
	LogFolder  string         `json:"log-folder"`
	MaxResults int64          `json:"max-results"`
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

	// Connect to S3
	minioClient, err := minio.New(
		applicationOptions.Storage.Server,
		applicationOptions.Storage.Key,
		applicationOptions.Storage.Secret,
		false)
	if err != nil {
		log.Panicf("Unable to connect to S3-storage: %v\n", err)
	}

	// Check the csv bucket
	bucketFound, err := minioClient.BucketExists("csv")
	if err != nil {
		log.Panicf("Connection problem S3-storage: %v\n", err)
	}
	if !bucketFound {
		err = minioClient.MakeBucket("csv", "us-east-1")
		if err != nil {
			log.Panicf("Connection problem S3-storage: %v\n", err)
		}
	}

	// Check the log bucket
	bucketFound, err = minioClient.BucketExists("log")
	if err != nil {
		log.Panicf("Connection problem S3-storage: %v\n", err)
	}
	if !bucketFound {
		err = minioClient.MakeBucket("log", "us-east-1")
		if err != nil {
			log.Panicf("Connection problem S3-storage: %v\n", err)
		}
	}

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(applicationOptions.Database)
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
		S3Client:       minioClient,
		DBClient:       client,
		DBContext:      context.TODO(),
		LogFolder:      applicationOptions.LogFolder,
		MaxResults:     applicationOptions.MaxResults,
		CountriesURL:   applicationOptions.Source.CountriesURL,
		RegionsURL:     applicationOptions.Source.RegionsURL,
		AirportsURL:    applicationOptions.Source.AirportsURL,
		RunwaysURL:     applicationOptions.Source.RunwaysURL,
		FrequenciesURL: applicationOptions.Source.FrequenciesURL}

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
