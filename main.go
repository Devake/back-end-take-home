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

var leastValidLayer = math.MaxInt32
var checkedAirports map[string]int

//initializes and starts executes the search
func startRouteSearch(origin string, dest string) (string, error) {
	if origin == dest {
		return "", errors.New("Origin and destination are the same")
	} else if _, ok := airports[origin]; !ok {
		return "", errors.New("Invalid origin airport")
	} else if _, ok := airports[dest]; !ok {
		return "", errors.New("Invalid dest airport")
	}
	leastValidLayer = math.MaxInt32
	checkedAirports = make(map[string]int, 0)
	startTime := time.Now()
	fmt.Println("****************search started ****************************")
	shortestPath := getShortestRoute(origin, dest, origin, nil, 0)
	endTime := time.Now()
	fmt.Println("****************search ended: ", endTime.Sub(startTime)*time.Nanosecond, "ms****************************")
	fmt.Println("shortestPath: ", shortestPath)
	if len(shortestPath) <= 0 {
		return "", errors.New("Invalid route")
	}
	return shortestPath, nil
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
func getShortestRoute(start string, dest string, origin string, currentPath []string, layer int) string {
	startAirports := airports[start]
	connectedAirports := startAirports.ConnectedAirports
	if currentPath == nil {
		currentPath = make([]string, 0)
	}
	currentPath = append(currentPath, start)
	//check for direct flight. assume it will always be the shortest
	if _, found := connectedAirports[dest]; found {
		if layer < leastValidLayer {
			leastValidLayer = layer
		}
		checkedAirports[start] = leastValidLayer
		return convertStringArrayToOutputString(append(currentPath, dest))
	}

	//Already longer or equal to shortest result. Skip searching
	if layer+1 >= leastValidLayer {
		return ""
	}

	//Already longer or equal to shortest result. Skip searching
	if checkedLayer, found := checkedAirports[start]; found {
		if layer >= checkedLayer {
			return ""
		}
	} else {
		checkedAirports[start] = layer
	}

	var ret = ""
	for key := range connectedAirports {
		result := getShortestRoute(key, dest, origin, currentPath, layer+1)
		if len(result) > 0 {
			ret = result
		}
		if layer < checkedAirports[key] {
			checkedAirports[key] = layer
		}
	}
	return ret
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
				fmt.Fprintln(w, "Invalid origin or destination parameters")
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
