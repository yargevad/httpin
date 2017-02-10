package main

import (
	"log"
	"net/http"

	"github.com/pressly/chi"
)

type App struct{}

type Foo struct {
	Name string
	ID   int

	app *App
}

type Bar struct {
	Desc string
	ID   int

	app *App
}

func main() {
	router := chi.NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}
