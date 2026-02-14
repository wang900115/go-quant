package storage

type Iterator interface {
	Next() bool

	Error() error

	Key() []byte

	Value() []byte

	Release()
}

type Iteratee interface {
	NewIterator(prefix []byte, start []byte) Iterator
}
