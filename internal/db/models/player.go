package db

type Player struct {
	Id          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	FirstName   string `json:"firstName" db:"first_name"`
	LastName    string `json:"lastName" db:"last_name"`
	Age         int    `json:"age" db:"age"`
	Nationality string `json:"nationality" db:"nationality"`
	Height      string `json:"height" db:"height"`
	Weight      string `json:"weight" db:"weight"`
	Photo       string `json:"photo" db:"photo"`
	CreatedAt   string `json:"createdAt" db:"created_at"`
}
