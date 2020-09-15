package main

import (
	"crypto/tls"
	"fmt"
	"github.com/Jxck/logger"
	"html/template"
	"log"
	minimalist_http2 "minimalist-http2"
	"net/http"
	"strconv"
)

var loglevel int = 4

func init() {
	logger.Level(loglevel)
	logger.Debug("%v", loglevel)
}

type Person struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var Persons []Person = []Person{
	Person{0, "a"},
	Person{1, "b"},
	Person{2, "c"},
	Person{3, "d"},
	Person{4, "e"},
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello world")
}

var indexTmpl = template.Must(template.ParseFiles("index.html"))
var personTmpl = template.Must(template.ParseFiles("person.html"))

func PersonHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "POST" {
		r.ParseForm()

		id := len(Persons)
		name := r.Form["name"][0]

		person := Person{
			ID:   id,
			Name: name,
		}

		Persons = append(Persons, person)

		http.Redirect(w, r, "/persons", 302)
		return
	} else if r.Method == "GET" {
		id := r.URL.Query().Get("id")

		if id == "" {
			indexTmpl.Execute(w, Persons)
		} else {
			i, err := strconv.Atoi(id)
			if err != nil {
				log.Fatal(err)
			}

			if i+1 > len(Persons) {
				http.Redirect(w, r, "/persons", 302)
				return
			}

			person := Persons[i]
			personTmpl.Execute(w, person)
		}
	}
}

func main() {
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/persons", PersonHandler)
	// http.ListenAndServe(":3000", nil)

	// setup TLS config
	cert := "../keys/cert.pem"
	key := "../keys/key.pem"
	config := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{minimalist_http2.VERSION},
	}

	// setup Server
	server := &http.Server{
		Addr:           ":3000",
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
		TLSConfig:      config,
		TLSNextProto:   minimalist_http2.TLSNextProto,
	}

	fmt.Println(server.ListenAndServeTLS(cert, key))
}
