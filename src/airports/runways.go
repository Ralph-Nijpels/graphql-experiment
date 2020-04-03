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

// Runways is the representation of the collection of runways. The runways are implemented
// as an element of the Airport, but it has some methods of it's own.
type Runways struct {
	parent *Airports
}

// Runway is the database representation of a runway belonging to an Airport
type Runway struct {
	Length  int         `bson:"length" json:"length"`
	Width   int         `bson:"width" json:"width"`
	Surface string      `bson:"surface" json:"surface"`
	Lighted bool        `bson:"lighted" json:"lighted"`
	Closed  bool        `bson:"closed" json:"closed"`
	LowEnd  *RunwaySide `bson:"low-end" json:"low-end"`
	HighEnd *RunwaySide `bson:"high-end" json:"high-end,omitempty"`
}

// RunwaySide expresses the two sides that a Runway usually has (except heliports)
type RunwaySide struct {
	RunwayCode string  `bson:"runway-code" json:"runway-code"`
	Latitude   float64 `bson:"latitude" json:"latitude,omitempty"`
	Longitude  float64 `bson:"longitude" json:"longitude,omitempty"`
	Elevation  int     `bson:"elevation" json:"elevation,omitempty"`
	Heading    int     `bson:"heading" json:"heading,omitempty"`
	Threshold  int     `bson:"threshold" json:"threshold,omitempty"`
}

// RunwayView expresses a model where the runway is flattened and made independent
// from the airport to be more suitable for graphql.
type RunwayView struct {
	AirportCode   string  `json:"icao-airport-code"`
	RunwayCode    string  `json:"runway-code"`
	AltRunwayCode string  `json:"alt-runway-code"`
	Latitude      float64 `json:"latitude,omitempty"`
	Longitude     float64 `json:"longitude,omitempty"`
	Elevation     int     `json:"elevation,omitempty"`
	Heading       int     `json:"heading,omitempty"`
	Threshold     int     `json:"threshold,omitempty"`
	Length        int     `json:"length"`
	Width         int     `json:"width"`
	Surface       string  `json:"surface"`
	Lighted       bool    `json:"lighted"`
	Closed        bool    `json:"closed"`
}

// NewRunways initializes the collection of runways
func (airports *Airports) NewRunways() *Runways {
	runways := Runways{
		parent: airports}
	return &runways
}

// Some support functions to clean up the code a little
func (runways *Runways) dbClient() *mongo.Client {
	return runways.parent.context.DBClient
}

func (runways *Runways) dbContext() context.Context {
	return runways.parent.context.DBContext
}

func (runways *Runways) maxResults() int64 {
	return runways.parent.context.MaxResults
}

func (runways *Runways) csvFile() (*os.File, error) {
	return os.Open(runways.parent.context.RunwaysCSV)
}

func (runways *Runways) logFile() (*os.File, error) {
	return runways.parent.context.LogFile("runways")
}

func (runways *Runways) logPrint(s string) {
	runways.parent.context.LogPrintln(s)
}

func (runways *Runways) logError(err error) {
	runways.parent.context.LogError(err)
}

