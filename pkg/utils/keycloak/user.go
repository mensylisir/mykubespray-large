package keycloak

type KeycloakUser struct {
	Username    string       `json:"username"`
	Email       string       `json:"email"`
	Enabled     bool         `json:"enabled"`
	FirstName   string       `json:"firstName"`
	LastName    string       `json:"lastName"`
	Credentials []Credential `json:"credentials"`
}

// Credential represents user credentials, including default password.
type Credential struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

// NewUser creates a new user with a default password.
func NewUser(username, email, firstName, lastName string) KeycloakUser {
	return KeycloakUser{
		Username:  username,
		Email:     email,
		Enabled:   true,
		FirstName: firstName,
		LastName:  lastName,
		Credentials: []Credential{
			{
				Type:      "password",
				Value:     "Def@u1tpwd",
				Temporary: false,
			},
		},
	}
}
