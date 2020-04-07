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

// The definition of the queries ------------------------------------------------------------------

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"country":     countryQuery,
			"countries":   countriesQuery,
			"region":      regionQuery,
			"regions":     regionsQuery,
			"airport":     airportQuery,
			"airports":    airportsQuery,
			"runway":      runwayQuery,
			"runways":     runwaysQuery,
			"frequency":   frequencyQuery,
			"frequencies": frequenciesQuery,
		},
	})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
	})

// The interface of this component ----------------------------------------------------------------

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

// Init sets up the graphql module
func Init(countries *countries.Countries, airports *airports.Airports) error {

	// Register link to the database
	theCountries = countries
	theAirports = airports

	// Add referencials seperately to prevent circular references
	addCountryToRegion()
	addRegionToCountry()
	addCountryToAirport()
	addAirportToCountry()
	addAirportToRegion()
	addAirportToRunway()
	addRunwayToAirport()
	addAirportToFrequency()
	addFrequencyToAirport()

	return nil
}
