package graphql

import (
	"github.com/graphql-go/graphql"
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
