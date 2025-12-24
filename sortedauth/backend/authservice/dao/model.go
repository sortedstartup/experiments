package dao

type User struct {
	ID    string `db:"user_id"`
	Email string `db:"email"`
	Name  string `db:"name"`
	Roles string `db:"roles"`
}

type Tenant struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	TenantType  int64  `db:"type"`
	CreatedBy   string `db:"created_by"`
}

type TenantUser struct {
	UserId   string `db:"user_id"`
	Name     string `db:"name"`
	Email    string `db:"email"`
	TenantId string `db:"tenant_id"`
	Role     string `db:"role"`
}

type Provider struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Enabled     bool   `db:"enabled"`
	OAuthURL    string `db:"oauth_url"`
	TokenURL    string `db:"token_url"`
	RedirectURL string `db:"redirect_url"`
	Scope       string `db:"scope"`
}
