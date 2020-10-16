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
	ID            int    `db:"-" json:"-"`
	Title         string `db:"title" json:"title"`
	Text          string `db:"text,omitempty" json:"text,omitempty"`
	Public        *bool  `db:"public,omitempty" json:"public,omitempty"`
	ReadID        int    `db:"read_id" json:"read_id,omitempty"`
	WriteID       int    `db:"write_id,omitempty" json:"write_id,omitempty"`
	Reported      bool   `db:"-" json:"-"`
	report_reason string `db:"-" json:"-"`
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
	r.Path("/api/v1/posts/{id:[0-9]+}").Methods("GET").HandlerFunc(getPostById)
	r.Path("/api/v1/posts/{id:[0-9]+}/report").Methods("POST").HandlerFunc(reportPost)

	fmt.Printf("listen to port %v...\n", httpPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r))
}

// Connect to database
func connectToDB() (*sqlx.DB, error) {
	connectionString := "assign1.db?mode=column&_fk=true"
	database, err := sqlx.Open("sqlite3", connectionString)
	return database, err
}

// Get all public posts
// Allow for filtering and sorting by creator and title, and pagination
func getPosts(w http.ResponseWriter, r *http.Request) {
	posts := []post{}

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
		fmt.Println(sort[0])
		if strings.EqualFold(sort[0], "title") || strings.EqualFold(sort[0], "title_desc") {
			queryString += fmt.Sprintf(" ORDER BY title")
			if sort[0] == "title_desc" {
				queryString += " DESC"
			}
		} else {
			http.Error(w, "Invalid url query", 400)
			return
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
				"total_count": len(posts),
				"total_pages": int(math.Ceil(float64(len(posts)) / float64(pageLimit))),
				"page":        currentPage,
				"per_page":    pageLimit,
			},
			"results": posts,
		}

		// Encode posts into JSON
		json.NewEncoder(w).Encode(envelope)
	}

}

// Create a new post
// Take in title (string), text (string), and public (bool) from POST request
func createPost(w http.ResponseWriter, r *http.Request) {
	var newPost post

	// Decode the request's body
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newPost)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Create write & read IDs
	// First, get all the existing read and write IDs from db
	var existingIDs []int
	queryString := "SELECT read_id FROM posts UNION SELECT write_id FROM posts ORDER BY read_id;"
	err = db.Select(&existingIDs, queryString)
	if err != nil {
		http.Error(w, err.Error(), 500)
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
		fmt.Printf("Check if readId = writeID. read = %d. write = %d\n", readID, writeID)
		if readID == writeID {
			continue
		}
		// Check if readID and writeID already exists
		fmt.Println("Check if readId and writeID already exists")
		readIndex := sort.Search(n, func(i int) bool { return existingIDs[i] >= readID })
		writeIndex := sort.Search(n, func(i int) bool { return existingIDs[i] >= writeID })
		fmt.Printf("readIndex = %d. writeIndex = %d\n", readIndex, writeIndex)

		if existingIDs[readIndex] == readID || existingIDs[writeIndex] == writeID {
			continue
		}
		unique = true
	}

	// Insert new post into db
	queryString = `INSERT INTO posts (title, text, public, read_id, write_id) VALUES ($1, $2, $3, $4, $5);`
	result, err := db.Exec(queryString, newPost.Title, newPost.Text, newPost.Public, readID, writeID)

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

// Get a post by its read or write ID
// If it's a read --> Return text, title, and report link
// If it's a write --> Return text, title, public, read and write links, update and delete links
func getPostById(w http.ResponseWriter, r *http.Request) {
	// Get id from path variables
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}

	// Find id
	var result post
	queryString := `SELECT title, text, public, read_id, write_id FROM posts where read_id = $1 or write_id = $1;`
	err = db.Get(&result, queryString, id)

	if err != nil {
		http.Error(w, "Invalid post id", 400)
		return
	}

	// If id is a read
	if result.ReadID == id {
		// Return title, text, and report link
		w.Header().Set("Content-type", "application/json")
		response := map[string]interface{}{
			"post_content": map[string]interface{}{
				"title": result.Title,
				"text":  result.Text,
			},
			"read_only_options": map[string]interface{}{
				"report_link": "/report",
			},
		}
		json.NewEncoder(w).Encode(response)
	} else { // If id is a write
		// Return title, text, public, read and write links, update and delete links
		w.Header().Set("Content-type", "application/json")
		response := map[string]interface{}{
			"post_content": result,
			"admin_options": map[string]interface{}{
				"update_link": "/",
				"delete_link": "/",
			},
		}
		json.NewEncoder(w).Encode(response)
	}

}

// Report a post
// Take in a reason (string), then store it in the reported post
func reportPost(w http.ResponseWriter, r *http.Request) {
	// Get id from path variables
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}

	// Decode the request's body to get report reason
	var reason struct{ reason string }
	err = json.NewDecoder(r.Body).Decode(&reason)

	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	// Add report reason to post, and set reported to true
	queryString := `UPDATE posts SET reported = 1, report_reason = $1 WHERE read_id = $2`
	result, err := db.Exec(queryString, reason.reason, id)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%d row(s) updated.\n", rowsAffected)
}
