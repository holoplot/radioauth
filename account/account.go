package account

import (
	"github.com/dchest/uniuri"
	"time"
)

// Account holds a user account
type Account struct {
	Username     string    `json:"username,omitempty"`
	Password     string    `json:"password,omitempty"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenExpiry  time.Time `json:"token_expiry,omitempty"`
}

// NewWithRandomPassword creates a new Account with a random password
func NewWithRandomPassword(username string) *Account {
	a := new(Account)
	a.Username = username
	a.Password = uniuri.NewLen(32)

	return a
}
