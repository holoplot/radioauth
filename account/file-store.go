package account

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// FileStore implements the Store interface and uses plain files
type FileStore struct {
	path string
}

// NewFileStore returns a new FileStore.
// It returns an error if the given path does not exist or is not a directory.
func NewFileStore(path string) (*FileStore, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("Path %s does not exist", path)
	}

	err = nil
	if !info.IsDir() {
		return nil, fmt.Errorf("Path %s is not a directiry", path)
	}

	st := new(FileStore)
	st.path = path

	return st, nil
}

func (st *FileStore) Read(id string) (*Account, error) {
	path := fmt.Sprintf("%s/%s.json", st.path, id)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	a := Account{}
	err = json.Unmarshal(b, &a)
	if err != nil {
		return nil, err
	}

	if id != a.Username {
		return nil, fmt.Errorf("Username of account %s does not match %s", a.Username, id)
	}

	return &a, nil
}

func (st *FileStore) Write(a *Account) error {
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s.json", st.path, a.Username)
	return ioutil.WriteFile(path, b, 0600)
}
