package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var tmpl *template.Template

type fullData struct {
	Txt string
	Bdy string
}

func main() {

	tmpl, err := template.New("index").Parse("<!DOCTYPE html><html lang=\"en\" style=\"background-color:#00C176\"><head><meta charset=\"utf-8\"><meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><meta name=\"description\" content=\"\"><meta name=\"author\" content=\"\"><title>v1.0.0</title></head><body style=\"display:flex;align-items:center;justify-content:center;color:#FFFFFF;font-family:sans-serif;font-size:6rem;margin:0;letter-spacing:-0.1em\"><h1>{{.Txt}} {{if .Bdy}} {{.Bdy}} {{end}}</h1></body></html>")
	if err != nil {
		panic(err)
	}

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
			fd := fullData{text, ""}
			if err := tmpl.Execute(w, fd); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			// initialize HTTP client
			client := &http.Client{}
			req, _ := http.NewRequest("GET", "http://"+next+"/", nil)

			// get headers
			for k, _ := range r.Header {
				// set tracing headers
				for _, otHeader := range otHeaders {
					if strings.ToLower(otHeader) == strings.ToLower(k) {
						req.Header.Set(k, r.Header.Get(k))
					}
				}
			}

			// do request
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(w, "couldn't connect to http://"+next)
				fmt.Printf("Error: %s", err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			fd := fullData{text, string(body)}
			if fdErr := tmpl.Execute(w, fd); fdErr != nil {
				http.Error(w, fdErr.Error(), http.StatusInternalServerError)
			}
		}
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(w, "%s", "OK")
	})

	port := ":8080"
	log.Printf("Listening on %s....", port)
	http.ListenAndServe(port, mux)
}
