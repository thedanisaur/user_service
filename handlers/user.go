package handlers

import (
    _ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"net/http"
	"fmt"
	"user_auth/types"
	"user_auth/db"
)

func UserRoute() string {
	return "/users"
}

func UserHandler(writer http.ResponseWriter, request *http.Request) {
	// txid := uuid.New()
	txid := types.UUID{ID: "1234567"}
	fmt.Printf("UserHandler | %s\n", txid.String())
	switch request.Method {
	case "GET":
		result := userGet()
		if result == nil {
			msg := fmt.Sprintf("%s %s failed: %s", request.Method, UserRoute(), txid.String())
			err := types.Error{Msg: msg}
			json.NewEncoder(writer).Encode(err)
		} else {
			json.NewEncoder(writer).Encode(result)
		}
	case "POST":
		result := userPost()
		// if result == nil {
		// 	msg := fmt.Sprintf("%s %s failed: %s", request.Method, UserRoute(), txid.String())
		// 	err := types.Error{Msg: msg}
		// 	json.NewEncoder(writer).Encode(err)
		// } else {
		json.NewEncoder(writer).Encode(result)
		// }
	default:
		msg := fmt.Sprintf("%s %s unavailable: %s", request.Method, UserRoute(), txid.String())
		result := types.Error{Msg: msg}
		json.NewEncoder(writer).Encode(result)
	}
}

func userGet() []types.User {
	fmt.Println("userGet")
	database := db.GetInstance()
	// Execute the query
	rows, err := database.Query("SELECT BIN_TO_UUID(person_id) person_id, person_username, person_password, person_email FROM people")
	if err != nil {
		fmt.Printf("Failed to query databse\n%s\n", err.Error())
		return nil
	}

	var users []types.User
	for rows.Next() {
		var user types.User
		err = rows.Scan(&user.ID, &user.Username, &user.Password, &user.Email)
		if err != nil {
			fmt.Printf("Failed to scan row\n%s\n", err.Error())
			return nil
		}
		users = append(users, user)
	}

	err = rows.Err()
	if err != nil {
		fmt.Printf("Failed after row scan\n%s\n", err.Error())
		return nil
	}

	return users
}

func userPost() types.Error {
	return types.Error{Msg: "POST"}
}