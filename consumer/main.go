// Copyright © 2017 AMIM KNABBEN <amim.knabben@gmail.com>

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/knabben/chatops/consumer/cmd"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "")
}

func main() {
	log.Println("Initializing HttpGet liveness")
	go func() {
		http.HandleFunc("/ping", handler)
		http.ListenAndServe(":8080", nil)
	}()
	// init subcommands
	cmd.Execute()
}
