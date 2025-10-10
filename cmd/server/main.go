package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type MemStorage struct {
	MetricTypes map[string]*metricType
}

type metricType struct {
	Name    string
	Gauge   float64
	Counter int64
}

var memStorage = MemStorage{
	MetricTypes: make(map[string]*metricType),
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	if path[1] != "update" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(path) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	typeMetric := path[2]
	metricName := path[3]
	valueStr := path[4]

	if metricName == "" || valueStr == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var value interface{}
	var err error

	switch typeMetric {
	case "gauge":
		value, err = strconv.ParseFloat(valueStr, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case "counter":
		value, err = strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updateMetric(metricName, typeMetric, value)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Updated successfully")
}

func updateMetric(metricName, typeMetric string, value interface{}) {
	if _, ok := memStorage.MetricTypes[metricName]; !ok {
		memStorage.MetricTypes[metricName] = &metricType{
			Counter: 0,
			Gauge:   0,
		}
	}
	if typeMetric == "gauge" {
		memStorage.MetricTypes[metricName].Gauge = value.(float64)
	} else {
		memStorage.MetricTypes[metricName].Counter += value.(int64)
	}

	//fmt.Println(metricName)
	//fmt.Println(memStorage.MetricTypes[metricName].Gauge)
	//fmt.Println(memStorage.MetricTypes[metricName].Counter)

}

func main() {
	http.HandleFunc("/update/", updateHandler)
	log.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
