package countries

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/bson"

	"../datatypes"
)

// regions implements the regions data

// Regions represents the connection to the database
type Regions struct {
	parent *Countries
}

// Region is the external representation for an ISO-Region including both a bson (for mongo)
// and a json (for REST/GRAPHQL) representation
type Region struct {
	RegionCode string `bson:"iso-region-code" json:"iso-region-code"`
	RegionName string `bson:"region-name" json:"region-name"`
	Wikipedia  string `bson:"wikipedia" json:"wikipedia,omitempty"`
}

// RegionView is the external representation 'flattened' so it is easier to handle in
// graphql
type RegionView struct {
	CountryCode string `json:"iso-country-code"`
	RegionCode  string `json:"iso-region-code"`
	RegionName  string `json:"region-name"`
	Wikipedia   string `json:"wikipedia,omitempty"`
}

// NewRegions establishes the connection to the database
func (countries *Countries) NewRegions() *Regions {
	regions := Regions{
		parent: countries}

	return &regions
}

// Below some simple support functions to help ease the code
func (regions *Regions) dbClient() *mongo.Client {
	return regions.parent.context.DBClient
}

func (regions *Regions) dbContext() context.Context {
	return regions.parent.context.DBContext
}

func (regions *Regions) csvFile() (*os.File, error) {
	return os.Open(regions.parent.context.RegionsCSV)
}

func (regions *Regions) maxResults() int64 {
	return regions.parent.context.MaxResults
}

func (regions *Regions) logFile() (*os.File, error) {
	return regions.parent.context.LogFile("regions")
}

func (regions *Regions) logPrint(s string) {
	regions.parent.context.LogPrintln(s)
}

func (regions *Regions) logError(err error) {
	regions.parent.context.LogError(err)
}

func (regions *Regions) importCSVLine(line []string, lineNumber int) error {
	// Skipping empty lines
	if len(line) == 0 {
		return nil
	}

	// Check Region Code
	regionCode, err := datatypes.ISORegionCode(line[2], false, false)
	if err != nil {
		return fmt.Errorf("Regions[%d].RegionCode(%s): %v", lineNumber, line[1], err)
	}

	// Check CountryID
	country, err := regions.parent.GetByCountryCode(line[5])
	if err != nil {
		return fmt.Errorf("Regions[%d].CountryCode(%s): %v", lineNumber, line[5], err)
	}

	// Build internal representation
	region := Region{
		RegionCode: regionCode,
		RegionName: line[3],
		Wikipedia:  line[6]}

	// replace or add region...
	found := false
	for i := range country.Regions {
		if country.Regions[i].RegionCode == region.RegionCode {
			country.Regions[i] = &region
			found = true
			break
		}
	}
	if !found {
		country.Regions = append(country.Regions, &region)
	}

	// Dump in mongo
	_, err = regions.parent.collection.UpdateOne(
		regions.dbContext(),
		bson.D{{Key: "iso-country-code", Value: country.CountryCode}},
		bson.M{"$set": country})

	if err != nil {
		return fmt.Errorf("Regions[%d]: %v", lineNumber, err)
	}

	return nil
}

// ImportCSV initializes the database from a CSV-file
func (regions *Regions) ImportCSV() error {
	// Open the regions.csv file
	csvFile, err := regions.csvFile()
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// Open the logfile
	_, err = regions.logFile()
	if err != nil {
		return err
	}

	regions.logPrint("Start Import")

	// Skip the headerline
	reader := csv.NewReader(bufio.NewReader(csvFile))
	line, err := reader.Read()
	if err != nil {
		return err
	}

	// Read the data
	// line Numbers start at 1 and we've done the header, hence 2
	lineNumber := 2
	line, err = reader.Read()
	for err == nil {
		err = regions.importCSVLine(line, lineNumber)
		regions.logError(err)
		line, err = reader.Read()
		lineNumber++
	}

	if err != io.EOF {
		return err
	}

	regions.logPrint("End Import")
	return nil
}

// Some support function on the region itself

// AsRegionView translates the internal view to the view more suitable for graphql
func AsRegionView(country *Country, region *Region) *RegionView {
	var result RegionView

	result.CountryCode = country.CountryCode
	result.RegionCode = region.RegionCode
	result.RegionName = region.RegionName
	result.Wikipedia = region.Wikipedia

	return &result
}