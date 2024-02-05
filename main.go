package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// _ = sqlx.MustConnect("mysql", os.Getenv("DATABASE_URL"))
func Database() gin.HandlerFunc {
	db := sqlx.MustConnect("mysql", "root:fidele95@tcp(localhost:3306)/goweb")
	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}

// data definitions
type userjson struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func insertUser(db *sqlx.DB, userData userjson) error {
	_, err := db.Exec("INSERT INTO users (username, email) VALUES (?, ?)", userData.Username, userData.Email)
	return err
}
func updateUser(db *sqlx.DB, userData userjson, id int) error {
	_, err := db.Exec("UPDATE users SET username = ?, email = ? WHERE id = ?", userData.Username, userData.Email, id)
	return err
}
func deleteUser(db *sqlx.DB, id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func getUserByEmailDb(db *sqlx.DB, email string) (User, error) {
	userd := User{}
	err := db.Get(&userd, "SELECT id, username, email FROM users WHERE email = ?", email)
	// if err != nil {
	// 	return nil, err
	// }
	return userd, err
}

func getUserByIdDb(db *sqlx.DB, id int) (User, error) {
	userd := User{}
	err := db.Get(&userd, "SELECT id, username, email FROM users WHERE id = ?", id)
	return userd, err
}

func getUsersDb(db *sqlx.DB) ([]User, error) {
	users := []User{}
	err := db.Select(&users, "SELECT id, username, email FROM users")
	if err != nil {
		return nil, err
	}
	return users, nil
}

type User struct {
	Id       int
	Username string
	Email    string
}

// Users slice to seed record User data.
// var Users = []User{
// 	{
// 		ID:   "1",
// 		Name: "ABC",
// 	},
// 	{
// 		ID:   "2",
// 		Name: "DEF",
// 	},
// 	{
// 		ID:   "3",
// 		Name: "GHI",
// 	},
// }

func main() {
	// a gin router to handle requests
	var router *gin.Engine = gin.Default()
	// insert request handlers
	router.Use(Database())
	router.POST("/users", postUsers)
	router.GET("/users", getUsers)
	router.GET("/users/:id", getUserByID)
	router.GET("/users/email/:email", getUserByEmail)
	router.PUT("/users/:id", updateUsers)
	router.DELETE("/users/:id", deleteUserByID)
	// Listen at http://localhost:8080
	router.Run(":8080")
	fmt.Println("Listen at http://localhost:8080")
}

func getUsers(context *gin.Context) {
	// IndentedJSON makes it look better
	db := context.MustGet("DB").(*sqlx.DB)
	users, err := getUsersDb(db)
	if err != nil {
		fmt.Println(err)
		context.IndentedJSON(http.StatusInternalServerError, gin.H{
			"message": "error has occured",
		})
		return
	}
	context.IndentedJSON(http.StatusOK, users)
}

func postUsers(context *gin.Context) {
	db := context.MustGet("DB").(*sqlx.DB)
	var newUser userjson
	// BindJSON to bind the received JSON to newUser
	if err := context.BindJSON(&newUser); err != nil {
		// log the error, respond and return
		fmt.Println(err)
		context.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request",
		})
		return
	}
	err := insertUser(db, newUser)
	if err != nil {
		fmt.Println(err)
		context.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "error has occured",
			"error":   err.Error(),
		})
		return
	}
	// respond as IndentedJSON
	context.IndentedJSON(http.StatusCreated, newUser)
}
func updateUsers(context *gin.Context) {
	db := context.MustGet("DB").(*sqlx.DB)
	var sid string = context.Param("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
		context.IndentedJSON(
			http.StatusBadRequest,
			// refer https://pkg.go.dev/github.com/gin-gonic/gin#H
			gin.H{
				"message": "bad request",
			},
		)
		return
	}
	var userData userjson
	// BindJSON to bind the received JSON to userData
	if err := context.BindJSON(&userData); err != nil {
		// log the error, respond and return
		fmt.Println(err)
		context.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request",
		})
		return
	}
	updateUser(db, userData, id)
	if err != nil {
		fmt.Println(err)
		context.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "error has occured",
			"error":   err.Error(),
		})
		return
	}
	// respond as IndentedJSON
	context.IndentedJSON(http.StatusOK, userData)
}

func getUserByID(context *gin.Context) {
	// get the id from request params
	db := context.MustGet("DB").(*sqlx.DB)
	var sid string = context.Param("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
		context.IndentedJSON(
			http.StatusBadRequest,
			// refer https://pkg.go.dev/github.com/gin-gonic/gin#H
			gin.H{
				"message": "bad request",
			},
		)
		return
	}

	userd, err := getUserByIdDb(db, id)
	if err != nil {
		fmt.Println(err)
		context.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "error has occured",
			"error":   err.Error(),
		})
		return
	}

	// respond 404
	if (User{}) != userd {
		context.IndentedJSON(http.StatusOK, userd)
	} else {
		context.IndentedJSON(
			http.StatusNotFound,
			// refer https://pkg.go.dev/github.com/gin-gonic/gin#H
			gin.H{
				"message": "User not found",
			},
		)
	}
}

func getUserByEmail(context *gin.Context) {
	// get the id from request params
	db := context.MustGet("DB").(*sqlx.DB)
	var email string = context.Param("email")

	userd, err := getUserByEmailDb(db, email)
	if err != nil {
		fmt.Println(err)
		context.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "error has occured",
			"error":   err.Error(),
		})
		return
	}

	// respond 404
	if (User{}) != userd {
		context.IndentedJSON(http.StatusOK, userd)
	} else {
		context.IndentedJSON(
			http.StatusNotFound,
			// refer https://pkg.go.dev/github.com/gin-gonic/gin#H
			gin.H{
				"message": "User not found",
			},
		)
	}
}
func deleteUserByID(context *gin.Context) {
	// get the id from request params
	db := context.MustGet("DB").(*sqlx.DB)
	var sid string = context.Param("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
		context.IndentedJSON(
			http.StatusBadRequest,
			// refer https://pkg.go.dev/github.com/gin-gonic/gin#H
			gin.H{
				"message": "bad request",
			},
		)
		return
	}
	deleteUser(db, id)
	context.IndentedJSON(http.StatusOK, gin.H{
		"message": "deleted successfully",
	})
}
