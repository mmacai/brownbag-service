package main

import (
	"fmt"
	"net/http"

	"./db"
	"./routes"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	const port = "7000"

	// New gorilla mux router
	router := mux.NewRouter()

	// Map handlers to endpoints
	router.HandleFunc("/api/votes", routes.GetVotes).Methods("GET")
	router.HandleFunc("/api/votes", routes.UpdateVotes).Methods("POST")
	router.HandleFunc("/api/docker", routes.GetContainerInfo).Methods("GET")

	// Map router to http server
	http.Handle("/", router)

	// Init database
	db.Init()

	// Start new http server
	fmt.Println("Server running at localhost:" + port)
	handler := cors.Default().Handler(router)
	panic(http.ListenAndServe(":"+port, handler))
}
