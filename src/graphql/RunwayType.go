package graphql

import (
	"fmt"

	"github.com/graphql-go/graphql"

	"../airports"
)

// runwayView expresses a model where the runway is flattened and a back-link to 
// the airport added to be more suitable for graphql.
type runwayView struct {
	AirportCode   string  `json:"icao-airport-code"`
	RunwayCode    string  `json:"runway-code"`
	AltRunwayCode string  `json:"alt-runway-code"`
	Latitude      float64 `json:"latitude,omitempty"`
	Longitude     float64 `json:"longitude,omitempty"`
	Elevation     int     `json:"elevation,omitempty"`
	Heading       int     `json:"heading,omitempty"`
	Threshold     int     `json:"threshold,omitempty"`
	Length        int     `json:"length"`
	Width         int     `json:"width"`
	Surface       string  `json:"surface"`
	Lighted       bool    `json:"lighted"`
	Closed        bool    `json:"closed"`
}

func asRunwayView(airport *airports.Airport, runway *airports.Runway) []*runwayView {
	var result []*runwayView

	if len(runway.LowEnd.RunwayCode) > 0 {
		var runwayView runwayView

		runwayView.AirportCode = airport.AirportCode
		runwayView.RunwayCode = runway.LowEnd.RunwayCode
		runwayView.AltRunwayCode = runway.HighEnd.RunwayCode
		runwayView.Latitude = runway.LowEnd.Latitude
		runwayView.Longitude = runway.LowEnd.Longitude
		runwayView.Elevation = runway.LowEnd.Elevation
		runwayView.Heading = runway.LowEnd.Heading
		runwayView.Threshold = runway.LowEnd.Threshold
		runwayView.Length = runway.Length
		runwayView.Width = runway.Width
		runwayView.Surface = runway.Surface
		runwayView.Lighted = runway.Lighted
		runwayView.Closed = runway.Closed

		result = append(result, &runwayView)
	}

	if len(runway.HighEnd.RunwayCode) > 0 {
		var runwayView runwayView

		runwayView.AirportCode = airport.AirportCode
		runwayView.RunwayCode = runway.HighEnd.RunwayCode
		runwayView.AltRunwayCode = runway.LowEnd.RunwayCode
		runwayView.Latitude = runway.HighEnd.Latitude
		runwayView.Longitude = runway.HighEnd.Longitude
		runwayView.Elevation = runway.HighEnd.Elevation
		runwayView.Heading = runway.HighEnd.Heading
		runwayView.Threshold = runway.HighEnd.Threshold
		runwayView.Length = runway.Length
		runwayView.Width = runway.Width
		runwayView.Surface = runway.Surface
		runwayView.Lighted = runway.Lighted
		runwayView.Closed = runway.Closed

		result = append(result, &runwayView)
	}

	return result
}

// runwayType is the GraphQL representation of a Runway
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
			runway := p.Source.(*runwayView)

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
			runwaySides := asRunwayView(airport, runway)
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

		var result []*runwayView
		for _, runway := range airport.Runways {
			runwaySides := asRunwayView(airport, runway)
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
