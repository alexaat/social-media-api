package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/golang-migrate/migrate/source/file"
)

var db *sql.DB

func CreateDatabase() (*sql.DB, error) {

	dbLocal, err := sql.Open("sqlite3", "./db/social.db")

	db = dbLocal

	if err != nil {
		return nil, err
	}

	//FixMigraton()

	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, err
	}

	fSrc, err := (&file.File{}).Open("./db/migrations")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("file", fSrc, "sqlite3", instance)
	if err != nil {
		return nil, err
	}

	// modify for Down
	if err := m.Up(); err != nil {
		return nil, err
	}

	return db, nil
}

func FixMigraton() {
	query := `
	UPDATE
	schema_migrations
	SET
	version = 7,
	dirty = false
	WHERE 
	version = 8`

	statement, err := db.Prepare(query)
	if err != nil {
		fmt.Println(err)
	}

	defer statement.Close()

	_, err = statement.Exec()
	if err != nil {
		fmt.Println(err)
	}
}
