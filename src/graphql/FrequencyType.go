package graphql

import (
	"fmt"

	"github.com/graphql-go/graphql"

	"../airports"
)

// frequencyView is a representation to help in graphql by adding a back-link to the airport
type frequencyView struct {
	AirportCode   string  `json:"icao-airport-code"`
	FrequencyType string  `json:"frequency-type"`
	Description   string  `json:"description,omitempty"`
	Frequency     float64 `json:"frequency-mhz"`
}

func asFrequencyView(airport *airports.Airport, frequency *airports.Frequency) *frequencyView {
	var result frequencyView

	result.AirportCode = airport.AirportCode
	result.FrequencyType = frequency.FrequencyType
	result.Description = frequency.Description
	result.Frequency = frequency.Frequency

	return &result
}

// frequencyType is the graphql representation of a frequency
var frequencyType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Frequency",
		Fields: graphql.Fields{
			"FrequencyType": &graphql.Field{
				Type: graphql.String,
			},
			"Description": &graphql.Field{
				Type: graphql.String,
			},
			"Frequency": &graphql.Field{
				Type: graphql.Float,
			},
		},
	})

func addFrequencyToAirport() {
	frequencyType.AddFieldConfig("Airport", &graphql.Field{
		Type: airportType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			frequency := p.Source.(*frequencyView)

			result, err := theAirports.GetByAirportCode(frequency.AirportCode)
			if err != nil {
				return nil, fmt.Errorf("Frequency.Airport: %v", err)
			}

			return result, nil
		},
	})
}

var frequencyQuery = &graphql.Field{
	Type: frequencyType,
	Args: graphql.FieldConfigArgument{
		"ICAOCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"IATACode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"FrequencyType": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		var airport *airports.Airport
		var err error

		icaoCode, ok := p.Args["ICAOCode"]
		if ok {
			airport, err = theAirports.GetByAirportCode(icaoCode.(string))
			if err != nil {
				return nil, fmt.Errorf("Frequency(%s): %v", icaoCode.(string), err)
			}
		}

		iataCode, ok := p.Args["IATACode"]
		if ok {
			airport, err = theAirports.GetByIATACode(iataCode.(string))
			if err != nil {
				return nil, fmt.Errorf("Frequency(%s): %v", iataCode.(string), err)
			}
		}

		frequencyType, ok := p.Args["FrequencyType"]
		if !ok {
			return nil, fmt.Errorf("Frequency: missing FrequencyType parameter")
		}

		for _, frequency := range airport.Frequencies {
			if frequency.FrequencyType == frequencyType {
				return asFrequencyView(airport, frequency), nil
			}
		}

		return nil, fmt.Errorf("Frequency: not found")
	}}

var frequenciesQuery = &graphql.Field{
	Type: graphql.NewList(frequencyType),
	Args: graphql.FieldConfigArgument{
		"FromICAOCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"UntilICAOCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"FromIATACode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"UntilIATACode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"FromFrequencyType": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"UntilFrequencyType": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {

		fromICAOCode, hasFromICAOCode := p.Args["FromICAOCode"]
		untilICAOCode, hasUntilICAOCode := p.Args["UntilICAOCode"]
		fromIATACode, hasFromIATACode := p.Args["FromIATACode"]
		untilIATACode, hasUntilIATACode := p.Args["UntilIATACode"]
		if !hasFromICAOCode && !hasUntilICAOCode && !hasFromIATACode && !hasUntilIATACode {
			return nil, fmt.Errorf("Frequencies: Missing From/Until airport selection")
		}
		if !hasFromICAOCode {
			fromICAOCode = ""
		}
		if !hasUntilICAOCode {
			untilICAOCode = ""
		}
		if !hasFromIATACode {
			fromIATACode = ""
		}
		if !hasUntilIATACode {
			untilIATACode = ""
		}

		fromFrequencyType, hasFromFrequencyType := p.Args["FromFrequencyType"]
		untilFrequencyType, hasUntilFrequencyType := p.Args["UntilFrequencyType"]
		if !hasFromFrequencyType {
			fromFrequencyType = ""
		}
		if !hasUntilFrequencyType {
			untilFrequencyType = ""
		}

		airportList, err := theAirports.GetList("", "",
			fromICAOCode.(string),
			untilICAOCode.(string),
			fromIATACode.(string),
			untilIATACode.(string))
		if err != nil {
			return nil, fmt.Errorf("Frequencies: %v", err)
		}

		var result []*frequencyView
		for _, airport := range airportList {
			for _, frequency := range airport.Frequencies {
				addView := true
				if hasFromFrequencyType && frequency.FrequencyType < fromFrequencyType.(string) {
					addView = false
				}
				if hasUntilFrequencyType && frequency.FrequencyType > untilFrequencyType.(string) {
					addView = false
				}
				if addView {
					result = append(result, asFrequencyView(airport, frequency))
				}
			}
		}

		if len(result) == 0 {
			return nil, fmt.Errorf("Frequencies: not found")
		}

		return result, nil
	}}
