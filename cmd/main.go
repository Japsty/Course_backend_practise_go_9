package main

import (
	"UNI/5sem/backend/9_pract/internal/config"
	"UNI/5sem/backend/9_pract/internal/file"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	router := chi.NewRouter()
	router.Get("/files", file.GetFilesHandler)
	router.Get("/files/{id}", file.GetFileHandler)
	router.Get("/files/{id}/info", file.GetFileInfoHandler)
	router.Post("/files", file.PostFile)
	router.Put("/files/{id}", file.UpdateFile)
	router.Delete("/files/{id}", file.DeleteFile)

	log.Printf("Server is running on port %s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
