package main

// Data-loader puts the Country, Region and Airport information from CSV files
// in the MongoDB.
//
// Files are retrieved from ourairports.com/data/xx.csv
//
// Note: it is written quite sloppily:
// - file names and database connection are hard-coded
// - error logging is not implemented

import (
	"fmt"
	"log"

	"../airports"
	"../countries"
	"../database"
	"../regions"
)

func main() {

	fmt.Println("Connecting to MongoDB..")
	mongoClient, err := database.NewContext()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Loading countries..")
	countries := countries.NewCountries(mongoClient)
	err = countries.ImportCSV()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Loading regions..")
	regions := regions.NewRegions(mongoClient, countries)
	err = regions.ImportCSV()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Loading airports..")
	airports := airports.NewAirports(mongoClient, countries, regions)
	err = airports.ImportCSV()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Loading runways..")
	runways := airports.NewRunways()
	err = runways.ImportCSV()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Data loaded.")
}
