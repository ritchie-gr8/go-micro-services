package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "80"

type Config struct{}

func main() {
	app := Config{}
	handler := app.routes()

	// define server
	log.Printf("Starting broker server on port: %s", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: handler,
	}

	// start the server
	if err := srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
