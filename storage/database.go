package storage

import "io"

// HOT DATA

type KVReader interface {
	Has(key []byte) (bool, error)
	Get(Key []byte) ([]byte, error)
}

type KVWriter interface {
	Put(key []byte, value []byte) error
	Delete(key []byte) error
}

type KVRangeDeleter interface {
	DeleteRange(startKey []byte, endKey []byte) error
}

type KVStater interface {
	Stat() (string, error)
}

type KVSyncer interface {
	Sync() error
}

type Compactor interface {
	Compact(start []byte, limit []byte) error
}

type KVStore interface {
	KVReader
	KVWriter
	KVRangeDeleter
	KVStater
	KVSyncer
	Batcher
	Iteratee
	Compactor
	io.Closer
}

// COLD DATA
type AncientReaderOp interface {
	Ancient(kind string, index uint64) ([]byte, error)

	AncientRange(kind string, start, count, maxBytes uint64) ([][]byte, error)

	Ancients() (uint64, error)

	Tail() (uint64, error)

	AncientSize(kind string) (uint64, error)
}

type AncientReader interface {
	AncientReaderOp

	ReadAncients(fn func(AncientReaderOp) error) error
}

type AncientWriterOp interface {
	Append(kind string, index uint64, data any) error

	AppendRaw(kind string, index uint64, data []byte) error
}

type AncientWriter interface {
	ModifyAncients(fn func(AncientWriterOp) error) error

	SyncAncient() error

	TruncateHead(n uint64) (uint64, error)

	TruncateTail(n uint64) (uint64, error)
}

type AncientStater interface {
	AncientStat() (string, error)
}

type Reader interface {
	KVReader
	AncientReader
}

type AncientStore interface {
	AncientReader
	AncientWriter
	AncientStater
	io.Closer
}

type RollbackAncientStore interface {
	AncientStore

	Rollback() error
}

// DATABASE: includes both hot and cold data
type Database interface {
	KVStore
	AncientStore
}
