package store

import (
	"reflect"
	"testing"
)

func TestMemStorage(t *testing.T) {
	store := NewMemStorage()

	key := []byte("key")
	value := []byte("value")

	err := store.Set(key, value)
	if err != nil {
		t.Errorf("failed storing value: %v\n", err)
	}

	value2, err := store.Get(key)
	if err != nil {
		t.Errorf("failed getting value: %v\n", err)
	}

	if !reflect.DeepEqual(value, value2) {
		t.Errorf("got invalid value")
	}

	_, err = store.Get([]byte("missing"))
	if err == nil {
		t.Errorf("got unexpected value")
	}
}
