package store

type Storage interface {
	Set(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
}
