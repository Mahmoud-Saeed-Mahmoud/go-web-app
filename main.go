package main

import (
	"log"
	"net/http"

	docs "go-web-app/docs"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// User represents a user in the system
type User struct {
	ID       int    `json:"id" db:"id"`
	UserName string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

// @Summary Get all users
// @Description Retrieve a list of users
// @Tags users
// @Accept  json
// @Produce  json
// @Success 200 {array} User
// @Router /users [get]
func getUsers(c *gin.Context) {
	// Get the database connection from context
	db := c.MustGet("db").(*sqlx.DB)

	var users []User
	err := db.Select(&users, "SELECT id, username, password FROM users")
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Debugging output
	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No users found"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// @Summary Create a new user
// @Description Create a new user in the system
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body User true "User Data"
// @Success 201 {object} User
// @Router /users [post]
func createUser(c *gin.Context) {
	var newUser User
	// Bind the incoming JSON to the User struct
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert the new user into the database
	// Ensure we are passing string parameters correctly into the query
	err := c.MustGet("db").(*sqlx.DB).QueryRowx(`
		INSERT INTO users (username, password) 
		VALUES ($1, $2) 
		RETURNING id`, newUser.UserName, newUser.Password).
		Scan(&newUser.ID) // Get the generated ID after insert

	// Check for any error during insertion
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with the newly created user
	c.JSON(http.StatusCreated, newUser)
}

// @title Go Web API
// @version 1.0
// @description This is a simple Go API with Gin and Swagger
// @host localhost:3333
func main() {
	// Connect to the PostgreSQL database
	db, err := sqlx.Connect("postgres", "user=postgres dbname=empty sslmode=disable password=123456 host=localhost")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	// Test the connection to the database
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully connected to the database")
	}

	// Create a new Gin router
	r := gin.Default()

	// Store the database connection in the Gin context
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	docs.SwaggerInfo.BasePath = ""

	// Setup routes
	r.GET("/users", getUsers)
	r.POST("/users", createUser)

	// Setup Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Run the server
	r.Run(":3333") // Listen on port 3333
}
