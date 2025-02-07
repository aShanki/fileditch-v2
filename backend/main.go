package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	db          *mongo.Database
	userService *UserService
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if (authHeader == "") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["user_id"])
		c.Set("username", claims["username"])
		c.Set("is_admin", claims["is_admin"])
		c.Next()
	}
}

func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

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

	// Initialize services
	storage := newLocalStorageService()
	fileService := newFileService(db, storage)
	userService = newUserService(db)

	// Create default admin user if it doesn't exist
	_, err = userService.CreateUser("admin", "admin123", true)
	if err != nil && err.Error() != "username already exists" {
		log.Printf("Error creating default admin: %v", err)
	}

	// Start cleanup goroutine
	go fileService.cleanupExpiredFiles()

	// Setup Gin router
	r := gin.Default()

	// CORS middleware - Apply it before any routes
	r.Use(corsMiddleware())

	// Group all routes under /api
	api := r.Group("/api")
	{
		// Public routes
		api.POST("/login", handleLogin)

		// Protected routes
		auth := api.Group("/")
		auth.Use(authMiddleware())
		{
			// File management
			auth.POST("/upload", func(c *gin.Context) {
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

			auth.GET("/files", func(c *gin.Context) {
				files, err := fileService.listFiles()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, files)
			})

			auth.DELETE("/files/:id", func(c *gin.Context) {
				id := c.Param("id")
				err := fileService.deleteFile(id)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.Status(http.StatusOK)
			})

			// Admin routes
			admin := auth.Group("/admin")
			admin.Use(adminMiddleware())
			{
				admin.GET("/users", handleListUsers)
				admin.POST("/users", handleCreateUser)
				admin.PUT("/users/:id", handleUpdateUser)
				admin.DELETE("/users/:id", handleDeleteUser)
				admin.PUT("/users/:id/password", handleChangePassword)
			}
		}

		// File download endpoint remains public but under /api
		api.GET("/files/:id", func(c *gin.Context) {
			id := c.Param("id")
			password := c.Query("password")

			err := fileService.serveFile(c, id, password)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "File not found or incorrect password"})
				return
			}
		})
	}

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func handleLogin(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := userService.AuthenticateUser(credentials.Username, credentials.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    tokenString,
		"username": user.Username,
		"isAdmin":  user.IsAdmin,
	})
}

func handleListUsers(c *gin.Context) {
	users, err := userService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func handleCreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"isAdmin"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := userService.CreateUser(req.Username, req.Password, req.IsAdmin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func handleUpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Username string `json:"username"`
		IsAdmin  bool   `json:"isAdmin"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := userService.UpdateUser(id, req.Username, req.IsAdmin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func handleDeleteUser(c *gin.Context) {
	id := c.Param("id")
	err := userService.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func handleChangePassword(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := userService.ChangePassword(id, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}