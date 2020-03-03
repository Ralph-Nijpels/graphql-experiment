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
	"../application"
	"../countries"
)

func main() {

	fmt.Println("Initializing..")
	context, err := application.GetContext()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Loading countries..")
	countries := countries.NewCountries(context)
	err = countries.ImportCSV()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Loading regions..")
	regions := countries.NewRegions()
	err = regions.ImportCSV()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Loading airports..")
	airports := airports.NewAirports(context, countries, regions)
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

	fmt.Println("Loading frequencies..")
	frequencies := airports.NewFrequencies()
	err = frequencies.ImportCSV()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Data loaded.")
}
