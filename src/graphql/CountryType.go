package graphql

import (
	"fmt"

	"../countries"
	"github.com/graphql-go/graphql"
)

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
		},
	})

func addCountryToRegion() {
	countryType.AddFieldConfig("Regions", &graphql.Field{
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
	})
}

func addCountryToAirport() {
	countryType.AddFieldConfig("Airports", &graphql.Field{
		Type: graphql.NewList(airportType),
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
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			country := p.Source.(*countries.Country)

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

			result, err := theAirports.GetList(country.CountryCode, "", fromICAOCode.(string), untilICAOCode.(string), fromIATACode.(string), untilIATACode.(string))
			if err != nil {
				return nil, fmt.Errorf("Country.Airports(): Not Found")
			}
			if len(result) == 0 {
				return nil, fmt.Errorf("Country.Airports(): Not found")
			}

			return result, nil
		},
	})

}
