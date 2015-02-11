//api.bookshelf.com
//Powers the bookshelf website & apps, uses PostgresSQL as datastore.
package main

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/unrolled/render"
)

//Constants for Condition field in book details.
const (
	Readable = 0
	Good     = 1
	New      = 2
)

//Book object to store the details of a book this includes the title,
//author, (link to the) image, condition as a byte it can be one of
//the values from the const declaration.
type Book struct {
	Title     string
	Author    string
	Image     string
	Condition byte
	Price     uint16
}

//Render object to render data in JSON format.
var (
	ren = render.New()
)

//Connect to Postgres instance
func SetupDB() *sql.DB {
	db, err := sql.Open("postgres", "dbname=bookshop sslmode=disable")
	Panicif(err)
	return db
}

//Panic if err is not nil.
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

//API homepage, shows available paths & methods.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]string)

	response["message"] = "Bookshelf API"
	response["/"] = "[GET] This message."
	response["/books"] = "[GET] List of books available."
	response["/new"] = "[POST] Add new book details."
	response["/book/:id"] = "[GET] Book info for corresponding id."
	response["/buy/:id"] = "[POST] Buy book if allowed"

	ren.JSON(w, http.StatusOK, response)
}

//Add new book to database.
func newHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	PanicIf(err)

	book := new(Book)
	decoder := schema.NewDecoder()
	err = decoder.Decode(book, r.PostForm)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		saveToDb(book, w)
	}
}

//Save validated book object to database
func saveToDb(book *Book, w http.ResponseWriter) {
	rows, err := db.Query(`insert into
							books (title, author, image, condition, price)
							values ($1, $2, $3, $4, $5)`,
		book.Title, book.Author, book.Image, book.Condition, book.Price)
	Panicif(err)
	defer rows.Close()
	fmt.Fprintf("Saved to database: %v", book)
}

//Match routes to their handlers and return a mux.Router object.
func getRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/new", newHandler).Methods("POST")
	return router
}

func main() {
	n := negroni.Classic()
	n.UseHandler(getRoutes())

	n.Run(":3000")
}
