package dao

type UserDAO interface {
	DoesUserExist(userID string) (bool, error)
	CreateUserIfNotExists(userID, email, name string) (string, error)
	GetUserIDByEmail(email string) (string, error)
	GetUsersList(page int64, pageSize int64) ([]*User, error)

	// add account related things here
	CreateAccountIfNotExists(userID string, provider string, providerAccountID string) (string, error)
}

// add tenant dao and user <-> tenant mgmt dao

type TenantDAO interface {
	CreateTenant(name string, description string, TenantType int64, createdBy string) (string, error)
	GetTenantUsers(tenantID string, page int64, pageSize int64) ([]*TenantUser, error)
	GetTenantsList(page int64, pageSize int64) ([]*Tenant, error)
	AddUserToTenant(tenantID string, userID string, role string) (string, error)
	RemoveUserFromTenant(tenantID string, userID string) (string, error)
}

type ProviderDAO interface {
	CreateProvider(name string, enabled bool, oauthURL string, tokenURL string, redirectURL string, scope string) (string, error)
	GetProvider(providerID string) (*Provider, error)
	GetProvidersList(page int64, pageSize int64) ([]*Provider, error)
}
