package airports

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"../application"
	"../countries"
	"../datatypes"
)

// airports implements the Airport Datamodel

// Airports is the representation of the collection of Airports in the geography database
type Airports struct {
	context    *application.Context
	collection *mongo.Collection
	countries  *countries.Countries
	regions    *countries.Regions
}

// Airport is the external representation for an ICAO-airport including both a bson (for mongo)
// and a json (for REST/GRAPHQL) representation
type Airport struct {
	Airport      primitive.ObjectID `bson:"_id" json:"-"`
	AirportCode  string             `bson:"icao-airport-code" json:"icao-airport-code"`
	AirportName  string             `bson:"airport-name" json:"airport-name"`
	AirportType  string             `bson:"airport-type" json:"airport-type"`
	Latitude     float64            `bson:"latitude" json:"latitude"`
	Longitude    float64            `bson:"longitude" json:"longitude"`
	Elevation    float64            `bson:"elevation" json:"elevation,omitempty"`
	Country      primitive.ObjectID `bson:"country-id" json:"-"`
	CountryCode  string             `bson:"iso-country-code" json:"iso-country-code"`
	RegionCode   string             `bson:"iso-region-code" json:"iso-region-code,omitempty"`
	Municipality string             `bson:"municipality" json:"municipality,omitempty"`
	IATA         string             `bson:"iata-airport-code" json:"iata-airport-code,omitempty"`
	Website      string             `bson:"website" json:"website,omitempty"`
	Wikipedia    string             `bson:"wikipedia" json:"wikipedia,omitempty"`
	Runways      []*Runway          `bson:"runways" json:"runways,omitempty"`
}

// insertAirport is the internal representation for an ICAO-airport used for importing
// airports from CSV. It ommits the Airport(ID) from the structure  to prevent race conditions
// while upserting new airports. It ommits Runways and Frequencies from the structure because
// these are imported seperately
type insertAirport struct {
	AirportCode  string             `bson:"icao-airport-code"`
	AirportName  string             `bson:"airport-name"`
	AirportType  string             `bson:"airport-type"`
	Latitude     float64            `bson:"latitude"`
	Longitude    float64            `bson:"longitude"`
	Elevation    int                `bson:"elevation"`
	Country      primitive.ObjectID `bson:"country-id"`
	CountryCode  string             `bson:"iso-country-code"`
	RegionCode   string             `bson:"iso-region-code"`
	Municipality string             `bson:"municipality"`
	IATA         string             `bson:"iata-airport-code"`
	Website      string             `bson:"website"`
	Wikipedia    string             `bson:"wikipedia"`
}

// NewAirports sets up the connection to the database
func NewAirports(application *application.Context, countries *countries.Countries, regions *countries.Regions) *Airports {
	airports := Airports{
		context:   application,
		countries: countries,
		regions:   regions}

	// Setup the Airport Collection
	airports.collection = application.DBClient.Database("flight-schedule").Collection("airports")
	airportIndex1 := mongo.IndexModel{Keys: bson.M{"icao-airport-code": 1}}
	airports.collection.Indexes().CreateOne(application.DBContext, airportIndex1)
	airportIndex2 := mongo.IndexModel{Keys: bson.M{"iata-airport-code": 1}}
	airports.collection.Indexes().CreateOne(application.DBContext, airportIndex2)

	return &airports
}

// Some support functions to clean up the code a little
func (airports *Airports) dbClient() *mongo.Client {
	return airports.context.DBClient
}

func (airports *Airports) dbContext() context.Context {
	return airports.context.DBContext
}

func (airports *Airports) maxResults() int64 {
	return airports.context.MaxResults
}

func (airports *Airports) csvFile() (*os.File, error) {
	return os.Open(airports.context.AirportsCSV)
}

func (airports *Airports) logFile() (*os.File, error) {
	return airports.context.LogFile("airports")
}

func (airports *Airports) logPrint(s string) {
	airports.context.LogPrintln(s)
}

