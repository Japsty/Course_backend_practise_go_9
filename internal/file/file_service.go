package file

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"io"
	"log"
	"mime/multipart"
)

func FindInGrid(ctx context.Context, bucket *gridfs.Bucket) (*mongo.Cursor, error) {
	cursor, err := bucket.Find(ctx, nil)
	if err != nil {
		log.Printf("Failed to find files in GridFS: %v", err)
		return nil, err
	}

	return cursor, nil
}

func UploadFile(bucket *gridfs.Bucket, file *multipart.FileHeader) (primitive.ObjectID, error) {
	uploadStream, err := bucket.OpenUploadStream(file.Filename)
	if err != nil {
		log.Printf("Error uploading file data: %v", err)
		return primitive.NilObjectID, err
	}
	defer uploadStream.Close()

	srcFile, err := file.Open()
	if err != nil {
		log.Printf("Error opening source file: %v", err)
		return primitive.NilObjectID, err
	}
	defer srcFile.Close()

	_, err = io.Copy(uploadStream, srcFile)
	if err != nil {
		log.Printf("Error copying file data: %v", err)
		return primitive.NilObjectID, err
	}
	objectID := uploadStream.FileID.(primitive.ObjectID)
	return objectID, nil
}
