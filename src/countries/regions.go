package countries

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"

	"../application"
	"../datatypes"
)

// regions implements the regions data

// Regions represents the connection to the database
type Regions struct {
	context *application.Context
	parent  *Countries
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
		context: countries.context,
		parent:  countries,
	}

	return &regions
}

// RetrieveFromURL downloads the file into the etc directory
func (regions *Regions) RetrieveFromURL() error {
	// Get the data
	resp, err := http.Get(regions.context.RegionsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(regions.context.RegionsCSV)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
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
		regions.context.DBContext,
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
	csvFile, err := os.Open(regions.context.RegionsCSV)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// Open the logfile
	_, err = regions.context.LogFile("regions")
	if err != nil {
		return err
	}

	regions.context.LogPrintln("Start Import")

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
		regions.context.LogError(err)
		line, err = reader.Read()
		lineNumber++
	}

	if err != io.EOF {
		return err
	}

	regions.context.LogPrintln("End Import")
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
