package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		latency := os.Getenv("LATENCY")
		if latency != "" {
			i, err := strconv.ParseInt(latency, 10, 64)
			if err != nil {
				fmt.Fprintf(w, "Env LATENCY needs to be a number")
				return
			}
			// imitate latency in request
			time.Sleep(time.Duration(i) * time.Second)
		}

		text := os.Getenv("TEXT")
		if text == "" {
			fmt.Fprintf(w, "send env TEXT to display something")
			return
		}

		next := os.Getenv("NEXT")
		if next == "" {
			fmt.Fprintf(w, "%s", text)
		} else {
			// initialize HTTP client
			client := &http.Client{}
			req, _ := http.NewRequest("GET", "http://"+next+"/", nil)

			// get headers
			for k := range r.Header {
				// set tracing headers
				for _, otHeader := range otHeaders {
					if strings.ToLower(otHeader) == strings.ToLower(k) {
						req.Header.Set(k, r.Header.Get(k))
					}
				}
			}

			// make the request
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(w, "couldn't connect to http://"+next)
				fmt.Printf("Error: %s", err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			fmt.Fprintf(w, "%s %s", text, body)
		}
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(w, "%s", "OK")
	})

	port := ":8080"
	log.Printf("Listening on %s....", port)
	http.ListenAndServe(port, mux)
}
