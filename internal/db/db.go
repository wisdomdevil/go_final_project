package db

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func checkFileExists(dbFile string) bool {
	log.Printf("Check file existance %s", dbFile)

	_, err := os.Stat(dbFile)

	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("DB file %s doesn't exist.", dbFile)
			return false
		}
		log.Fatal(err)
		return false
	}
	log.Printf("DB file %s exists.", dbFile)
	return true
}

func dbCreate(dbFilePath string) {
	// формируем строку для дальнейшего создания таблицы task (в тестах scheduler)
	taskTableCreateQuery := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date    VARCHAR(8) NOT NULL,
		title   VARCHAR(128) NOT NULL,
		comment VARCHAR(250),
		repeat  VARCHAR(128)
	);
	CREATE INDEX IF NOT EXISTS scheduler_date ON scheduler(date);
	`

	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// в методе Exec базе данных отправляется строка запроса taskTableCreateQuery на выполнение
	_, err = db.Exec(taskTableCreateQuery)
	if err != nil {
		log.Fatal(err)
	}
}

// dbConnection checks DB existance and creates if it doesn't exist.
func CreateDatabase(dbPath string) {
	if !checkFileExists(dbPath) {
		dbCreate(dbPath)
	}
}