func (airports *Airports) logError(err error) {
	airports.context.LogError(err)
}

// GetByAirportCode retieves an Airport from the database based on its ICAO-Code
func (airports *Airports) GetByAirportCode(airportCode string) (*Airport, error) {
	var result Airport

	parameter, err := datatypes.ICAOAirportCode(airportCode, false, false)
	if err != nil {
		return nil, fmt.Errorf("GetByAirportCode.AirportCode(%s): %v", airportCode, err)
	}

	err = airports.collection.FindOne(airports.dbContext(),
		bson.D{{Key: "icao-airport-code", Value: parameter}}).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("Not Found")
	}

	return &result, nil
}

// GetByIATACode retrieves an Airport from the database based on its IATA-Code
func (airports *Airports) GetByIATACode(iataCode string) (*Airport, error) {
	var result Airport

	parameter, err := datatypes.IATAAirportCode(iataCode, false, false)
	if err != nil {
		return nil, fmt.Errorf("GetByIATACode.AirportCode(%s): %v", iataCode, err)
	}

	err = airports.collection.FindOne(airports.dbContext(),
		bson.D{{Key: "iata-airport-code", Value: parameter}}).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("Not Found")
	}

	return &result, nil
}

// GetList retrieves a list of Airports based on filter arguments
func (airports *Airports) GetList(countryCode string, regionCode string,
	fromICAO string, untilICAO string, fromIATA string, untilIATA string) ([]*Airport, error) {

	var result []*Airport
	var query = bson.D{{}}

	parameter, err := datatypes.ISOCountryCode(countryCode, false, true)
	if err != nil {
		return nil, fmt.Errorf("GetList.CountryCode(%s): %v", countryCode, err)
	}
	if len(parameter) != 0 {
		query = append(query, bson.E{Key: "iso-country-code", Value: parameter})
	}

	parameter, err = datatypes.ISORegionCode(regionCode, false, true)
	if err != nil {
		return nil, fmt.Errorf("GetList.RegionCode(%s): %v", regionCode, err)
	}
	if len(parameter) != 0 {
		query = append(query, bson.E{Key: "iso-region-code", Value: parameter})
	}

	parameter, err = datatypes.ICAOAirportCode(fromICAO, true, true)
	if err != nil {
		return nil, fmt.Errorf("GetList.FromICAO(%s): %v", fromICAO, err)
	}
	if len(parameter) != 0 {
		query = append(query, bson.E{Key: "icao-airport-code", Value: bson.D{{Key: "$gte", Value: parameter}}})
	}

	parameter, err = datatypes.ICAOAirportCode(untilICAO, true, true)
	if err != nil {
		return nil, fmt.Errorf("GetList.UntilICAO(%s): %v", untilICAO, err)
	}
	if len(parameter) != 0 {
		query = append(query, bson.E{Key: "icao-airport-code", Value: bson.D{{Key: "$lte", Value: parameter}}})
	}

	parameter, err = datatypes.IATAAirportCode(fromIATA, true, true)
	if err != nil {
		return nil, fmt.Errorf("GetList.FromIATA(%s): %v", fromIATA, err)
	}
	if len(parameter) != 0 {
		query = append(query, bson.E{Key: "iata-airport-code", Value: bson.D{{Key: "$gte", Value: parameter}}})
	}

	parameter, err = datatypes.IATAAirportCode(untilIATA, true, true)
	if err != nil {
		return nil, fmt.Errorf("GetList.UntilIATA(%s): %v", untilIATA, err)
	}
	if len(parameter) != 0 {
		query = append(query, bson.E{Key: "iata-airport-code", Value: bson.D{{Key: "$lte", Value: parameter}}})
	}

	findOptions := options.Find()
	findOptions.SetLimit(airports.maxResults() + 1)

	cur, err := airports.collection.Find(airports.dbContext(), query, findOptions)
	if err != nil {
		return nil, fmt.Errorf("Not found")
	}

	for cur.Next(airports.dbContext()) {
		var airport Airport
		cur.Decode(&airport)
		result = append(result, &airport)
	}

	cur.Close(airports.dbContext())

	if int64(len(result)) > airports.maxResults() {
		return nil, fmt.Errorf("Too many results")
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("Not found")
	}

	return result, nil
}

