// CMPT 315 (Fall 2020)
// Assignment 1
//
// Author: Ben Ha

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// post represents the data stored in a post
type getPostsResponse struct {
	Creator string `db:"creator"`
	Title   string `db:"title"`
	Text    string `db:"text"`
}
type createPostRequest struct {
	Creator string `json:"creator"`
	Title   string `json:"title"`
	Text    string `json:"text"`
	Public  bool   `json:"public"`
}

var db *sqlx.DB

func main() {
	// Connect to Database
	var err error
	db, err = connectToDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to database: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Connected to database.")
	}

	// Connect to router
	r := mux.NewRouter()
	httpPort := 8123

	// Create Routes
	r.Path("/api/v1/posts").Methods("GET").HandlerFunc(getPosts)
	r.Path("/api/v1/posts").Methods("POST").HandlerFunc(createPost)

	fmt.Printf("listen to port %v...\n", httpPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r))
}

func connectToDB() (*sqlx.DB, error) {
	connectionString := "assign1.db?mode=column&_fk=true"
	database, err := sqlx.Open("sqlite3", connectionString)
	return database, err
}

func login() string {
	var username string

	fmt.Printf("Enter username: ")
	fmt.Scanln(&username)

	query := `SELECT username FROM users WHERE username=$1;`
	var result string
	err := db.Get(&result, query, username)

	if err == nil {
		fmt.Println("Login successful")
		return username
	} else {
		var userInput string
		fmt.Println("Username not found.")
		fmt.Print("Would you like to create a new account with this username? (y/n) ")
		fmt.Scanln(&userInput)
		if strings.EqualFold("y", userInput) {
			query := `INSERT INTO users (username) VALUES ($1);`
			_, err := db.Exec(query, username)

			if err == nil {
				fmt.Println("New account created")
				return username
			} else {
				fmt.Printf("Cannot create new account: %v\n", err)
				os.Exit(2)
			}
		}
		os.Exit(3)
	}
	return ""
}

// Get all public posts
// Allows for filtering and sorting by creator and title, and pagination
func getPosts(w http.ResponseWriter, r *http.Request) {
	posts := []getPostsResponse{}

	queryString := "Select creator, title, text FROM posts WHERE public = 1"
	args := []interface{}{}

	// Get query parameters
	urlQuery := r.URL.Query()
	// Filtering
	if creator, exists := urlQuery["creator"]; exists {
		queryString = fmt.Sprintf("%s AND creator = ?", queryString)
		args = append(args, creator[0])
	}
	if title, exists := urlQuery["title"]; exists {
		queryString = fmt.Sprintf("%s AND title LIKE ?", queryString)
		args = append(args, fmt.Sprintf("%%%s%%", title[0]))
	}
	// Sorting
	if sort, exists := urlQuery["sort"]; exists {
		queryString += fmt.Sprintf(" ORDER BY %s", sort[0][1:])
		if sort[0][0:1] == "-" {
			queryString += " DESC"
		}
	}
	// Pagination
	currentPage := 1
	pageLimit := 20

	if page, exists := urlQuery["page"]; exists {
		if pageInt, err := strconv.Atoi(page[0]); err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), 400)
		} else {
			currentPage = pageInt
		}
	}

	// Execute query and store resulted rows in posts
	queryString += ";"
	fmt.Println(queryString)

	err := db.Select(&posts, queryString, args...)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		// Configure response
		w.Header().Set("Content-type", "application/json")
		envelope := map[string]interface{}{
			"metadata": map[string]interface{}{
				"total count": len(posts),
				"total pages": int(math.Ceil(float64(len(posts)) / float64(pageLimit))),
				"page":        currentPage,
				"per page":    pageLimit,
				"links":       map[string]interface{}{},
			},
			"posts": posts,
		}

		// Encode posts into JSON
		json.NewEncoder(w).Encode(envelope)
	}

}

func createPost(w http.ResponseWriter, r *http.Request) {
	var newPost createPostRequest

	// Decode the request's body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newPost)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Insert new post into db
	queryString := `INSERT INTO posts (creator, title, text, public) VALUES ($1, $2, $3, $4);`
	result, err := db.Exec(queryString, newPost.Creator, newPost.Title, newPost.Text, newPost.Public)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%d row(s) created.\n", rowsAffected)
}
