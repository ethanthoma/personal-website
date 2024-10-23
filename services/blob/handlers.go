package main

import (
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	uploadLimit = 10 << 20
)

func handlerDelete(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/files/")
	if filename == "" {
		http.Error(w, "filename not specified", http.StatusBadRequest)
		return
	}

	if err := FileDelete(filename); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to delete file", http.StatusInternalServerError)
			log.Printf("blob-service: error deleting file (%v)", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("file deleted successfully"))
}

func handlerDownload(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/files/")
	if filename == "" {
		http.Error(w, "filename not specified", http.StatusBadRequest)
		return
	}

	file, err := FileGet(filename)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to retrieve file", http.StatusInternalServerError)
			log.Printf("blob-service: error retrieving file (%v)", err)
		}
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Content-Type", contentType)

	if _, err = io.Copy(w, file); err != nil {
		http.Error(w, "failed to send file", http.StatusInternalServerError)
		log.Printf("blob-service: error sending file (%v)", err)
	}
}

func handlerUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)

	filename := strings.TrimPrefix(r.URL.Path, "/files/")
	if filename == "" {
		http.Error(w, "filename not specified", http.StatusBadRequest)
		return
	}

	err := r.ParseMultipartForm(uploadLimit)
	if err != nil {
		http.Error(w, "failed to parse form data", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err = FileSave(filename, file); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		log.Printf("blob-service: error saving file (%v)", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("file uploaded successfully"))
}

func handlerList(w http.ResponseWriter, r *http.Request) {
	files, err := FilesList()
	if err != nil {
		http.Error(w, "failed to list files", http.StatusInternalServerError)
		log.Printf("blob-service: error listing files (%v)", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
