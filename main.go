// CMPT 315 (Fall 2020)
// Assignment 1
//
// Author: Ben Ha

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// post represents the data stored in a post
type post struct {
	ID       int    `db:"-" json:"-"`
	Title    string `db:"title" json:"title,omitempty"`
	Text     string `db:"text,omitempty" json:"text,omitempty"`
	Public   *bool  `db:"public,omitempty" json:"public,omitempty"`
	ReadID   int    `db:"read_id" json:"read_id,omitempty"`
	WriteID  int    `db:"write_id,omitempty" json:"write_id,omitempty"`
	Reported bool   `db:"-" json:"-"`
}

var db *sqlx.DB

func main() {
	// Generate seed for randomizing numbers
	rand.Seed(time.Now().UTC().UnixNano())

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
	r.Path("/api/v1/posts/{id}").Methods("GET").HandlerFunc(getPostByID)
	r.Path("/api/v1/posts/{id}/reports").Methods("POST").HandlerFunc(reportPost)
	r.Path("/api/v1/posts/{id}").Methods("PUT").HandlerFunc(updatePost)
	r.Path("/api/v1/posts/{id}").Methods("DELETE").HandlerFunc(deletePost)
	r.PathPrefix("/").HandlerFunc(catchAllHandlerFunc)

	fmt.Printf("listen to port %v...\n", httpPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r))
}

// Connect to database
func connectToDB() (*sqlx.DB, error) {
	connectionString := "assign.db?mode=column&_fk=true"
	database, err := sqlx.Open("sqlite3", connectionString)
	return database, err
}

// Log a request
func logRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.RequestURI)
	queryString := `INSERT INTO logs (method, uri) VALUES ($1, $2);`
	_, err := db.Exec(queryString, r.Method, r.RequestURI)

	if err != nil {
		fmt.Printf("Fail to log: %s\n", err.Error())
		return
	}
}

// Get all public posts
// Allow for filtering and sorting by creator and title, and pagination
func getPosts(w http.ResponseWriter, r *http.Request) {
	// Log request
	logRequest(w, r)

	posts := []post{}

	// Start building query string
	queryString := "Select title, read_id FROM posts WHERE public = 1 AND reported = 0"
	args := []interface{}{}

	// Get query parameters
	urlQuery := r.URL.Query()
	// Filtering by title
	if title, exists := urlQuery["title"]; exists {
		queryString = fmt.Sprintf("%s AND title LIKE ?", queryString)
		args = append(args, fmt.Sprintf("%%%s%%", title[0]))
	}
	// Sorting by title
	if sort, exists := urlQuery["sort"]; exists {
		if strings.EqualFold(sort[0], "title") {
			queryString += fmt.Sprintf(" ORDER BY title")
		} else if strings.EqualFold(sort[0], "title_desc") {
			queryString += fmt.Sprintf(" ORDER BY title DESC")
		} else {
			writeJSONResponse(w, fmt.Sprintf("Invalid query value: %v", sort[0]), 400)
			return
		}
	}
	// Pagination
	currentPage := 1
	pageLimit := 20

	if page, exists := urlQuery["page"]; exists {
		if pageInt, err := strconv.Atoi(page[0]); err != nil {
			writeJSONResponse(w, fmt.Sprintf("Invalid query value: %v", page[0]), 400)
			return
		} else {
			currentPage = pageInt
		}
	}

	// Execute query and store resulted rows in posts
	queryString += ";"

	err := db.Select(&posts, queryString, args...)
	if err != nil {
		writeJSONResponse(w, http.StatusText(500), 500)
		return
	}
	if len(posts) == 0 {
		writeJSONResponse(w, "No results found", 200)
		return
	}

	// Get specified page from posts
	var pagePosts []post
	firstOfPage := (currentPage - 1) * pageLimit // e.g. Page 1 starts from index 0
	lastOfPage := currentPage * pageLimit        // e.g. Page 1 ends before index 20
	if firstOfPage >= len(posts) {               // If first of page is beyond posts' range, return empty array
		pagePosts = []post{}
	} else if lastOfPage > len(posts) { // If last of page is beyond posts' range, only take up to last item
		pagePosts = posts[firstOfPage:]
	} else {
		pagePosts = posts[firstOfPage:lastOfPage]
	}

	// Configure response
	w.Header().Set("Content-type", "application/json")
	response := map[string]interface{}{
		"message": fmt.Sprintf("Showing %d of %d post(s) found", len(pagePosts), len(posts)),
		"metadata": map[string]interface{}{
			"total_count": len(posts),
			"total_pages": int(math.Ceil(float64(len(posts)) / float64(pageLimit))),
			"page":        currentPage,
			"per_page":    pageLimit,
		},
		"results": pagePosts,
	}
	json.NewEncoder(w).Encode(response)

}

