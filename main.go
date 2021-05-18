package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"
	"gopkg.in/yaml.v2"
)

//http.ServeMux is an HTTP request multiplexer. It matches the URL of each incoming request
//against a list of registered patterns and calls the handler for the pattern that most closely
//matches the URL
func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	//HandlerFunc registers the handler function for the given pattern in the defaultServeMux
	mux.HandleFunc("/", hello) //arguments are pattern and handler
	return mux
}

//http.ResponseWriter is used by an http handler to construct an http response while http.Request
//is the request that is to be received by the server or to be sent by a client
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "Hello, World!")
}

//A handler responds to an http request
//ServeHTTP should write reply headers and data to the ResponseWriter and then return.
//Returning signals that the request is finished. It is not valid to use the ResponseWriter or read
//from the Request.Body after or concurrently with the completion of the ServeHTTP call.

//HandlerFunc type is an adapter to allow the use of ordinary functions such as HTTP handlers
// If it is a function with the appropriate signature HandlerFunc(f) is a Handler that calls f
func MapHandler(pathsToURLs map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if path, ok := pathsToURLs[r.URL.Path]; ok {
			http.Redirect(w, r, path, http.StatusFound)
		}
		fallback.ServeHTTP(w, r)
	}
}

func JSONHandler(jsondata []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var pathsToURLs []struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	}
	if err := json.Unmarshal(jsondata, &pathsToURLs); err != nil {
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

func YAMLHandler(yamldata []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var pathsToURLs []struct {
		Path string `yaml:"path"`
		URL  string `yaml:"url"`
	}

	if err := yaml.Unmarshal(yamldata, &pathsToURLs); err != nil {
		return nil, error
	}
	return func(w http.ResponseWriter, r *http.Request) {
		for _, urlPath := range pathsToURLs {
			if urlPath.Path == r.URL.Path {
				http.Redirect(w, r, urlPath.URL, http.StatusFound)
			}
		}
		fallback.ServeHTTP(w, r)
	}, nil
}

func BOLTHandler(db *bolt.DB, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("pathsToURLs"))
			if bucket != nil {
				cursor := bucket.Cursor()
				for path, url := cursor.First(); path != nil; url = cursor.Next() {
					if string(path) == r.URL.Path {
						http.Redirect(w, r, string(url), http.StatusFound)
						return nil
					}
				}
			}
			return nil
		}); err != nil {
			panic(err)
		}
		fallback.ServeHTTP(w, r)
	}
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
		yamlData, err := ioutil.ReadFile(yamlFile)
		if err != nil {
			panic(err)
		}
		//Build the YAMLHandler using the mapHandler as the fallback
		yamlHandler, err := YAMLHandler([]byte(yamlData), mapHandler)
		if err != nil {
			panic(err)
		}
		fmt.Println("Starting the server on:8080")
		http.ListenAndServe(":8080", yamlHandler)
	} else if boltFile != nil {
		db, err := bolt.Open(boltFile, 0600, nil)
		if err != nil {
			panic(err)
		}
		defer db.Close() //Close the database

		//Build the BoltHandler using the mapHandler as the fallback
		boltHandler := BOLTHandler(db, mapHandler)
		fmt.Println("Starting the server on:8080")
		http.ListenAndServe(":8080", boltHandler)
	} else {
		fmt.Println("Starting the server on:8080")
		http.ListenAndServe(":8080", mapHandler)
	}
}
