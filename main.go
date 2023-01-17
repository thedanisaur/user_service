package main

import (
	"net/http"
	"log"
	"user_auth/handlers"
	"user_auth/db"
)

func main() {
	defer db.GetInstance().Close()
	http.HandleFunc(handlers.UserRoute(), handlers.UserHandler)
	err := http.ListenAndServe(":4321", nil)
	if err != nil {
		log.Fatal("Error starting http server:", err)
		return
	}
}