// Create a new post
// Take in title (string), text (string), and public (bool) from POST request
func createPost(w http.ResponseWriter, r *http.Request) {
	// Log request
	logRequest(w, r)

	var newPost post

	// Decode the request's body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newPost)

	if err != nil {
		writeJSONResponse(w, err.Error(), 400)
		return
	}

	// Create write & read IDs
	// First, get all the existing read and write IDs from db
	var existingIDs []int
	queryString := "SELECT read_id FROM posts UNION SELECT write_id FROM posts ORDER BY read_id;"
	err = db.Select(&existingIDs, queryString)
	if err != nil {
		writeJSONResponse(w, err.Error(), 500)
		return
	}
	// Then, randomize until we have two new unique IDs
	unique := false
	readID := 0
	writeID := 0
	n := len(existingIDs)
	for !unique {
		readID = rand.Intn(10000)
		writeID = rand.Intn(1000)
		// Check if readID = writeID
		if readID == writeID {
			continue
		}
		// Check if readID and writeID already exists
		readIndex := sort.Search(n, func(i int) bool { return existingIDs[i] >= readID })
		writeIndex := sort.Search(n, func(i int) bool { return existingIDs[i] >= writeID })

		if (readIndex < n && existingIDs[readIndex] == readID) || (writeIndex < n && existingIDs[writeIndex] == writeID) {
			continue
		}
		unique = true
	}

	// Insert new post into db
	queryString = `INSERT INTO posts (title, text, public, read_id, write_id) VALUES ($1, $2, $3, $4, $5);`
	result, err := db.Exec(queryString, newPost.Title, newPost.Text, newPost.Public, readID, writeID)

	if err != nil {
		writeJSONResponse(w, err.Error(), 500)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%d row(s) created.\n", rowsAffected)

	// Send the post back to client
	newPost.ReadID = readID
	newPost.WriteID = writeID
	w.Header().Set("Content-type", "application/json")
	response := map[string]interface{}{
		"message":      "New post created",
		"post_content": newPost,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		writeJSONResponse(w, err.Error(), 500)
	}
}

// Get a post by its read or write ID
// If it's a read --> Return text, title, and report link
// If it's a write --> Return text, title, public, read and write links, update and delete links
func getPostByID(w http.ResponseWriter, r *http.Request) {
	// Log request
	logRequest(w, r)

	// Get id from path variables
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		writeJSONResponse(w, fmt.Sprintf("Invalid post id: %v", mux.Vars(r)["id"]), 400)
		return
	}

	// Find id
	var result post
	queryString := `SELECT title, text, public, read_id, write_id FROM posts where read_id = $1 or write_id = $1;`
	err = db.Get(&result, queryString, id)

	if err == sql.ErrNoRows {
		writeJSONResponse(w, fmt.Sprintf("No post with id %d", id), http.StatusNotFound)
		return
	} else if err != nil {
		writeJSONResponse(w, err.Error(), 400)
		return
	}

	// If id is a read
	if result.ReadID == id {
		// Return title, text, and report link
		w.Header().Set("Content-type", "application/json")
		response := map[string]interface{}{
			"message": "Post found",
			"post_content": map[string]interface{}{
				"title": result.Title,
				"text":  result.Text,
			},
			"read_only_options": map[string]interface{}{
				"report_link": "/reports",
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			writeJSONResponse(w, err.Error(), 500)
		}
	} else { // If id is a write
		// Return title, text, public, read and write links, update and delete links
		w.Header().Set("Content-type", "application/json")
		response := map[string]interface{}{
			"message":      "Post found",
			"post_content": result,
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			writeJSONResponse(w, err.Error(), 500)
		}
	}

}

// Report a post
// Take in a reason (string), then store it in the reported post
func reportPost(w http.ResponseWriter, r *http.Request) {
	// Log request
	logRequest(w, r)

	// Get id from path variables
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		writeJSONResponse(w, fmt.Sprintf("Invalid id: %v\n", mux.Vars(r)["id"]), 400)
		return
	}

	// Check if id is valid
	exists, idType := checkPostID(w, id)
	if !exists {
		writeJSONResponse(w, fmt.Sprintf("No post with id %d", id), http.StatusNotFound)
		return
	}
	if idType == "w" {
		writeJSONResponse(w, "Invalid report uri", http.StatusBadRequest)
		return
	}

	// Decode the request's body to get report reason
	var reason struct {
		Reason string `json:"reason"`
	}
	err = json.NewDecoder(r.Body).Decode(&reason)

	if err != nil {
		writeJSONResponse(w, err.Error(), 400)
		return
	}

	// Set post's reported value to true
	queryString1 := `UPDATE posts SET reported = 1 WHERE read_id = $1;`
	_, err = db.Exec(queryString1, id)

	if err != nil {
		writeJSONResponse(w, err.Error(), 500)
		return
	}

	// Create a report in the reports table
	queryString2 := `
	INSERT INTO reports (post_id, reason)
	SELECT id, $1 FROM posts WHERE read_id = $2;`
	fmt.Println(reason.Reason)
	fmt.Println(id)
	_, err = db.Exec(queryString2, reason.Reason, id)

	if err != nil {
		writeJSONResponse(w, err.Error(), 500)
		return
	}
}

