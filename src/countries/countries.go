package countries

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"../application"
	"../datatypes"
)

// Countries implements the datamodel for countries

// Countries is the representation of the Countries Collection in the database
type Countries struct {
	context    *application.Context
	collection *mongo.Collection
}

// Country is the external representation for an ISO-Country including both a bson (for mongo)
// and a json (for REST/GRAPHQL) representation
type Country struct {
	Country     primitive.ObjectID `bson:"_id" json:"-"`
	CountryCode string             `bson:"iso-country-code" json:"iso-country-code"`
	CountryName string             `bson:"country-name" json:"country-name"`
	Continent   string             `bson:"continent" json:"continent"`
	Wikipedia   string             `bson:"wikipedia" json:"wikipedia,omitempty"`
}

// insertCountry is the internal representation for an ISO-Country
// it ommits the Country(ID) from the structure to prevent race conditions while
// upserting new countries.
type insertCountry struct {
	CountryCode string `bson:"iso-country-code"`
	CountryName string `bson:"country-name"`
	Continent   string `bson:"continent"`
	Wikipedia   string `bson:"wikipedia"`
}

// NewCountries instantiates the connection to the database collection
func NewCountries(application *application.Context) *Countries {
	countries := Countries{context: application}

	// Country Collection
	countries.collection = application.DBClient.Database("flight-schedule").Collection("countries")
	countryIndex := mongo.IndexModel{Keys: bson.M{"iso-country-code": 1}}
	countries.collection.Indexes().CreateOne(application.DBContext, countryIndex)

	return &countries
}

// Below you find a number of support functions to simplify the code a little.
func (countries *Countries) dbClient() *mongo.Client {
	return countries.context.DBClient
}

func (countries *Countries) dbContext() context.Context {
	return countries.context.DBContext
}

func (countries *Countries) maxResults() int64 {
	return countries.context.MaxResults
}

func (countries *Countries) csvFile() (*os.File, error) {
	return os.Open(countries.context.CountriesCSV)
}

func (countries *Countries) logFile() (*os.File, error) {
	return countries.context.LogFile("countries")
}

func (countries *Countries) logPrint(s string) {
	countries.context.LogPrintln(s)
}

func (countries *Countries) logError(err error) {
	countries.context.LogError(err)
}

// GetByCountryCode retrieves a country based on a CountryCode.
func (countries *Countries) GetByCountryCode(countryCode string) (*Country, error) {
	var result Country

	countryCode, err := datatypes.ISOCountryCode(countryCode, false, false)
	if err != nil {
		return nil, err
	}

	err = countries.collection.FindOne(countries.dbContext(),
		bson.D{{Key: "iso-country-code", Value: countryCode}}).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("Not found")
	}

	return &result, nil
}

// GetList retrieves a list of countries [fromCountryCode .. untilCountryCode].
func (countries *Countries) GetList(fromCountryCode string, untilCountryCode string) ([]*Country, error) {
	var result []*Country
	var query = bson.D{{}}

	fromCountryCode, err := datatypes.ISOCountryCode(fromCountryCode, true, true)
	if err != nil {
		return nil, fmt.Errorf("fromCountry(%s): %v", fromCountryCode, err)
	}
	if len(fromCountryCode) != 0 {
		query = append(query, bson.E{Key: "iso-country-code",
			Value: bson.D{{Key: "$gte", Value: fromCountryCode}}})
	}

	untilCountryCode, err = datatypes.ISOCountryCode(untilCountryCode, true, true)
	if err != nil {
		return nil, fmt.Errorf("untilCountry(%s): %v", untilCountryCode, err)
	}
	if len(untilCountryCode) != 0 {
		query = append(query, bson.E{Key: "iso-country-code",
			Value: bson.D{{Key: "$lte", Value: untilCountryCode}}})
	}

	findOptions := options.Find()
	findOptions.SetLimit(countries.maxResults() + 1)

	cur, err := countries.collection.Find(countries.dbContext(), query, findOptions)
	if err != nil {
		return nil, fmt.Errorf("Not found")
	}

	for cur.Next(countries.dbContext()) {
		var country Country
		cur.Decode(&country)
		result = append(result, &country)
	}

	cur.Close(countries.dbContext())

	if int64(len(result)) > countries.maxResults() {
		return nil, fmt.Errorf("Too many results")
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("Not found")
	}

	return result, nil
}

func (countries *Countries) importCSVLine(line []string, lineNumber int) error {
	// Skipping empty lines
	if len(line) == 0 {
		return nil
	}

	// Check Country Code
	countryCode, err := datatypes.ISOCountryCode(line[1], false, false)
	if err != nil {
		return fmt.Errorf("Countries[%d].CountryCode(%s): %v", lineNumber, line[1], err)
	}

	// Build internal representation
	country := insertCountry{
		CountryCode: countryCode,
		CountryName: line[2],
		Continent:   line[3],
		Wikipedia:   line[4],
	}

	// Dump in mongo
	_, err = countries.collection.UpdateOne(countries.dbContext(),
		bson.D{{Key: "iso-country-code", Value: country.CountryCode}},
		bson.M{"$set": country},
		options.Update().SetUpsert(true))

	if err != nil {
		return fmt.Errorf("Countries[%d]: %v", lineNumber, err)
	}

	return nil
}

// ImportCSV imports a list of countries from a CSV-file
func (countries *Countries) ImportCSV() error {
	// Open the country.csv file
	csvFile, err := countries.csvFile()
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// Open the logfile
	_, err = countries.logFile()
	if err != nil {
		return err
	}

	countries.logPrint("Start Import")

	// Skip the headerline
	reader := csv.NewReader(bufio.NewReader(csvFile))
	line, err := reader.Read()
	if err != nil {
		return err
	}

	// Read the data
	// LineNumbers start at 1 and we've done the header (hence 2)
	lineNumber := 2
	line, err = reader.Read()
	for err == nil {
		err = countries.importCSVLine(line, lineNumber)
		countries.logError(err)
		line, err = reader.Read()
		lineNumber++
	}

	if err != io.EOF {
		return err
	}

	countries.logPrint("End Import")
	return nil
}
