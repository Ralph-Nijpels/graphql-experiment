package airports

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"../datatypes"
)

// Frequencies is the representation of the collection of frequencies as found in
// import / export tables
type Frequencies struct {
	parent *Airports
}

// Frequency is the external representation of a single frequency at an airport
type Frequency struct {
	FrequencyType string  `bson:"frequency-type" json:"frequency-type"`
	Description   string  `bson:"description" json:"description,omitempty"`
	Frequency     float64 `bson:"frequency-mhz" json:"frequency-mhz"`
}

// FrequencyView is a representation to help in graphql
type FrequencyView struct {
	AirportCode   string  `json:"icao-airport-code"`
	FrequencyType string  `json:"frequency-type"`
	Description   string  `json:"description,omitempty"`
	Frequency     float64 `json:"frequency-mhz"`
}

// NewFrequencies initializes the collection of frequencies
func (airports *Airports) NewFrequencies() *Frequencies {
	frequencies := Frequencies{
		parent: airports}
	return &frequencies
}

// Some support functions to clean up the code a little
func (frequencies *Frequencies) dbClient() *mongo.Client {
	return frequencies.parent.context.DBClient
}

func (frequencies *Frequencies) dbContext() context.Context {
	return frequencies.parent.context.DBContext
}

func (frequencies *Frequencies) maxResults() int64 {
	return frequencies.parent.context.MaxResults
}

func (frequencies *Frequencies) csvFile() (*os.File, error) {
	return os.Open(frequencies.parent.context.FrequenciesCSV)
}

func (frequencies *Frequencies) logFile() (*os.File, error) {
	return frequencies.parent.context.LogFile("frequencies")
}

func (frequencies *Frequencies) logPrint(s string) {
	frequencies.parent.context.LogPrintln(s)
}

func (frequencies *Frequencies) logError(err error) {
	frequencies.parent.context.LogError(err)
}

func (frequencies *Frequencies) importCSVLine(lineNumber int, line []string) error {
	// Skipping empty lines
	if len(line) == 0 {
		return nil
	}

	// Fetch the airport
	airport, err := frequencies.parent.GetByAirportCode(line[2])
	if err != nil {
		return fmt.Errorf("Frequencies[%d].AirportCode(%s): %v", lineNumber, line[2], err)
	}

	frequencyMhz, err := datatypes.Frequency(line[5], false)
	// build internal representation
	frequency := Frequency{
		FrequencyType: line[3],
		Description:   line[4],
		Frequency:     frequencyMhz}

	// replace or add frequency...
	found := false
	for i := range airport.Frequencies {
		if airport.Frequencies[i].FrequencyType == frequency.FrequencyType {
			airport.Frequencies[i] = &frequency
			found = true
			break
		}
	}
	if !found {
		airport.Frequencies = append(airport.Frequencies, &frequency)
	}

	// Dump in mongo
	_, err = frequencies.parent.collection.UpdateOne(
		frequencies.dbContext(),
		bson.D{{Key: "icao-airport-code", Value: airport.AirportCode}},
		bson.M{"$set": airport})

	if err != nil {
		return fmt.Errorf("Frequencies[%d]: %v", lineNumber, err)
	}

	return nil
}

// ImportCSV imports runways into the airport collection
func (frequencies *Frequencies) ImportCSV() error {

	// Open the frequencies.csv file
	csvFile, err := frequencies.csvFile()
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// Skip the headerline
	reader := csv.NewReader(bufio.NewReader(csvFile))
	line, err := reader.Read()
	if err != nil {
		return err
	}

	// open the logfile
	_, err = frequencies.logFile()
	if err != nil {
		return err
	}
	frequencies.logPrint("Start Import")

	// Read the data
	// lineNumbers start at 1 and we've done the header
	lineNumber := 2
	line, err = reader.Read()
	for err == nil {
		err = frequencies.importCSVLine(lineNumber, line)
		frequencies.logError(err)
		line, err = reader.Read()
		lineNumber++
	}

	if err != io.EOF {
		return err
	}

	frequencies.logPrint("End Import")

	return nil
}

// AsFrequencyView translates the Frequency into a FrequencyView
func AsFrequencyView(airport *Airport, frequency *Frequency) *FrequencyView {
	var result FrequencyView

	result.AirportCode = airport.AirportCode
	result.FrequencyType = frequency.FrequencyType
	result.Description = frequency.Description
	result.Frequency = frequency.Frequency

	return &result
}