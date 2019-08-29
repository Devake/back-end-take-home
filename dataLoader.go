package main

import (
	"math"
	"strconv"
)

type Route struct {
	AirlineID   string
	Origin      string
	Destination string
	Distance    float64
}

type Airport struct {
	Name              string
	City              string
	Country           string
	IATA3             string
	Lat               float64
	Long              float64
	ConnectedAirports map[string]float64
}

type Airline struct {
	Name           string
	TwoDigitCode   string
	ThreeDigitCode string
	Country        string
}

//Helper function to populate airline data from CSV
func decodeAirlineData(record []string, airline map[string]Airline) error {
	key := record[1]
	a := Airline{
		Name:           record[0],
		TwoDigitCode:   record[1],
		ThreeDigitCode: record[2],
		Country:        record[3],
	}
	airline[key] = a
	return nil
}

//Helper function to populate airport data from CSV
func decodeAirportData(record []string, airport map[string]Airport) error {
	key := record[3]

	if _, ok := airport[key]; ok {
		return nil
	}

	latFloat, _ := strconv.ParseFloat(record[4], 64)
	longFloat, _ := strconv.ParseFloat(record[5], 64)
	a := Airport{
		Name:              record[0],
		City:              record[1],
		Country:           record[2],
		IATA3:             record[3],
		Lat:               latFloat,
		Long:              longFloat,
		ConnectedAirports: make(map[string]float64),
	}
	airport[key] = a
	return nil
}

//Helper function to populate route data from CSV. Also calculates and stores the distance between airports
func decodeRouteData(record []string, route map[string]Route, airports map[string]Airport) error {
	if len(airports) <= 0 {
		return nil
	}

	origin := airports[record[1]]
	destination := airports[record[2]]
	key := record[1] + "_" + record[2]
	distance := math.Sqrt(math.Pow(destination.Lat-origin.Lat, 2.0) + math.Pow(destination.Long-origin.Long, 2.0))
	r := Route{
		AirlineID:   record[0],
		Origin:      record[1],
		Destination: record[2],
		Distance:    distance,
	}

	if _, ok := routes[key]; !ok {
		routes[key] = r
		if a, ok := airports[r.Origin]; ok {
			if _, ok := a.ConnectedAirports[r.Destination]; !ok {
				a.ConnectedAirports[r.Destination] = r.Distance
			}
		}
	}
	return nil
}
