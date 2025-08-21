package domain

type User struct {
	ID       string
	Email    string
	Password string
}

func (u *User) SetID(id string) {
	u.ID = id
}

func (u *User) SetPass(pass string) {
	u.Password = pass
}
