package graphql

import (
	"fmt"

	"github.com/graphql-go/graphql"

	"../airports"
)

var runwayType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Runway",
		Fields: graphql.Fields{
			"RunwayCode": &graphql.Field{
				Type: graphql.String,
			},
			"AltRunwayCode": &graphql.Field{
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
			"Heading": &graphql.Field{
				Type: graphql.Int,
			},
			"Threshold": &graphql.Field{
				Type: graphql.Int,
			},
			"Length": &graphql.Field{
				Type: graphql.Int,
			},
			"Width": &graphql.Field{
				Type: graphql.Int,
			},
			"Surface": &graphql.Field{
				Type: graphql.String,
			},
			"Lighted": &graphql.Field{
				Type: graphql.Boolean,
			},
			"Closed": &graphql.Field{
				Type: graphql.Boolean,
			},
		},
	})

func addRunwayToAirport() {
	runwayType.AddFieldConfig("Airport", &graphql.Field{
		Type: airportType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			runway := p.Source.(*airports.RunwayView)

			result, err := theAirports.GetByAirportCode(runway.AirportCode)
			if err != nil {
				return nil, fmt.Errorf("Runway.Airport: %v", err)
			}

			return result, nil
		},
	})
}

var runwayQuery = &graphql.Field{
	Type: runwayType,
	Args: graphql.FieldConfigArgument{
		"ICAOCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"IATACode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"RunwayCode": &graphql.ArgumentConfig{
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
				return nil, fmt.Errorf("Runway(%s): %v", icaoCode.(string), err)
			}
		}

		iataCode, ok := p.Args["IATACode"]
		if ok {
			airport, err = theAirports.GetByIATACode(iataCode.(string))
			if err != nil {
				return nil, fmt.Errorf("Runway(%s): %v", iataCode.(string), err)
			}
		}

		if airport == nil {
			return nil, fmt.Errorf("Runway: Missing AirportCode or IATACode parameter")
		}

		runwayCode, ok := p.Args["RunwayCode"]
		if !ok {
			return nil, fmt.Errorf("Runway: Missing RunwayCode parameter")
		}

		for _, runway := range airport.Runways {
			runwaySides := airports.AsRunwayView(airport, runway)
			for _, runwayView := range runwaySides {
				if runwayView.RunwayCode == runwayCode.(string) {
					return runwayView, nil
				}
			}
		}

		return nil, fmt.Errorf("Not Found")
	},
}

var runwaysQuery = &graphql.Field{
	Type: graphql.NewList(runwayType),
	Args: graphql.FieldConfigArgument{
		"ICAOCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"IATACode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
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
		var airport *airports.Airport
		var err error

		icaoCode, hasICAOCode := p.Args["ICAOCode"]
		if hasICAOCode {
			airport, err = theAirports.GetByAirportCode(icaoCode.(string))
			if err != nil {
				return nil, fmt.Errorf("Runways(%s): %v", icaoCode.(string), err)
			}
		}

		iataCode, hasIATACode := p.Args["IATACode"]
		if hasIATACode {
			airport, err = theAirports.GetByIATACode(iataCode.(string))
			if err != nil {
				return nil, fmt.Errorf("Runways(%s): %v", iataCode.(string), err)
			}
		}

		if airport == nil {
			return nil, fmt.Errorf("Runways: Missing AirportCode or IATACode parameter")
		}

		fromRunwayCode, hasFromRunwayCode := p.Args["FromRunwayCode"]
		untilRunwayCode, hasUntilRunwayCode := p.Args["UntilRunwayCode"]
		fromHeading, hasFromHeading := p.Args["FromHeading"]
		untilHeading, hasUntilHeading := p.Args["UntilHeading"]
		fromLength, hasFromLength := p.Args["FromLength"]
		untilLength, hasUntilLength := p.Args["UntilLength"]
		closed, hasClosed := p.Args["Closed"]

		var result []*airports.RunwayView
		for _, runway := range airport.Runways {
			runwaySides := airports.AsRunwayView(airport, runway)
			for _, runwayView := range runwaySides {
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
					result = append(result, runwayView)
				}
			}
		}

		if len(result) == 0 {
			return nil, fmt.Errorf("Runways: Not Found")
		}

		return result, nil
	},
}
