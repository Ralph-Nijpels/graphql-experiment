package graphql

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

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

var regionQuery = &graphql.Field{
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
		return nil, fmt.Errorf("Region:Not found")
	}}
