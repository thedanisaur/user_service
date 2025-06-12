package types

import "github.com/google/uuid"

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	CreatedOn string    `json:"created_on"`
}

type UserUpdatePassword struct {
	Current  string    `json:"current"`
	Updated  string    `json:"updated"`
}
