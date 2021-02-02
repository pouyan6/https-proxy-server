package mydb

import (
	"database/sql"
	"log"
	"os"

	// sqlite3 package
	_ "github.com/mattn/go-sqlite3"
)

// Record is...
type Record struct {
	ReqURL string
}

var dbConn *sql.DB

// Init is setup db
func Init() {
	file, err := os.Open("traffic-info.db")
	if err != nil {
		_, err := os.Create("traffic-info.db")
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	log.Println("Creating db...")
	file.Close()
	log.Println("db created")
	dbConn, err = sql.Open("sqlite3", "traffic-info.db")
	if err != nil {
		log.Fatal("error")
	}
	_, ok := dbConn.Query("select * from traffic")
	if ok != nil {
		createTable()
	}

}

func createTable() {
	createStudentTableSQL := `CREATE TABLE traffic (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"url" TEXT
		);`
	statement, err := dbConn.Prepare(createStudentTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
	log.Println("Table created")
}

// CloseDatabase database
func CloseDatabase() {
	dbConn.Close()
}

// InsertRecord add a new record
func InsertRecord(data Record) {
	// log.Println("Inserting record ...")
	insertDataSQL := `INSERT INTO traffic (url) VALUES (?)`
	statement, err := dbConn.Prepare(insertDataSQL) // Prepare statement.

	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(data.ReqURL)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// DisplayRecords show records
func DisplayRecords(db *sql.DB) {
	row, err := db.Query("SELECT * FROM traffic")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var id int
		var data string
		row.Scan(&id, &data)
		log.Println("Record: ", data)
	}
}
