package graphql

import (
	"fmt"

	"github.com/graphql-go/graphql"

	"../countries"
)

// regionView is the external representation 'flattened' so it is easier to handle in
// graphql, for instance for back-linking in the graph
type regionView struct {
	CountryCode string `json:"iso-country-code"`
	RegionCode  string `json:"iso-region-code"`
	RegionName  string `json:"region-name"`
	Wikipedia   string `json:"wikipedia,omitempty"`
}

// asRegionView translates the internal view to the view more suitable for graphql:
// It contains a 'back-link' to the country; perhaps not normal for an optimized
// storage, but very useful for traversing graphs...
func asRegionView(country *countries.Country, region *countries.Region) *regionView {
	var result regionView

	result.CountryCode = country.CountryCode
	result.RegionCode = region.RegionCode
	result.RegionName = region.RegionName
	result.Wikipedia = region.Wikipedia

	return &result
}

// regionType is the representation of a region in GraphQL itself
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

// addRegionToCountry creates the link to countries seperately otherwise
// golang starts complaining about circular references.
func addRegionToCountry() {
	regionType.AddFieldConfig("Country", &graphql.Field{
		Type: countryType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			region := p.Source.(*regionView)

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
				return asRegionView(country, region), nil
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

		var result []*regionView
		countryList, err := theCountries.GetList(fromCountryCode.(string), untilCountryCode.(string))
		if err != nil {
			return nil, fmt.Errorf("Regions: %v", err)
		}
		for _, country := range countryList {
			for _, region := range country.Regions {
				if (!hasFromRegionCode || region.RegionCode >= fromRegionCode.(string)) && (!hasUntilRegionCode || region.RegionCode <= untilRegionCode.(string)) {
					result = append(result, asRegionView(country, region))
				}
			}
		}
		if len(result) == 0 {
			return nil, fmt.Errorf("Regions: Not found")
		}

		return result, nil
	}}
