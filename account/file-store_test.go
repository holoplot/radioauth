package account

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestFileStore_Write(t *testing.T) {
	dir, err := ioutil.TempDir("", "filestoretest")
	if err != nil {
		t.Errorf("Cannot create tempdir: %v", err)
		return
	}
	defer os.RemoveAll(dir)

	st, err := NewFileStore(dir)
	if err != nil {
		t.Error(err)
		return
	}

	a1 := NewWithRandomPassword("foo")
	err = st.Write(a1)
	if err != nil {
		t.Error(err)
		return
	}

	a2, err := st.Read("foo")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(a1, a2) {
		t.Errorf("a1 != a2! %v %v", a1, a2)
		return
	}
}
