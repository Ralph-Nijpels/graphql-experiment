package graphql

import (
	"fmt"

	"../airports"
	"github.com/graphql-go/graphql"
)

var airportType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Airport",
		Fields: graphql.Fields{
			"ICAOCode": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					airport := p.Source.(*airports.Airport)
					return airport.AirportCode, nil
				},
			},
			"AirportName": &graphql.Field{
				Type: graphql.String,
			},
			"AirportType": &graphql.Field{
				Type: graphql.String,
			},
			"Latitude": &graphql.Field{
				Type: graphql.Float,
			},
			"Longitude": &graphql.Field{
				Type: graphql.Float,
			},
			"Elevation": &graphql.Field{
				Type: graphql.Int,
			},
			"Region": &graphql.Field{
				Type: regionType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					airport := p.Source.(*airports.Airport)
					country, err := theCountries.GetByCountryCode(airport.CountryCode)
					if err != nil {
						return nil, fmt.Errorf("Airport.Region: %v", err)
					}
					for _, region := range country.Regions {
						if region.RegionCode == airport.RegionCode {
							return region, nil
						}
					}
					return nil, fmt.Errorf("Airport.Region: Not Found")
				},
			},
			"Municipality": &graphql.Field{
				Type: graphql.String,
			},
			"IATACode": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					airport := p.Source.(*airports.Airport)
					return airport.IATA, nil
				},
			},
			"Website": &graphql.Field{
				Type: graphql.String,
			},
			"Wikipedia": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

func addAirportToCountry() {
	airportType.AddFieldConfig("Country", &graphql.Field{
		Type: countryType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			airport := p.Source.(*airports.Airport)
			country, err := theCountries.GetByCountryCode(airport.CountryCode)
			if err != nil {
				return nil, fmt.Errorf("Airport.Country: %v", err)
			}
			return country, nil
		},
	})
}

func addAirportToRegion() {
	airportType.AddFieldConfig("Region", &graphql.Field{
		Type: regionType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			airport := p.Source.(*airports.Airport)
			country, err := theCountries.GetByCountryCode(airport.CountryCode)
			if err != nil {
				return nil, fmt.Errorf("Airport.Region: %v", err)
			}
			for _, region := range country.Regions {
				if region.RegionCode == airport.RegionCode {
					return region, nil
				}
			}
			return nil, fmt.Errorf("Airport.Region: Not Found")
		},
	})

}

func addAirportToRunway() {
	airportType.AddFieldConfig("Runways", &graphql.Field{
		Type: graphql.NewList(runwayType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			airport := p.Source.(*airports.Airport)

			var runways []*airports.RunwayView
			for _, runway := range airport.Runways {
				runwayViews := airports.AsRunwayView(runway)
				for _, runwayView := range runwayViews {
					runways = append(runways, runwayView)
				}
			}

			return runways, nil
		},
	})
}

func addAirportToFrequency() {
	airportType.AddFieldConfig("Frequencies", &graphql.Field{
		Type: graphql.NewList(frequencyType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			airport := p.Source.(*airports.Airport)
			return airport.Frequencies, nil
		},
	})
}