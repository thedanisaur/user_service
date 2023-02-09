package types

type DbConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"db_name"`
	Driver   string `json:"db_driver"`
}
