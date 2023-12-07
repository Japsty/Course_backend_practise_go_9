package file

import (
	"UNI/5sem/backend/9_pract/internal/db"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"io"
	"log"
	"net/http"
)

type FileInfo struct {
	FileID   primitive.ObjectID `json:"fileID"`
	Filename string             `json:"filename"`
	Length   int64              `json:"length"`
}

func GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	client, err := db.ConnectToMongo()
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.DisconnectMongo(context.Background(), client)

	datab := client.Database("MONGODB")
	bucket := db.NewBucket(datab)

	cursor, err := FindInGrid(context.Background(), bucket)
	if err != nil {
		log.Printf("Error finding files in GridFS: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var filesInfo []FileInfo

	for cursor.Next(context.Background()) {
		var file gridfs.File
		if err := cursor.Decode(&file); err != nil {
			log.Printf("Error decoding file: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fileInfo := FileInfo{
			FileID:   file.ID.(primitive.ObjectID),
			Filename: file.Name,
			Length:   file.Length,
		}
		filesInfo = append(filesInfo, fileInfo)
	}

	responseJSON, err := json.MarshalIndent(filesInfo, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func GetFileHandler(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")
	fieldId, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Printf("Error parsing file ID: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	client, err := db.ConnectToMongo()
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.DisconnectMongo(context.Background(), client)

	database := client.Database("MONGODB")
	bucket := db.NewBucket(database)

	file, err := bucket.OpenDownloadStream(fieldId)
	if err != nil {
		log.Printf("Error finding file in GridFS: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("Error writing file content: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func GetFileInfoHandler(w http.ResponseWriter, r *http.Request) {
	fileID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Error parsing file ID: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	client, err := db.ConnectToMongo()
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.DisconnectMongo(context.Background(), client)

	database := client.Database("MONGODB")
	bucket := db.NewBucket(database)

	filter := bson.D{{"_id", fileID}}
	fileCursor, err := bucket.FindContext(context.Background(), filter)
	if err != nil {
		log.Printf("Error finding file in GridFS: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer fileCursor.Close(context.Background())

	var fileInfo FileInfo
	if fileCursor.Next(context.Background()) {
		if err := fileCursor.Decode(&fileInfo); err != nil {
			log.Printf("Error decoding file info: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fileInfo = FileInfo{
			FileID:   fileInfo.FileID,
			Filename: fileInfo.Filename,
			Length:   fileInfo.Length,
		}
	} else {
		log.Printf("File not found in GridFS")
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	responseJSON, err := json.Marshal(fileInfo)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func PostFile(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	client, err := db.ConnectToMongo()
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.DisconnectMongo(context.Background(), client)
	datab := client.Database("MONGODB")
	bucket := db.NewBucket(datab)

	ID, err := UploadFile(bucket, handler)
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("File uploaded: %v", ID)
	JSONID, _ := ID.MarshalJSON()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded:"))
	w.Write(JSONID)
}

func UpdateFile(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")
	fileID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		log.Printf("Error parsing file ID: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	client, err := db.ConnectToMongo()
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.DisconnectMongo(context.Background(), client)

	database := client.Database("MONGODB")
	bucket := db.NewBucket(database)

	err = bucket.Delete(fileID)
	if err != nil {
		log.Printf("Error deleting file from GridFS: %v", err)
		w.Write([]byte("File not found\n"))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	newFile, newHandler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get new file from request", http.StatusBadRequest)
		return
	}
	defer newFile.Close()

	newUploadStream, err := bucket.OpenUploadStreamWithID(fileID, newHandler.Filename)
	if err != nil {
		log.Printf("Error opening new upload stream: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer newUploadStream.Close()

	_, err = io.Copy(newUploadStream, newFile)
	if err != nil {
		log.Printf("Error copying file data: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File updated successfully"))
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")
	fieldId, _ := primitive.ObjectIDFromHex(ID)

	client, err := db.ConnectToMongo()
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.DisconnectMongo(context.Background(), client)

	datab := client.Database("MONGODB")
	bucket := db.NewBucket(datab)

	err = bucket.Delete(fieldId)
	if err != nil {
		log.Printf("Error deleting file from GridFS: %v", err)
		w.Write([]byte("File not found\n"))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("File deleted"))
	w.WriteHeader(http.StatusOK)
}
