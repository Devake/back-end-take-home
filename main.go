package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Route struct {
	AirlineID   string
	Origin      string
	Destination string
	Distance    float64
}

var routes map[string]Route

type Airport struct {
	Name              string
	City              string
	Country           string
	IATA3             string
	Lat               float64
	Long              float64
	ConnectedAirports map[string]float64
}

var airports map[string]Airport

type Airline struct {
	Name           string
	TwoDigitCode   string
	ThreeDigitCode string
	Country        string
}

var airlines map[string]Airline

type PathDistancePair struct {
	Path     string
	Distance float64
}

type BackendTestResponse struct {
	Result string `json:"Result"`
	Error  string `json:"Error"`
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

//main import function
func importData(dataFilePath string) error {

	fileInfo, err := ioutil.ReadDir(dataFilePath)
	if err != nil {
		log.Fatal("Failed to read data directory:  ", err.Error(), ", ", dataFilePath)
		return err
	}

	if "/" != dataFilePath[len(dataFilePath)-1:] {
		dataFilePath += "/"
	}

	if airlines == nil {
		airlines = make(map[string]Airline)
	}
	if airports == nil {
		airports = make(map[string]Airport)
	}
	if routes == nil {
		routes = make(map[string]Route)
	}

	for _, file := range fileInfo {
		data, err := ioutil.ReadFile(dataFilePath + file.Name())
		if err != nil {
			log.Fatal("Failed to read data from file:  ", err.Error(), ", ", file.Name())
			return err
		}

		dataString := string(data)
		reader := csv.NewReader(strings.NewReader(dataString))
		if reader == nil {
			log.Fatal("File to load data to csv reader. File: ", dataFilePath)
		}

		rowCount := 0
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			if rowCount > 0 {
				if file.Name() == "airlines.csv" {
					decodeAirlineData(record, airlines)
				} else if file.Name() == "airports.csv" {
					decodeAirportData(record, airports)
				} else if file.Name() == "routes.csv" {
					decodeRouteData(record, routes, airports)
				}
			}
			rowCount++
		}
	}

	if len(airports) <= 0 {
		return errors.New("Airports data does not exist or invalid")
	}
	if len(routes) <= 0 {
		return errors.New("Routes data does not exist or invalid")
	}
	if len(airlines) <= 0 {
		return errors.New("Airlines data does not exist or invalid")

	}

	return nil
}

//initializes and starts executes the search
func startRouteSearch(origin string, dest string) (string, error) {
	checkedAirports := make(map[string]bool)  //airports which were already checked
	shortestPath := PathDistancePair{"", 0.0} //valid paths
	distance := 0.0                           //distance for current path
	if origin == dest {
		return "", errors.New("Origin and destination are the same")
	} else if _, ok := airports[origin]; !ok {
		return "", errors.New("Invalid origin airport")
	} else if _, ok := airports[dest]; !ok {
		return "", errors.New("Invalid dest airport")
	}

	checkedAirports[origin] = true
	getShortestRoute(origin, dest, nil, &shortestPath, checkedAirports, &distance)
	log.Println("shortest route: ", shortestPath.Path, " distance: ", shortestPath.Distance)
	if len(shortestPath.Path) <= 0 {
		return "", errors.New("Invalid route")
	}
	return shortestPath.Path, nil
}

//Recursively check all path from origin to dest
//Returns immediately if neighbour is destination or distance is greater than stored shortest distance
//Skips checking if current node was already checked
func getShortestRoute(origin string, dest string, currentPath []string, shortestPath *PathDistancePair, checked map[string]bool, distance *float64) error {
	originAirport := airports[origin]
	connectedAirports := originAirport.ConnectedAirports
	entryCount := 0
	if currentPath == nil {
		currentPath = make([]string, 0)
	}
	currentPath = append(currentPath, origin)
	//check if dest are neigbours
	for k, v := range connectedAirports {
		entryCount++
		if k == dest {
			totalDistance := *distance + v
			if shortestPath.Distance == 0.0 || totalDistance < shortestPath.Distance {
				currentPath = append(currentPath, k)
				var key string
				count := 0
				for _, airport := range currentPath {
					key += airport
					if count < len(currentPath)-1 {
						key += "->"
					}
					count++
				}
				shortestPath.Path = key
				shortestPath.Distance = totalDistance
			}
			currentPath = make([]string, 0)
			return nil
		}
	}
	checked[origin] = true
	//recurse through neighbours
	for k, v := range connectedAirports {
		currentDistance := 0.0
		// keep going
		currentDistance += *distance + v
		if shortestPath.Distance == 0.0 || currentDistance < shortestPath.Distance {
			if _, ok := checked[k]; !ok {
				if err := getShortestRoute(k, dest, currentPath, shortestPath, checked, &currentDistance); err != nil {
				}
			}
		}
	}
	checked = make(map[string]bool)
	return nil
}

func main() {
	args := os.Args
	if len(args) < 3 {
		log.Println("Usage: back-end-take-home <data directory> <server port>")
		return
	}

	//load data to memory
	filePath := os.Args[1]
	port := os.Args[2]

	start := time.Now()
	err := importData(filePath)
	end := time.Now()

	fmt.Println("Load time: ", end.Sub(start))
	if err != nil {
		log.Fatal("Failed to import data: ", err.Error())
		return
	}

	//GET method handler
	backendTestHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/backendTest" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}

		switch r.Method {
		case "GET":
			origin := r.URL.Query()["origin"]
			destination := r.URL.Query()["destination"]
			if len(origin[0]) <= 0 || len(destination[0]) <= 0 {
				fmt.Fprintf(w, "Invalid origin or destination parameters")
			} else {
				var result string
				var err error
				result, err = startRouteSearch(origin[0], destination[0])
				if err != nil {
					fmt.Fprintf(w, err.Error())
				} else {
					fmt.Fprintf(w, result)
				}
			}
		default:
			fmt.Fprintf(w, "Invalid method.")
		}
	}

	//HTTP endpoint
	http.HandleFunc("/backendTest", backendTestHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
