package graphql

import (
	"github.com/graphql-go/graphql"
)

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

