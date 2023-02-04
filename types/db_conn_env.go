package types

type DBConnEnv struct {
	Username string `json:"MSDBUSERNAME"`
	Password string `json:"MSDBPASSWORD"`
	Name     string `json:"MSDBNAME"`
	Driver   string `json:"MSDBDRIVER"`
}
