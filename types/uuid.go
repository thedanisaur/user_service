package types

type UUID struct {
	ID string `json:"string"`
}

func (id UUID) String() string {
	return "1234567"
}