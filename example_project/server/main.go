package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed templates/view.tmpl
	viewTmpl string
	//go:embed templates/edit.tmpl
	editTmpl string
	//go:embed sql/schema.sql
	schema string
	//go:embed sql/person/find.sql
	findPersonStmt string
	//go:embed sql/person/create.sql
	createPersonStmt string
	//go:embed sql/person/update.sql
	updatePersonStmt string
	db               *sqlx.DB
)

type Person struct {
	ID        int          `db:"id"`
	FirstName string       `db:"first_name"`
	LastName  string       `db:"last_name"`
	Email     string       `db:"email"`
	CreatedAt sql.NullTime `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	var err error
	db, err = sqlx.Open("mysql", "tcp(127.0.0.1:3306)/test?parseTime=true")
	checkErr(err)
	defer db.Close()

	_, err = db.Exec(schema)
	checkErr(err)

	_, err = db.Exec(createPersonStmt)
	checkErr(err)

	http.Handle("/", cors(index))
	http.Handle("/edit", cors(edit))

	fmt.Println("listening on port", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		person := loadUser()
		person.FirstName = r.FormValue("first_name")
		person.LastName = r.FormValue("last_name")
		person.Email = r.FormValue("email")
		if _, err := db.NamedExec(updatePersonStmt, person); err != nil {
			log.Println(err)
		}
	}
	tmpl := template.Must(template.New("view").Parse(viewTmpl))
	tmpl.Execute(w, loadUser())
}

func edit(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("edit").Parse(editTmpl))
	tmpl.Execute(w, loadUser())
}

func loadUser() Person {
	person := Person{}
	if err := db.Get(&person, findPersonStmt); err != nil {
		log.Println(err)
	}
	return person
}

func cors(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
