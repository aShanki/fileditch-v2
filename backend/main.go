package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	db *mongo.Database
)

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongodb:27017/filehost"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db = client.Database("filehost")

	// Initialize storage service
	storage := newLocalStorageService()

	// Initialize file service
	fileService := newFileService(db, storage)

	// Start cleanup goroutine
	go fileService.cleanupExpiredFiles()

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(corsMiddleware())

	// File upload endpoint
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
			return
		}

		duration := c.PostForm("duration")
		password := c.PostForm("password")

		fileInfo, err := fileService.uploadFile(c, file, duration, password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, fileInfo)
	})

	// File download endpoint
	r.GET("/files/:id", func(c *gin.Context) {
		id := c.Param("id")
		password := c.Query("password")

		err := fileService.serveFile(c, id, password)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found or incorrect password"})
			return
		}
	})

	// List files endpoint
	r.GET("/files", func(c *gin.Context) {
		files, err := fileService.listFiles()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, files)
	})

	// Delete file endpoint
	r.DELETE("/files/:id", func(c *gin.Context) {
		id := c.Param("id")
		err := fileService.deleteFile(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}