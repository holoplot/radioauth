package account

// Store is an interface to store and retrieve account details
type Store interface {
	Read(id string) (*Account, error)
	Write(a *Account) error
}
