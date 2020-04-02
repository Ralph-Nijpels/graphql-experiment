package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"../airports"
	"../application"
	"../countries"
	"../graphql"
)

var theCountries *countries.Countries

//var theRegions *countries.Regions
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

	context, err := application.GetContext()
	if err != nil {
		log.Panic(err)
	}

	theCountries = countries.NewCountries(context)
	theAirports = airports.NewAirports(context, theCountries)

	graphql.Init(theCountries, theAirports)

	myRouter := mux.NewRouter()
	myRouter.HandleFunc("/geography/countries", getCountries).Methods("GET")
	myRouter.HandleFunc("/geography/countries/{country-code}", getCountry).Methods("GET")
	myRouter.HandleFunc("/geography/airports", getAirports).Methods("GET")
	myRouter.HandleFunc("/geography/airports/{airport-code}", getAirport).Methods("GET")
	myRouter.HandleFunc("/geography/graphql", graphql.Handler).Methods("POST")

	http.ListenAndServe(":8090", myRouter)

}
