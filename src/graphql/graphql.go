package graphql

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"

	"../airports"
	"../countries"
)

var theCountries *countries.Countries
var theAirports *airports.Airports

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"country": &graphql.Field{
				Type: countryType,
				Args: graphql.FieldConfigArgument{
					"CountryCode": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					countryCode, ok := p.Args["CountryCode"]
					if !ok {
						return nil, fmt.Errorf("Missing CountryCode parameter")
					}
					country, err := theCountries.GetByCountryCode(countryCode.(string))
					if err != nil {
						return nil, err
					}
					return country, nil
				},
			},
			"countries": &graphql.Field{
				Type: graphql.NewList(countryType),
				Args: graphql.FieldConfigArgument{
					"FromCountryCode": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"UntilCountryCode": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					fromCountryCode, ok := p.Args["FromCountryCode"]
					if !ok {
						return nil, fmt.Errorf("Missing FromCountryCode parameter")
					}
					untilCountryCode, ok := p.Args["UntilCountryCode"]
					if !ok {
						return nil, fmt.Errorf("Missing UntilCountryCode parameter")
					}
					countries, err := theCountries.GetList(fromCountryCode.(string), untilCountryCode.(string))
					if err != nil {
						return nil, err
					}
					return countries, nil
				},
			},
			"region": &graphql.Field{
				Type: regionType,
				Args: graphql.FieldConfigArgument{
					"CountryCode": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"RegionCode": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					countryCode, ok := p.Args["CountryCode"]
					if !ok {
						return nil, fmt.Errorf("Missing CountryCode parameter")
					}
					regionCode, ok := p.Args["RegionCode"]
					if !ok {
						return nil, fmt.Errorf("Missing RegionCode parameter")
					}
					country, err := theCountries.GetByCountryCode(countryCode.(string))
					if err != nil {
						return nil, err
					}
					for _, region := range country.Regions {
						if region.RegionCode == regionCode.(string) {
							return region, nil
						}
					}
					return nil, fmt.Errorf("Region:Not found")
				},
			},
			"airport": &graphql.Field{
				Type: airportType,
				Args: graphql.FieldConfigArgument{
					"ICAOCode": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"IATACode": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					airportCode, ok := p.Args["ICAOCode"]
					if ok {
						airport, err := theAirports.GetByAirportCode(airportCode.(string))
						if err != nil {
							return nil, fmt.Errorf("Airport(%s): %v", airportCode.(string), err)
						}
						return airport, nil
					}
					iataCode, ok := p.Args["IATACode"]
					if ok {
						airport, err := theAirports.GetByIATACode(iataCode.(string))
						if err != nil {
							return nil, fmt.Errorf("Airport(%s): %v", iataCode.(string), err)
						}
						return airport, nil
					}
					return nil, fmt.Errorf("Airport: Missing AirportCode or IATACode parameter")
				},
			},
		},
	})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
	})

// Handler resolves the calls to the graphql end-point
func Handler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	fmt.Println(query)

	output := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(output.Errors) > 0 {
		errors := fmt.Sprintf("failed to execute graphql operation, errors: %+v", output.Errors)
		http.Error(w, errors, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	result := json.NewEncoder(w)
	result.Encode(output)
}


// Init does it
func Init(countries *countries.Countries, airports *airports.Airports) error {

	// Register link to the database
	theCountries = countries
	theAirports = airports

	// Add referencials to prevent circular references
	addCountryToRegion()
	addCountryToAirport()
	addAirportToCountry()
	addAirportToRegion()
	addAirportToRunway()
	addAirportToFrequency()

	return nil
}
