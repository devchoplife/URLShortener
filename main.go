package main

import (
	"encoding/json"
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

func MapHandler(pathsToURLs map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if path, ok := pathsToURLs[r.URL.Path]; ok {
			http.Redirect(w, r, path, http.StatusFound)
		}
		fallback.ServeHTTP(w, r)
	}
}

func JSONHandler(jsonData []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var pathsToURLs []struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	}
	if err := json.Unmarshal(jsonData, &pathsToURLs); err != nil {
		return nil, err
	}
	return func(w http.ResponseWriter, r *http.Request) {
		for _, urlPath := range pathsToURLs {
			if urlPath.Path == r.URL.Path {
				http.Redirect(w, r, urlPath, http.StatusFound)
			}
		}
		fallback.ServeHTTP(w, r)
	}, nil
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
		if err != nil {
			panic(err)
		}

		//Build the JSONHandler using the mapHandler as the fallback
		jsonHandler, err := JSONHandler([]byte(jsonData), mapHandler)
		if err != nil {
			panic(err)
		}
		fmt.Println("Starting the server on:8080")
		http.ListenAndServe(":8080", jsonHandler)
	} else if yamlFile != "" {
		
	}
