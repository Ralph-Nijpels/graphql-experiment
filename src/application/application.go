package application

import (
	"context"
	"encoding/json"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Context describes the environment of the application including
// permanent connections and defaults
type Context struct {
	DBClient       *mongo.Client
	DBContext      context.Context
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
		MaxResults:     applicationOptions.MaxResults,
		CountriesCSV:   applicationOptions.CountriesCSV,
		RegionsCSV:     applicationOptions.RegionsCSV,
		AirportsCSV:    applicationOptions.AirportsCSV,
		RunwaysCSV:     applicationOptions.RunwaysCSV,
		FrequenciesCSV: applicationOptions.FrequenciesCSV}

	return &context, nil

}
