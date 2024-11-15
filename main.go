package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`   // primaryKey makes this a unique id
	Name      string    `json:"name"`                   // name of user
	Username  string    `json:"username" gorm:"unique"` // userame of user
	Email     string    `json:"email" gorm:"unique"`    // email of user
	Password  []byte    `json:"-"`                      // to hide password in JSON reponse
	CreatedAt time.Time `json:"createdAt"`
}

var db *gorm.DB
var err error

func main() {
	router := gin.Default()
	//intializing database

	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to open database")
	}
	db.AutoMigrate(&User{}) //auto-migrate User struct to update/create the database schema

	// Public routes
	router.POST("/signup", signup)
	router.POST("/login", login)
	//routes
	router.GET("/users", getUsers)
	router.POST("/users", createUser)
	router.GET("/users/:id", getUserByID)
	router.PUT("/users/:id", updateUser)
	router.DELETE("/users/:id", deleteUser)
	// run server
	router.Run("localhost:8080")
}

// func for to read users
func getUsers(c *gin.Context) {
	var users []User
	db.Find(&users)
	c.JSON(http.StatusOK, users)
}

// func to create new users

func createUser(c *gin.Context) {
	var newUser User
	if err := c.BindJSON(&newUser); err != nil { // to bind JSON input to &newUser
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

// read users by id

func getUserByID(c *gin.Context) {
	var user User
	if err := db.First(&user, "id=?", c.Param("id")).Error; err != nil { // to get user via id
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user) // return user detail as JSON
}

// func for updating user

func updateUser(c *gin.Context) {
	var user User
	if err := db.Find(&user, "id=?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err := c.BindJSON(&user); err != nil { // bind json input to update user fields
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Save(&user) //save updates to database
	c.JSON(http.StatusOK, user)
}

// delete user

func deleteUser(c *gin.Context) {
	var user User
	if err := db.First(&user, "id = ?", c.Param("id")).Error; err != nil { // Find user by ID
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	db.Delete(&user) // Delete user from the database
	c.JSON(http.StatusNoContent, gin.H{})
}

// sign up-func

func signup(c *gin.Context) {
	var newUser User
	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// To hass the password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	newUser.Password = hashedPass
	newUser.CreatedAt = time.Now()

	if err := db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create new user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user created"})
}

func login(c *gin.Context) {
	var inputUser User
	var storedUser User

	if err := c.BindJSON(&inputUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// find user by username
	if err := db.Where("username = ?", inputUser.Username).First(&storedUser).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	if err := bcrypt.CompareHashAndPassword(storedUser.Password, []byte(inputUser.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

}
