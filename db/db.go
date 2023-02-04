package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
	"user_service/types"

	_ "github.com/go-sql-driver/mysql"
)

var lock = &sync.Mutex{}

var database *sql.DB

func GetInstance() *sql.DB {
	if database == nil {
		lock.Lock()
		defer lock.Unlock()
		if database == nil {
			env, err := readDatabaseEnv()
			if err != nil {
				log.Printf("Connection error: %s", err.Error())
			}
			conn_str := fmt.Sprintf("%s:%s@/%s", env.Username, env.Password, env.Name)
			db, err := sql.Open(env.Driver, conn_str)
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
				log.Printf("Connected to: %s", env.Name)
			}
			database = db
		}
	}
	return database
}

func readDatabaseEnv() (*types.DBConnEnv, error) {
	username, username_set := os.LookupEnv("MSDBUSERNAME")
	password, password_set := os.LookupEnv("MSDBPASSWORD")
	db_name, db_name_set := os.LookupEnv("MSDBNAME")
	db_driver, db_driver_set := os.LookupEnv("MSDBDRIVER")
	var db_conn types.DBConnEnv
	if !username_set || !password_set || !db_name_set || !db_driver_set {
		json_file, err := os.Open("./secrets/db.env")
		if err != nil {
			log.Printf("Reading env file error: %s", err.Error())
			return nil, errors.New("Could not read db env file aborting...")
		}
		defer json_file.Close()
		bytes, _ := ioutil.ReadAll(json_file)
		json.Unmarshal(bytes, &db_conn)
	} else {
		db_conn = types.DBConnEnv{
			Username: username,
			Password: password,
			Name:     db_name,
			Driver:   db_driver,
		}
	}
	return &db_conn, nil
}
