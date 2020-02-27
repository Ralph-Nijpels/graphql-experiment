package regions

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"../application"
	"../countries"
	"../datatypes"
)

// regions implements the regions data

// Regions represents the connection to the database
type Regions struct {
	dbClient   *mongo.Client
	dbContext  context.Context
	collection *mongo.Collection
	countries  *countries.Countries
	csvFile    string
	maxResults int64
}

// Region is the external representation for an ISO-Region including both a bson (for mongo)
// and a json (for REST/GRAPHQL) representation
type Region struct {
	Region      primitive.ObjectID `bson:"_id" json:"-"`
	RegionCode  string             `bson:"iso-region-code" json:"iso-region-code"`
	RegionName  string             `bson:"region-name" json:"region-name"`
	Country     primitive.ObjectID `bson:"country-id" json:"-"`
	CountryCode string             `bson:"iso-country-code" json:"iso-country-code"`
	Wikipedia   string             `bson:"wikipedia" json:"wikipedia,omitempty"`
}

// insertRegion is the internal representation for an ISO-Region
// it ommits the Region(ID) from the structure to prevent race conditions while
// upserting new regions.
type insertRegion struct {
	RegionCode  string             `bson:"iso-region-code"`
	RegionName  string             `bson:"region-name"`
	Country     primitive.ObjectID `bson:"country-id"`
	CountryCode string             `bson:"iso-country-code"`
	Wikipedia   string             `bson:"wikipedia"`
}

// NewRegions establishes the connection to the database
func NewRegions(application *application.Context, countries *countries.Countries) *Regions {
	regions := Regions{
		dbClient:   application.DBClient,
		dbContext:  application.DBContext,
		csvFile:    application.RegionsCSV,
		countries:  countries,
		maxResults: application.MaxResults}

	// Region Collection
	regions.collection = application.DBClient.Database("flight-schedule").Collection("regions")
	regionIndex := mongo.IndexModel{Keys: bson.M{"iso-region-code": 1}}
	regions.collection.Indexes().CreateOne(application.DBContext, regionIndex)

	return &regions
}

// GetByRegionCode retrieves a region from the database based on a RegionCode
func (regions *Regions) GetByRegionCode(regionCode string) (*Region, error) {
	var result Region

	regionCode, err := datatypes.ISORegionCode(regionCode, false, false)
	if err != nil {
		return nil, err
	}

	err = regions.collection.FindOne(regions.dbContext,
		bson.D{{Key: "iso-region-code", Value: regionCode}}).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("Not found")
	}

	return &result, nil
}

// GetList retrieves a list of regions from the database based on the provided filter
func (regions *Regions) GetList(countryCode string, fromRegionCode string, untilRegionCode string) ([]*Region, error) {
	var result []*Region
	var query = bson.D{{}}

	parameter, err := datatypes.ISOCountryCode(countryCode, false, true)
	if err != nil {
		return nil, fmt.Errorf("GetList.CountryCode(%s): %v", countryCode, err)
	}
	if len(parameter) != 0 {
		query = append(query, bson.E{Key: "iso-country-code", Value: parameter})
	}

	parameter, err = datatypes.ISORegionCode(fromRegionCode, true, true)
	if err != nil {
		return nil, fmt.Errorf("GetList.fromRegionCode(%s): %v", fromRegionCode, err)
	}
	if len(parameter) != 0 {
		query = append(query, bson.E{Key: "iso-region-code", Value: bson.D{{Key: "$gte", Value: parameter}}})
	}

	parameter, err = datatypes.ISORegionCode(untilRegionCode, true, true)
	if err != nil {
		return nil, fmt.Errorf("GetList.untilRegionCode(%s): %v", untilRegionCode, err)
	}
	if len(parameter) != 0 {
		query = append(query, bson.E{Key: "iso-region-code", Value: bson.D{{Key: "$lte", Value: parameter}}})
	}

	findOptions := options.Find()
	findOptions.SetLimit(regions.maxResults + 1)

	cur, err := regions.collection.Find(regions.dbContext, query, findOptions)
	if err != nil {
		return nil, fmt.Errorf("Not found")
	}

	for cur.Next(regions.dbContext) {
		var region Region
		cur.Decode(&region)
		result = append(result, &region)
	}

	cur.Close(regions.dbContext)

	if int64(len(result)) > regions.maxResults {
		return nil, fmt.Errorf("Too many results")
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("Not found")
	}

	return result, nil
}

func (regions *Regions) importCSVLine(line []string, lineNumber int) error {
	// Skipping empty lines
	if len(line) == 0 {
		return nil
	}

	// Check Region Code
	regionCode, err := datatypes.ISORegionCode(line[1], false, false)
	if err != nil {
		return fmt.Errorf("Regions[%d].RegionCode(%s): %v", lineNumber, line[1], err)
	}

	// Check CountryID
	country, err := regions.countries.GetByCountryCode(line[5])
	if err != nil {
		return fmt.Errorf("Regions[%d].CountryCode(%s): %v", lineNumber, line[5], err)
	}

	// Build internal representation
	region := insertRegion{
		RegionCode:  regionCode,
		RegionName:  line[3],
		Country:     country.Country,
		CountryCode: country.CountryCode,
		Wikipedia:   line[6],
	}

	// Dump in mongo
	_, err = regions.collection.UpdateOne(
		regions.dbContext,
		bson.D{{Key: "iso-region-code", Value: region.RegionCode}},
		bson.M{"$set": region},
		options.Update().SetUpsert(true))

	if err != nil {
		return fmt.Errorf("Regions[%d]: %v", lineNumber, err)
	}

	return nil
}

// ImportCSV initializes the database from a CSV-file
func (regions *Regions) ImportCSV() error {
	// Open the regions.csv file
	csvFile, err := os.Open(regions.csvFile)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	reader := csv.NewReader(bufio.NewReader(csvFile))

	// Skip the headerline
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
		// TODO: Add logging
		line, err = reader.Read()
		lineNumber++
	}

	if err != io.EOF {
		return err
	}

	return nil
}
