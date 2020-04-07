package graphql

import (
	"fmt"

	"../countries"
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

func addRegionToCountry() {
	regionType.AddFieldConfig("Country", &graphql.Field{
		Type: countryType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			region := p.Source.(*countries.RegionView)

			result, err := theCountries.GetByCountryCode(region.CountryCode)
			if err != nil {
				return nil, fmt.Errorf("Region.Country: %v", err)
			}

			return result, nil
		},
	})
}

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
				return countries.AsRegionView(country, region), nil
			}
		}
		return nil, fmt.Errorf("Region:Not found")
	}}

var regionsQuery = &graphql.Field{
	Type: graphql.NewList(regionType),
	Args: graphql.FieldConfigArgument{
		"FromCountryCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"UntilCountryCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"FromRegionCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"UntilRegionCode": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {

		fromCountryCode, hasFromCountryCode := p.Args["FromCountryCode"]
		untilCountryCode, hasUntilCountryCode := p.Args["UntilCountryCode"]
		if !hasFromCountryCode && !hasUntilCountryCode {
			return nil, fmt.Errorf("Missing From/Until CountryCode parameter")
		}
		if !hasFromCountryCode {
			fromCountryCode = ""
		}
		if !hasUntilCountryCode {
			untilCountryCode = ""
		}

		fromRegionCode, hasFromRegionCode := p.Args["FromRegionCode"]
		untilRegionCode, hasUntilRegionCode := p.Args["UntilRegionCode"]
		if !hasFromRegionCode {
			fromRegionCode = ""
		}
		if !hasUntilRegionCode {
			untilRegionCode = ""
		}

		var result []*countries.RegionView
		countryList, err := theCountries.GetList(fromCountryCode.(string), untilCountryCode.(string))
		if err != nil {
			return nil, fmt.Errorf("Regions: %v", err)
		}
		for _, country := range countryList {
			for _, region := range country.Regions {
				if (!hasFromRegionCode || region.RegionCode >= fromRegionCode.(string)) && (!hasUntilRegionCode || region.RegionCode <= untilRegionCode.(string)) {
					result = append(result, countries.AsRegionView(country, region))
				}
			}
		}
		if len(result) == 0 {
			return nil, fmt.Errorf("Regions: Not found")
		}

		return result, nil
	}}