func (airports *Airports) importCSVLine(lineNumber int, line []string) error {

	// Skipping empty lines
	if len(line) == 0 {
		return nil
	}

	// Skip non-ICAO Airports
	airportCode, err := datatypes.ICAOAirportCode(line[1], false, false)
	if err != nil {
		return fmt.Errorf("Airport[%d].ICAO-Airport(%s): %v", lineNumber, line[1], err)
	}

	// Fill only valid IATA codes
	airportIATA, err := datatypes.IATAAirportCode(line[13], false, true)
	if err != nil {
		return fmt.Errorf("Airport[%d].IATA-Airport(%s): %v", lineNumber, line[13], err)
	}

	// Check for valid Country
	country, err := airports.countries.GetByCountryCode(line[8])
	if err != nil {
		return fmt.Errorf("Airport[%d].Country(%s): %v", lineNumber, line[8], err)
	}

	// Check for valid Region
	// The region key in the file is composed from the CountryCode and RegionCode
	regionKey := strings.Split(line[9], "-")
	if len(regionKey) != 2 {
		return fmt.Errorf("Airport[%d].Region(%s): %s", lineNumber, line[9], "Bad region key")
	}
	var region *countries.Region
	for i := range country.Regions {
		if country.Regions[i].RegionCode == regionKey[1] {
			region = country.Regions[i]
		}
	}
	if region == nil {
		return fmt.Errorf("Airport[%d].Region(%s): %v", lineNumber, line[9], "not found")
	}

	// Check Lattitude
	latitude, err := datatypes.Latitude(line[4], false)
	if err != nil {
		return fmt.Errorf("Airport[%d].Latitude: %v", lineNumber, err)
	}

	// Check Longitude
	longitude, err := datatypes.Longitude(line[5], false)
	if err != nil {
		return fmt.Errorf("Airport[%d].Longitude: %v", lineNumber, err)
	}

	// Check Elevation
	elevation, err := datatypes.Elevation(line[6], true)
	if err != nil {
		return fmt.Errorf("Airport[%d].Elevation: %v", lineNumber, err)
	}

	// Build internal representation
	airport := insertAirport{
		AirportCode:  airportCode,
		AirportName:  line[3],
		AirportType:  line[2],
		Latitude:     latitude,
		Longitude:    longitude,
		Elevation:    elevation,
		Country:      country.Country,
		CountryCode:  country.CountryCode,
		RegionCode:   region.RegionCode,
		Municipality: line[10],
		IATA:         airportIATA,
		Website:      line[15],
		Wikipedia:    line[16],
	}

	// Dump in mongo
	_, err = airports.collection.UpdateOne(airports.dbContext(),
		bson.D{{Key: "icao-airport-code", Value: airport.AirportCode}},
		bson.M{"$set": airport},
		options.Update().SetUpsert(true))

	if err != nil {
		return err
	}

	return nil
}

// ImportCSV imports a csv file into the Airports collection
func (airports *Airports) ImportCSV() error {

	// Open the airports.csv file
	csvFile, err := airports.csvFile()
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// Open the logfile
	_, err = airports.logFile()
	if err != nil {
		return err
	}

	// Skip the headerline
	reader := csv.NewReader(bufio.NewReader(csvFile))
	line, err := reader.Read()
	if err != nil {
		return err
	}

	airports.logPrint("Start Import")

	// Read the data
	// Line Numbers start at 1 and we've done the header
	lineNumber := 2
	line, err = reader.Read()
	for err == nil {
		err = airports.importCSVLine(lineNumber, line)
		airports.logError(err)
		line, err = reader.Read()
		lineNumber++
	}

	if err != io.EOF {
		return err
	}

	airports.logPrint("End Import")

	return nil
}
