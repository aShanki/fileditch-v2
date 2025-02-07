package main

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type FileInfo struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string            `bson:"name" json:"name"`
	Size        int64             `bson:"size" json:"size"`
	ContentType string            `bson:"contentType" json:"contentType"`
	UploadDate  time.Time         `bson:"uploadDate" json:"uploadDate"`
	ExpiryDate  *time.Time        `bson:"expiryDate,omitempty" json:"expiryDate,omitempty"`
	Password    string            `bson:"password,omitempty" json:"-"`
}

type FileService struct {
	db      *mongo.Database
	storage StorageService
}

func newFileService(db *mongo.Database, storage StorageService) *FileService {
	return &FileService{
		db:      db,
		storage: storage,
	}
}

func (s *FileService) uploadFile(c *gin.Context, file *multipart.FileHeader, duration string, password string) (*FileInfo, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Generate a new ObjectID for the file
	id := primitive.NewObjectID()

	// Calculate expiry date
	var expiryDate *time.Time
	if duration != "permanent" {
		expiry := calculateExpiryDate(duration)
		expiryDate = &expiry
	}

	// Hash password if provided
	var hashedPassword string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hashedPassword = string(hash)
	}

	// Create file info
	fileInfo := &FileInfo{
		ID:          id,
		Name:        file.Filename,
		Size:        file.Size,
		ContentType: file.Header.Get("Content-Type"),
		UploadDate:  time.Now(),
		ExpiryDate:  expiryDate,
		Password:    hashedPassword,
	}

	// Store file in storage service
	if err := s.storage.Store(id.Hex(), src); err != nil {
		return nil, err
	}

	// Save file info to database
	_, err = s.db.Collection("files").InsertOne(context.Background(), fileInfo)
	if err != nil {
		// Cleanup stored file if database insert fails
		s.storage.Delete(id.Hex())
		return nil, err
	}

	return fileInfo, nil
}

func (s *FileService) serveFile(c *gin.Context, id string, password string) error {
	// Convert string ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Get file info from database
	var fileInfo FileInfo
	err = s.db.Collection("files").FindOne(context.Background(), bson.M{"_id": objID}).Decode(&fileInfo)
	if err != nil {
		return err
	}

	// Check if file has expired
	if fileInfo.ExpiryDate != nil && time.Now().After(*fileInfo.ExpiryDate) {
		s.deleteFile(id)
		return fmt.Errorf("file has expired")
	}

	// Check password if file is password protected
	if fileInfo.Password != "" {
		if password == "" {
			return fmt.Errorf("password required")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(fileInfo.Password), []byte(password)); err != nil {
			return fmt.Errorf("incorrect password")
		}
	}

	// Get file from storage
	reader, err := s.storage.Get(id)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Set response headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name))
	c.Header("Content-Type", fileInfo.ContentType)

	// Stream file to response
	_, err = io.Copy(c.Writer, reader)
	return err
}

func (s *FileService) listFiles() ([]FileInfo, error) {
	cursor, err := s.db.Collection("files").Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var files []FileInfo
	if err := cursor.All(context.Background(), &files); err != nil {
		return nil, err
	}

	return files, nil
}

func (s *FileService) deleteFile(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Delete from database
	_, err = s.db.Collection("files").DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		return err
	}

	// Delete from storage
	return s.storage.Delete(id)
}

func (s *FileService) cleanupExpiredFiles() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		cursor, err := s.db.Collection("files").Find(context.Background(), bson.M{
			"expiryDate": bson.M{"$lt": time.Now()},
		})
		if err != nil {
			continue
		}

		var files []FileInfo
		if err := cursor.All(context.Background(), &files); err != nil {
			cursor.Close(context.Background())
			continue
		}

		for _, file := range files {
			s.deleteFile(file.ID.Hex())
		}

		cursor.Close(context.Background())
	}
}

func calculateExpiryDate(duration string) time.Time {
	now := time.Now()
	switch duration {
	case "1h":
		return now.Add(time.Hour)
	case "1d":
		return now.Add(24 * time.Hour)
	case "7d":
		return now.Add(7 * 24 * time.Hour)
	case "30d":
		return now.Add(30 * 24 * time.Hour)
	default:
		return now.Add(24 * time.Hour) // Default to 1 day
	}
}