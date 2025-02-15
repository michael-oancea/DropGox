package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Response is a helper structure for our JSON responses.
type Response struct {
	Message string `json:"message"`
}

func main() {
	// Define a command-line flag for the port with a default of "8080".
	// We'll use both -port and -p as aliases.
	var port string
	flag.StringVar(&port, "port", "8080", "port to run the server on")
	flag.StringVar(&port, "p", "8080", "port to run the server on (shorthand)")
	flag.Parse()

	// Initialize a new router
	router := mux.NewRouter()

	// Health-check endpoint
	router.HandleFunc("/health", HealthCheckHandler).Methods("GET")

	// Example file endpoints (placeholders)
	router.HandleFunc("/upload", UploadHandler).Methods("POST")
	router.HandleFunc("/download/{filename}", DownloadHandler).Methods("GET")

	// Start the server on the specified port
	addr := ":" + port
	log.Printf("DropGox Backend is running on port %s...\n", port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

// HealthCheckHandler provides a simple health-check endpoint.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{Message: "DropGox Backend is up and running!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UploadHandler is a placeholder for file upload functionality.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{Message: "Upload endpoint - to be implemented"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DownloadHandler is a placeholder for file download functionality.
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	response := Response{Message: "Download endpoint for file: " + filename + " - to be implemented"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
