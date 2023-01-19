package db

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var lock = &sync.Mutex{}

var database *sql.DB

func GetInstance() *sql.DB {
	if database == nil {
		lock.Lock()
		defer lock.Unlock()
		if database == nil {
			db, err := sql.Open("mysql", "thedanisaur:toor@/movie_sunday")
			if err != nil {
				log.Println(err)
			}
			// defer db.Close()
			// See "Important settings" section.
			db.SetConnMaxLifetime(time.Minute * 3)
			db.SetMaxOpenConns(10)
			db.SetMaxIdleConns(10)

			// Open doesn't open a connection. Validate DSN data:
			err = db.Ping()
			if err != nil {
				panic(err.Error()) // proper error handling instead of panic in your app
			} else {
				// TODO write the actual db name to logs
				log.Println("Connected to: movie_sunday")
			}
			database = db
		}
	}
	return database
}
