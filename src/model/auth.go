package model

type Auth struct {
	ID       int    `gorm:"primary_key" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func CheckAuth(username, password string) bool {
	var auth Auth
	Db.Select("id").Where(Auth{Username: username, Password: password}).First(&auth)
	return auth.ID > 0
}