// Update a Post
// Take in new title (string), text (string), and public (bool)
func updatePost(w http.ResponseWriter, r *http.Request) {
	// Log request
	logRequest(w, r)

	// Get id from path variables
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		writeJSONResponse(w, fmt.Sprintf("Invalid id: %v\n", mux.Vars(r)["id"]), 400)
		return
	}

	// Check if id is valid
	exists, idType := checkPostID(w, id)
	if !exists {
		writeJSONResponse(w, fmt.Sprintf("No post with id %d", id), http.StatusNotFound)
		return
	}
	if idType == "r" {
		writeJSONResponse(w, "Not authorized", http.StatusForbidden)
		return
	}

	// Decode the request's body to get update info
	var updatedPost post
	err = json.NewDecoder(r.Body).Decode(&updatedPost)

	if err != nil {
		writeJSONResponse(w, err.Error(), 400)
		return
	}

	// Build query string
	queryString := `UPDATE posts SET id = id`
	args := []interface{}{}
	if updatedPost.Title != "" {
		queryString += ", title = ?"
		args = append(args, updatedPost.Title)
	}
	if updatedPost.Text != "" {
		queryString += ", text = ?"
		args = append(args, updatedPost.Text)
	}
	if updatedPost.Public != nil {
		queryString += ", public = ?"
		args = append(args, updatedPost.Public)
	}
	queryString += " WHERE write_id = ?"
	args = append(args, id)

	// Execute query to update post
	result, err := db.Exec(queryString, args...)
	if err != nil {
		writeJSONResponse(w, err.Error(), 500)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%d row(s) updated.\n", rowsAffected)

	// Send the updated post back to client
	queryString = `SELECT title, text, public, read_id, write_id FROM posts WHERE write_id = ?`
	err = db.Get(&updatedPost, queryString, id)
	if err != nil {
		writeJSONResponse(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-type", "application/json")
	response := map[string]interface{}{
		"message":      "Post updated",
		"post_content": updatedPost,
		"admin_options": map[string]interface{}{
			"update_link": "/",
			"delete_link": "/",
		},
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		writeJSONResponse(w, err.Error(), 500)
		return
	}
}

// Delete a Post
func deletePost(w http.ResponseWriter, r *http.Request) {
	// Log request
	logRequest(w, r)

	// Get id from path variables
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		writeJSONResponse(w, fmt.Sprintf("Invalid id: %v\n", mux.Vars(r)["id"]), 400)
		return
	}

	// Check if id is valid
	exists, idType := checkPostID(w, id)
	if !exists {
		writeJSONResponse(w, fmt.Sprintf("No post with id %d", id), http.StatusNotFound)
		return
	}
	if idType == "r" {
		writeJSONResponse(w, "Not authorized", http.StatusForbidden)
		return
	}

	// Execute query to delete post
	queryString := `DELETE FROM posts WHERE write_id = $1`
	result, err := db.Exec(queryString, id)

	if err != nil {
		writeJSONResponse(w, err.Error(), 500)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%d row(s) deleted.\n", rowsAffected)

	// Send a response body
	w.Header().Set("Content-type", "application/json")
	response := map[string]interface{}{
		"message": "Post deleted",
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		writeJSONResponse(w, err.Error(), 500)
		return
	}
}

// Any request not supported will return a 404
func catchAllHandlerFunc(w http.ResponseWriter, r *http.Request) {
	// Log request
	logRequest(w, r)

	writeJSONResponse(w, "Not Found", http.StatusNotFound)
	return
}

// Check to see if a post exists
// Returns a bool, and a string indicating if the id is a read or write
func checkPostID(w http.ResponseWriter, id int) (bool, string) {
	queryString := `SELECT read_id FROM posts where read_id = $1 or write_id = $1;`

	var readResult int

	err := db.Get(&readResult, queryString, id)

	if err == sql.ErrNoRows {
		// No post with this Id
		return false, ""
	} else if err != nil {
		writeJSONResponse(w, err.Error(), 400)
		return false, ""
	} else if readResult == id {
		// This id is a read
		return true, "r"
	} else {
		// This id is a write
		return true, "w"
	}

}

// Write a response body with a status code
func writeJSONResponse(w http.ResponseWriter, message string, code int) {
	fmt.Println(message)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(code)
	response := map[string]interface{}{
		"message": message,
	}
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println(err)
	}
}
