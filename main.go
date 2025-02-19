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
	"time"

	"github.com/gorilla/mux"
	"github.com/michael-oancea/DropGox/utils"
)

type Response struct {
	Message string `json:"message"`
}

// FileMeta represents basic metadata about a file.
type FileMeta struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
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

	// Public endpoint: health-check.
	router.HandleFunc("/health", HealthCheckHandler).Methods("GET")

	// Secure endpoints using Keycloak middleware.
	secure := router.PathPrefix("/").Subrouter()
	secure.Use(auth.KeycloakMiddleware)
	secure.HandleFunc("/upload", UploadHandler(storageDir)).Methods("POST")
	secure.HandleFunc("/download/{filename}", DownloadHandler(storageDir)).Methods("GET")
	secure.HandleFunc("/delete/{filename}", DeleteHandler(storageDir)).Methods("DELETE")
	secure.HandleFunc("/rename/{oldName}", RenameHandler(storageDir)).Methods("PUT")
	secure.HandleFunc("/metadata/{filename}", MetadataHandler(storageDir)).Methods("GET")

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
		// Limit upload size to 10 MB.
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

func DeleteHandler(storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filename := vars["filename"]
		filePath := filepath.Join(storageDir, filename)
		if err := os.Remove(filePath); err != nil {
			http.Error(w, "Error deleting file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Message: "File deleted successfully"})
	}
}

func RenameHandler(storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		oldName := vars["oldName"]
		oldPath := filepath.Join(storageDir, oldName)

		// Expect the new filename to be provided as a query parameter, e.g., ?newName=newfilename.txt
		newName := r.URL.Query().Get("newName")
		if newName == "" {
			http.Error(w, "New file name not provided", http.StatusBadRequest)
			return
		}
		newPath := filepath.Join(storageDir, newName)

		if err := os.Rename(oldPath, newPath); err != nil {
			http.Error(w, "Error renaming file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Message: "File renamed successfully"})
	}
}

func MetadataHandler(storageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filename := vars["filename"]
		filePath := filepath.Join(storageDir, filename)
		info, err := os.Stat(filePath)
		if err != nil {
			http.Error(w, "Error retrieving file metadata: "+err.Error(), http.StatusInternalServerError)
			return
		}
		meta := FileMeta{
			Name:    info.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(meta)
	}
}
