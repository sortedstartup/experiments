-- SQLite complete schema migration

CREATE TABLE IF NOT EXISTS userservice_users (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_userservice_users_email ON userservice_users(email);
----
new models

1 user can have many accounts

Account
--------

id - uuid string
userid -> fk to user.id
provider - string
provider_account_id - string,

Providers
---------
id - uuid string
name - string
enabled - boolean
oauthurl - string
tokenurl - string
redirecturl - string
scope - string - space seperated - should be hardcoded for now - "openid email profile" for every new provider

// in the seed script populate the db with google, linkedin, github, facebook, apple providers

// we are not storing clientid and secret for safety reasons


		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
			TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
		},
		// provider.Endpoint(),
		Scopes: []string{oidc.ScopeOpenID, "email", "profile"},

//-----


1 user can have many tenants
1 tenant can have many users

review ad then use below - 
CREATE TABLE userservice_tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    type INTEGER NOT NULL DEFAULT 1, //  0 - system,1 - organization, 2 - personal, 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL REFERENCES userservice_users(id)
);

CREATE TABLE userservice_tenant_users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES userservice_tenants(id),
    user_id TEXT NOT NULL REFERENCES userservice_users(id),
    role TEXT NOT NULL DEFAULT 'user', -- Simple string: 'admin'
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, user_id)
);

prefix all tables with sortedauth_