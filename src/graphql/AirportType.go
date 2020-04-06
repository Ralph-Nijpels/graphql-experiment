package graphql

import (
	"fmt"

	"../airports"
	"../countries"
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
					return countries.AsRegionView(country, region), nil
				}
			}
			return nil, fmt.Errorf("Airport.Region: Not Found")
		},
	})

}

func addAirportToRunway() {
	airportType.AddFieldConfig("Runways", &graphql.Field{
		Type: graphql.NewList(runwayType),
		Args: graphql.FieldConfigArgument{
			"FromRunwayCode": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"UntilRunwayCode": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"FromHeading": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"UntilHeading": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"FromLength": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"UntilLength": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"Closed": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			airport := p.Source.(*airports.Airport)

			fromRunwayCode, hasFromRunwayCode := p.Args["FromRunwayCode"]
			untilRunwayCode, hasUntilRunwayCode := p.Args["UntilRunwayCode"]
			fromHeading, hasFromHeading := p.Args["FromHeading"]
			untilHeading, hasUntilHeading := p.Args["UntilHeading"]
			fromLength, hasFromLength := p.Args["FromLength"]
			untilLength, hasUntilLength := p.Args["UntilLength"]
			closed, hasClosed := p.Args["Closed"]

			var runways []*airports.RunwayView
			for _, runway := range airport.Runways {
				runwayViews := airports.AsRunwayView(airport, runway)
				for _, runwayView := range runwayViews {
					var addView = true
					if hasFromRunwayCode && runwayView.RunwayCode < fromRunwayCode.(string) {
						addView = false
					}
					if hasUntilRunwayCode && runwayView.RunwayCode > untilRunwayCode.(string) {
						addView = false
					}
					if hasFromHeading && runwayView.Heading < fromHeading.(int) {
						addView = false
					}
					if hasUntilHeading && runwayView.Heading > untilHeading.(int) {
						addView = false
					}
					if hasFromLength && runwayView.Length < fromLength.(int) {
						addView = false
					}
					if hasUntilLength && runwayView.Length > untilLength.(int) {
						addView = false
					}
					if hasClosed && runwayView.Closed != closed.(bool) {
						addView = false
					}
					if addView {
						runways = append(runways, runwayView)
					}
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

var airportQuery = &graphql.Field{
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
	}}

var airportsQuery = &graphql.Field{
	Type: graphql.NewList(airportType),
	Args: graphql.FieldConfigArgument{
		"CountryCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"RegionCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
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
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		countryCode, ok := p.Args["CountryCode"]
		if !ok {
			countryCode = ""
		}

		regionCode, ok := p.Args["RegionCode"]
		if !ok {
			regionCode = ""
		}

		fromICAOCode, ok := p.Args["FromICAOCode"]
		if !ok {
			fromICAOCode = ""
		}

		untilICAOCode, ok := p.Args["UntilICAOCode"]
		if !ok {
			untilICAOCode = ""
		}

		fromIATACode, ok := p.Args["FromIATACode"]
		if !ok {
			fromIATACode = ""
		}

		untilIATACode, ok := p.Args["UntilIATACode"]
		if !ok {
			untilIATACode = ""
		}

		result, err := theAirports.GetList(
			countryCode.(string),
			regionCode.(string),
			fromICAOCode.(string),
			untilICAOCode.(string),
			fromIATACode.(string),
			untilIATACode.(string))

		if err != nil {
			return nil, fmt.Errorf("Airports: %v", err)
		}

		return result, nil
	}}
