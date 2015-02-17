//api.bookshelf.com
//Powers the bookshelf website & apps, uses PostgresSQL as datastore.
package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/unrolled/render"

	_ "github.com/lib/pq"
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

var (
	//Render object to render data in JSON format.
	ren = render.New()
	//Databse instance
	db = SetupDB()
)

//SetupDB Connect to Postgres instance
func SetupDB() *sql.DB {
	db, err := sql.Open("postgres", "dbname=bookshelf sslmode=disable")
	PanicIf(err)
	return db
}

//PanicIf Panic if err is not nil.
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
	response["/book"] = "[POST] Add new book details."
	response["/book/:id"] = "[GET] Book info for corresponding id."
	response["/buy/:id"] = "[POST] Buy book if allowed"

	ren.JSON(w, http.StatusOK, response)
}

//Add new book to database.
func newBookHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	PanicIf(err)

	//Extract book object from POST request.
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
	PanicIf(err)
	defer rows.Close()
	fmt.Fprintf(w, "Saved to database: %v", book)
}

//Get book corresponding to a unique id.
func getBookByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	rows, err := db.Query(`select title, author, image, condition, price
							from books where id = $1`, id)

	PanicIf(err)
	defer rows.Close()

	book := Book{}
	if rows.Next() {
		PanicIf(rows.Err())
		err := rows.Scan(&book.Title, &book.Author, &book.Image, &book.Condition, &book.Price)
		PanicIf(err)
		ren.JSON(w, http.StatusOK, book)
	} else {
		ren.JSON(w, http.StatusBadRequest, map[string]string{"Book": "None"})
	}
}

func getAllBooksHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`select title, author, image, condition, price
							from books`)

	PanicIf(err)
	defer rows.Close()

	books := []Book{}
	for rows.Next() {
		PanicIf(rows.Err())
		book := Book{}
		err := rows.Scan(&book.Title, &book.Author, &book.Image, &book.Condition, &book.Price)
		PanicIf(err)
		books = append(books, book)
	}

	ren.JSON(w, http.StatusOK, books)
}

//Match routes to their handlers and return a mux.Router object.
func getRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/books", getAllBooksHandler)
	router.HandleFunc("/book", newBookHandler).Methods("POST")
	router.HandleFunc("/book/{id:[0-9]+}", getBookByIDHandler)
	return router
}

func main() {
	n := negroni.Classic()
	defer db.Close()
	n.UseHandler(getRoutes())
	n.Run(":8000")
}
