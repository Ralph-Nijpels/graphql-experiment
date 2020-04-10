package graphql

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	var graphqlRequest struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}

	// Verify content type
	contentType := r.Header.Get("Content-type")
	if contentType != "application/json" {
		errors := fmt.Sprintf("failed to execute graphql operation, use application/json")
		http.Error(w, errors, http.StatusBadRequest)
		return
	}

	// Parse the request
	buffer, err := ioutil.ReadAll(r.Body)
	if err == nil {
		err = json.Unmarshal(buffer, &graphqlRequest)
	}
	if err != nil {
		errors := fmt.Sprintf("failed to execute graphql operation, errors: %v", err)
		http.Error(w, errors, http.StatusBadRequest)
		return
	}

	// Run the query
	output := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  graphqlRequest.Query,
		VariableValues: graphqlRequest.Variables,
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
