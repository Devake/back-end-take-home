package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//PathDistancePair is a small struct to hold one pair of path, distance pair
type PathDistancePair struct {
	Path     string
	Distance float64
}

var routes map[string]Route
var airports map[string]Airport
var airlines map[string]Airline

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

func convertStringArrayToOutputString(path []string) string {
	ret := ""
	count := 0
	for _, airport := range path {
		ret += airport
		if count < len(path)-1 {
			ret += "->"
		}
		count++
	}
	return ret
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
				key := convertStringArrayToOutputString(currentPath)
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
				getShortestRoute(k, dest, currentPath, shortestPath, checked, &currentDistance)
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
