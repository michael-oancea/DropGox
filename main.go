package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

// Response is a helper structure for our JSON responses.
type Response struct {
	Message string `json:"message"`
}

func main() {
	// Define command-line flags for the port.
	var port string
	flag.StringVar(&port, "port", "8080", "port to run the server on")
	flag.StringVar(&port, "p", "8080", "port to run the server on (shorthand)")
	flag.Parse()

	// Ensure the storage directory exists
	storageDir := "./files"
	if err := os.MkdirAll(storageDir, os.ModePerm); err != nil {
		log.Fatal("Error creating storage directory:", err)
	}

	// Initialize a new router.
	router := mux.NewRouter()

	// Health-check endpoint.
	router.HandleFunc("/health", HealthCheckHandler).Methods("GET")

	// File operation endpoints.
	router.HandleFunc("/upload", UploadHandler(storageDir)).Methods("POST")
	router.HandleFunc("/download/{filename}", DownloadHandler(storageDir)).Methods("GET")

	// Start the server on the specified port.
	addr := ":" + port
	log.Printf("DropGox Backend is running on port %s...\n", port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

/// HealthCheckHandler provides a simple health-check endpoint.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{Message: "DropGox Backend is up and running!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UploadHandler returns a handler function for file uploads.
// It takes storageDir as an argument so we can save the uploaded files there.
func UploadHandler(storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Limit upload size to 10 MB.
		r.Body = http.MaxBytesReader(w, r.Body, 1000<<20)

		// Parse the multipart form.
		if err := r.ParseMultipartForm(1000 << 20); err != nil {
			http.Error(w, "File too big or malformed form data", http.StatusBadRequest)
			return
		}

		// Retrieve the file from the form data.
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to retrieve file from form data", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Create the destination file.
		dstPath := filepath.Join(storageDir, header.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "Unable to create file on the server", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the destination file.
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Message: "File uploaded successfully"})
	}
}

// DownloadHandler returns a handler function for file downloads.
// It takes storageDir as an argument so we know where to look for files.
func DownloadHandler(storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filename := vars["filename"]
		filePath := filepath.Join(storageDir, filename)

		// Check if the file exists.
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Set headers to force download.
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		http.ServeFile(w, r, filePath)
	}
}
