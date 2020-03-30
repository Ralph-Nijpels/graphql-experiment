package graphql

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"

	"../countries"
)

var theCountries *countries.Countries

var regionType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Region",
		Fields: graphql.Fields{
			"RegionCode": &graphql.Field{
				Type: graphql.String,
			},
			"RegionName": &graphql.Field{
				Type: graphql.String,
			},
			"Wikipedia": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

var countryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Country",
		Fields: graphql.Fields{
			"CountryCode": &graphql.Field{
				Type: graphql.String,
			},
			"CountryName": &graphql.Field{
				Type: graphql.String,
			},
			"Continent": &graphql.Field{
				Type: graphql.String,
			},
			"Wikipedia": &graphql.Field{
				Type: graphql.String,
			},
			"Regions": &graphql.Field{
				Type: graphql.NewList(regionType),
				Args: graphql.FieldConfigArgument{
					"FromRegionCode": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"UntilRegionCode": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					country := p.Source.(*countries.Country)
					fromRegionCode, ok := p.Args["FromRegionCode"]

					if !ok {
						fromRegionCode = ""
					}

					untilRegionCode, ok := p.Args["UntilRegionCode"]
					if !ok {
						untilRegionCode = "ZZ"
					}

					var result []*countries.Region
					for _, region := range country.Regions {
						if region.RegionCode >= fromRegionCode.(string) && region.RegionCode <= untilRegionCode.(string) {
							result = append(result, region)
						}
					}

					if len(result) == 0 {
						return nil, fmt.Errorf("Not found")
					}

					return result, nil
				},
			},
		},
	})

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
					return nil, fmt.Errorf("Not found")
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
func Init(countries *countries.Countries) error {
	theCountries = countries
	return nil
}
