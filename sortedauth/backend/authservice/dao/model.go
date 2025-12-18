package dao

type User struct {
	ID    string `db:"user_id"`
	Email string `db:"email"`
	Name  string `db:"name"`
	Roles string `db:"roles"`
}
