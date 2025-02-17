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
	"github.com/michael-oancea/DropGox/utils/auth"
	"github.com/gorilla/mux"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "port to run the server on")
	flag.StringVar(&port, "p", "8080", "port to run the server on (shorthand)")
	flag.Parse()

	storageDir := "./files"
	if err := os.MkdirAll(storageDir, os.ModePerm); err != nil {
		log.Fatal("Error creating storage directory:", err)
	}

	router := mux.NewRouter()

	// Health-check endpoint doesn't require authentication.
	router.HandleFunc("/health", HealthCheckHandler).Methods("GET")

	// Secure file endpoints using the Keycloak middleware.
	secure := router.PathPrefix("/").Subrouter()
	secure.Use(auth.KeycloakMiddleware)
	secure.HandleFunc("/upload", UploadHandler(storageDir)).Methods("POST")
	secure.HandleFunc("/download/{filename}", DownloadHandler(storageDir)).Methods("GET")

	addr := ":" + port
	log.Printf("DropGox Backend is running on port %s...\n", port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{Message: "DropGox Backend is up and running!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UploadHandler(storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File too big or malformed form data", http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to retrieve file from form data", http.StatusBadRequest)
			return
		}
		defer file.Close()

		dstPath := filepath.Join(storageDir, header.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "Unable to create file on the server", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Message: "File uploaded successfully"})
	}
}

func DownloadHandler(storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filename := vars["filename"]
		filePath := filepath.Join(storageDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		http.ServeFile(w, r, filePath)
	}
}