func (runways *Runways) importCSVLine(lineNumber int, line []string) error {
	// Skipping empty lines
	if len(line) == 0 {
		return nil
	}

	// Fetch the airport
	airport, err := runways.parent.GetByAirportCode(line[2])
	if err != nil {
		return fmt.Errorf("Runway[%d].AirportCode(%s): %v", lineNumber, line[2], err)
	}

	runwayLength, err := datatypes.RunwayLength(line[3], false)
	if err != nil {
		return fmt.Errorf("Runway[%d].Length(%s): %v", lineNumber, line[3], err)
	}

	runwayWidth, err := datatypes.RunwayWidth(line[4], true)
	if err != nil {
		return fmt.Errorf("Runway[%d].Width(%s): %v", lineNumber, line[4], err)
	}

	runwayLighted, err := datatypes.RunwayLighted(line[6], true)
	if err != nil {
		return fmt.Errorf("Runway[%d].Lighted(%s): %v", lineNumber, line[6], err)
	}

	runwayClosed, err := datatypes.RunwayClosed(line[7], true)
	if err != nil {
		return fmt.Errorf("Runway[%d].Closed(%s): %v", lineNumber, line[7], err)
	}

	// build internal representation
	runway := Runway{
		Length:  runwayLength,
		Width:   runwayWidth,
		Surface: line[5],
		Lighted: runwayLighted,
		Closed:  runwayClosed}

	// Check for any low-end identifier
	if len(line[8]) == 0 {
		return fmt.Errorf("Runway[%d].Lowend.Code(%s): Missing", lineNumber, line[8])
	}

	lowendCode, err := datatypes.RunwayCode(line[8], false, false)
	if err != nil {
		return fmt.Errorf("Runway[%d].Lowend.Code(%s): %v", lineNumber, line[8], err)
	}

	lowendLatitude, err := datatypes.Latitude(line[9], true)
	if err != nil {
		return fmt.Errorf("Runway[%d].Lowend.Latitude(%s): %v", lineNumber, line[9], err)
	}

	lowendLongitude, err := datatypes.Longitude(line[10], true)
	if err != nil {
		return fmt.Errorf("Runway[%d].Lowend.Longitude(%s): %v", lineNumber, line[10], err)
	}

	lowendElevation, err := datatypes.Elevation(line[11], true)
	if err != nil {
		return fmt.Errorf("Runway[%d].Lowend.Elevation(%s): %v", lineNumber, line[11], err)
	}

	lowendHeading, err := datatypes.RunwayHeading(line[12], true)
	if err != nil {
		return fmt.Errorf("Runway[%d].Lowend.Heading(%s): %v", lineNumber, line[12], err)
	}

	lowendThreshold, err := datatypes.RunwayThreshold(line[13], true)
	if err != nil {
		return fmt.Errorf("Runway[%d].Lowend.Threshold(%s): %v", lineNumber, line[13], err)
	}

	runway.LowEnd = &RunwaySide{
		RunwayCode: lowendCode,
		Latitude:   lowendLatitude,
		Longitude:  lowendLongitude,
		Elevation:  lowendElevation,
		Heading:    lowendHeading,
		Threshold:  lowendThreshold}

	if len(line[14]) > 0 {
		highendCode, err := datatypes.RunwayCode(line[14], false, false)
		if err != nil {
			return fmt.Errorf("Runway[%d].Highend.Code(%s): %v", lineNumber, line[14], err)
		}

		highendLatitude, err := datatypes.Latitude(line[15], true)
		if err != nil {
			return fmt.Errorf("Runway[%d].Highend.Latitude(%s): %v", lineNumber, line[15], err)
		}

		highendLongitude, err := datatypes.Longitude(line[16], true)
		if err != nil {
			return fmt.Errorf("Runway[%d].Highend.Longitude(%s): %v", lineNumber, line[16], err)
		}

		highendElevation, err := datatypes.Elevation(line[17], true)
		if err != nil {
			return fmt.Errorf("Runway[%d].Highend.Elevation(%s): %v", lineNumber, line[17], err)
		}

		highendHeading, err := datatypes.RunwayHeading(line[18], true)
		if err != nil {
			return fmt.Errorf("Runway[%d].Highend.Heading(%s): %v", lineNumber, line[18], err)
		}

		highendThreshold, err := datatypes.RunwayThreshold(line[19], true)
		if err != nil {
			return fmt.Errorf("Runway[%d].Highend.Threshold(%s): %v", lineNumber, line[19], err)
		}

		if highendCode != runway.LowEnd.RunwayCode {
			runway.HighEnd = &RunwaySide{
				RunwayCode: highendCode,
				Latitude:   highendLatitude,
				Longitude:  highendLongitude,
				Elevation:  highendElevation,
				Heading:    highendHeading,
				Threshold:  highendThreshold}
		}
	}

	// replace or add runway...
	found := false
	for i := range airport.Runways {
		if airport.Runways[i].LowEnd.RunwayCode == runway.LowEnd.RunwayCode {
			airport.Runways[i] = &runway
			found = true
			break
		}
	}
	if !found {
		airport.Runways = append(airport.Runways, &runway)
	}

	// Dump in mongo
	_, err = runways.parent.collection.UpdateOne(
		runways.dbContext(),
		bson.D{{Key: "icao-airport-code", Value: airport.AirportCode}},
		bson.M{"$set": airport})

	if err != nil {
		return fmt.Errorf("Runways[%d]: %v", lineNumber, err)
	}

	return nil
}

// ImportCSV imports runways into the airport collection
func (runways *Runways) ImportCSV() error {

	// Open the airports.csv file
	csvFile, err := runways.csvFile()
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
	_, err = runways.logFile()
	if err != nil {
		return err
	}
	runways.logPrint("Start Import")

	// Read the data
	// lineNumbers start at 1 and we've done the header
	lineNumber := 2
	line, err = reader.Read()
	for err == nil {
		err = runways.importCSVLine(lineNumber, line)
		runways.logError(err)
		line, err = reader.Read()
		lineNumber++
	}

	if err != io.EOF {
		return err
	}

	runways.logPrint("End Import")

	return nil
}

// Some functions on the Runway itself

// AsRunwayView changes the runway into a more flattened model
func AsRunwayView(airport *Airport, runway *Runway) []*RunwayView {
	var result []*RunwayView

	if len(runway.LowEnd.RunwayCode) > 0 {
		var runwayView RunwayView

		runwayView.AirportCode = airport.AirportCode
		runwayView.RunwayCode = runway.LowEnd.RunwayCode
		runwayView.AltRunwayCode = runway.HighEnd.RunwayCode
		runwayView.Latitude = runway.LowEnd.Latitude
		runwayView.Longitude = runway.LowEnd.Longitude
		runwayView.Elevation = runway.LowEnd.Elevation
		runwayView.Heading = runway.LowEnd.Heading
		runwayView.Threshold = runway.LowEnd.Threshold
		runwayView.Length = runway.Length
		runwayView.Width = runway.Width
		runwayView.Surface = runway.Surface
		runwayView.Lighted = runway.Lighted
		runwayView.Closed = runway.Closed

		result = append(result, &runwayView)
	}

	if len(runway.HighEnd.RunwayCode) > 0 {
		var runwayView RunwayView

		runwayView.AirportCode = airport.AirportCode
		runwayView.RunwayCode = runway.HighEnd.RunwayCode
		runwayView.AltRunwayCode = runway.LowEnd.RunwayCode
		runwayView.Latitude = runway.HighEnd.Latitude
		runwayView.Longitude = runway.HighEnd.Longitude
		runwayView.Elevation = runway.HighEnd.Elevation
		runwayView.Heading = runway.HighEnd.Heading
		runwayView.Threshold = runway.HighEnd.Threshold
		runwayView.Length = runway.Length
		runwayView.Width = runway.Width
		runwayView.Surface = runway.Surface
		runwayView.Lighted = runway.Lighted
		runwayView.Closed = runway.Closed

		result = append(result, &runwayView)
	}

	return result
}
