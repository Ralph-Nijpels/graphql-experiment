package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"../airports"
	"../countries"
	"../database"
	"../regions"
)

var theCountries *countries.Countries
var theRegions *regions.Regions
var theAirports *airports.Airports

func getCountries(w http.ResponseWriter, r *http.Request) {

	fromCountry := r.FormValue("from")
	untilCountry := r.FormValue("until")

	countryList, err := theCountries.GetList(fromCountry, untilCountry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	result := json.NewEncoder(w)
	result.Encode(countryList)
}

func getCountry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	countryCode := vars["country-code"]

	country, err := theCountries.GetByCountryCode(countryCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	result := json.NewEncoder(w)
	result.Encode(country)
}

func getRegions(w http.ResponseWriter, r *http.Request) {
	countryCode := r.FormValue("country")
	fromRegion := r.FormValue("from")
	untilRegion := r.FormValue("until")

	regionList, err := theRegions.GetList(countryCode, fromRegion, untilRegion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	result := json.NewEncoder(w)
	result.Encode(regionList)
}

func getRegion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	regionCode := vars["region-code"]

	region, err := theRegions.GetByRegionCode(regionCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	result := json.NewEncoder(w)
	result.Encode(region)
}

func getAirports(w http.ResponseWriter, r *http.Request) {
	countryCode := r.FormValue("country")
	regionCode := r.FormValue("region")
	fromICAO := r.FormValue("from")
	untilICAO := r.FormValue("until")
	fromIATA := r.FormValue("from-iata")
	untilIATA := r.FormValue("until-iata")

	airportList, err := theAirports.GetList(countryCode, regionCode, fromICAO, untilICAO, fromIATA, untilIATA)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	result := json.NewEncoder(w)
	result.Encode(airportList)
}

func getAirport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	airportCode := vars["airport-code"]

	region, err := theAirports.GetByAirportCode(airportCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	result := json.NewEncoder(w)
	result.Encode(region)
}

func main() {
	var err error

	dbContext, err := database.NewContext()
	if err != nil {
		log.Panic(err)
	}

	theCountries = countries.NewCountries(dbContext)
	theRegions = regions.NewRegions(dbContext, theCountries)
	theAirports = airports.NewAirports(dbContext, theCountries, theRegions)

	myRouter := mux.NewRouter()
	myRouter.HandleFunc("/geography/countries", getCountries).Methods("GET")
	myRouter.HandleFunc("/geography/countries/{country-code}", getCountry).Methods("GET")
	myRouter.HandleFunc("/geography/regions", getRegions).Methods("GET")
	myRouter.HandleFunc("/geography/regions/{region-code}", getRegion).Methods("GET")
	myRouter.HandleFunc("/geography/airports", getAirports).Methods("GET")
	myRouter.HandleFunc("/geography/airports/{airport-code}", getAirport).Methods("GET")

	http.ListenAndServe(":8090", myRouter)

}
