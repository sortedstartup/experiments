package dao

type UserDAO interface {
	DoesUserExist(userID string) (bool, error)
	CreateUserIfNotExists(userID, email, name, roles, oAuthProvider, oAuthUserID string, isFederated bool) (string, error)
	GetUserIDByEmail(email string) (string, error)
	GetUsersList(page int64, pageSize int64) ([]*User, error)
}
