package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func main() {
	var jsonFile, yamlFile, boltFile string
	flag.StringVar(&jsonFile, "json", "", "path to the json file")
	flag.StringVar(&yamlFile, "yaml", "", "path to the yaml file")
	flag.StringVar(&boltFile, "bolt", "", "path to the boltdb file")

	flag.Parse()

	mux := defaultMux()

	//build the maphandler using the mux as its fallback
	pathsToURLs := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}

	mapHandler := MapHandler(pathsToURLs, mux)

	if jsonFile != "" {
		jsonData, err := ioutil.ReadFile(jsonFile)
	}
}
