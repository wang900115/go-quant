package pebble

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cockroachdb/pebble"
	"github.com/wang900115/quant/storage"
)

func NewDatabase(path string) (*Database, error) {
	db, err := pebble.Open(path, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &Database{
		db:           db,
		writeOptions: pebble.Sync,
	}, nil
}

type Database struct {
	db *pebble.DB

	quitLock sync.RWMutex
	closed   bool

	writeOptions *pebble.WriteOptions
}

func (d *Database) Close() error {
	d.quitLock.Lock()
	defer d.quitLock.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	return d.db.Close()
}

func (d *Database) Has(key []byte) (bool, error) {
	d.quitLock.RLock()
	defer d.quitLock.RUnlock()
	if d.closed {
		return false, pebble.ErrClosed
	}
	_, closer, err := d.db.Get(key)
	if err == pebble.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if err := closer.Close(); err != nil {
		return false, err
	}
	return true, nil
}

func (d *Database) Get(key []byte) ([]byte, error) {
	d.quitLock.RLock()
	defer d.quitLock.RUnlock()
	if d.closed {
		return nil, pebble.ErrClosed
	}
	data, closer, err := d.db.Get(key)
	if err != nil {
		return nil, err
	}
	ret := make([]byte, len(data))
	copy(ret, data)
	if err := closer.Close(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *Database) Put(key []byte, value []byte) error {
	d.quitLock.RLock()
	defer d.quitLock.RUnlock()
	if d.closed {
		return pebble.ErrClosed
	}
	return d.db.Set(key, value, d.writeOptions)
}

func (d *Database) Delete(key []byte) error {
	d.quitLock.RLock()
	defer d.quitLock.RUnlock()
	if d.closed {
		return pebble.ErrClosed
	}
	return d.db.Delete(key, d.writeOptions)
}

func (d *Database) DeleteRange(start []byte, end []byte) error {
	d.quitLock.RLock()
	defer d.quitLock.RUnlock()
	if d.closed {
		return pebble.ErrClosed
	}
	return d.db.DeleteRange(start, end, d.writeOptions)
}

func (d *Database) Stat() (string, error) {
	d.quitLock.RLock()
	defer d.quitLock.RUnlock()
	if d.closed {
		return "", pebble.ErrClosed
	}
	return fmt.Sprintf("%+v", d.db.Metrics()), nil
}

func (d *Database) Sync() error {
	d.quitLock.RLock()
	defer d.quitLock.RUnlock()
	if d.closed {
		return pebble.ErrClosed
	}
	return d.db.Flush()
}

func (d *Database) Compact(start []byte, limit []byte) error {
	d.quitLock.RLock()
	defer d.quitLock.RUnlock()
	if d.closed {
		return pebble.ErrClosed
	}
	return d.db.Compact(start, limit, true)
}

func (d *Database) NewBatch() storage.Batch {
	return &batch{b: d.db.NewBatch(), db: d}
}

func (d *Database) NewBatchWithSize(size int) storage.Batch {
	return &batch{b: d.db.NewBatchWithSize(size), db: d, size: size}
}

// iteratorUpperBound returns the smallest key that is strictly greater than
// all keys with the given prefix, so pebble stops scanning at the right point.
func iteratorUpperBound(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	return nil // all-0xFF prefix: no upper bound
}

func (d *Database) NewIterator(prefix []byte, start []byte) storage.Iterator {
	iter, _ := d.db.NewIter(&pebble.IterOptions{
		LowerBound: append(prefix, start...),
		UpperBound: iteratorUpperBound(prefix),
	})
	iter.First()
	return &iterator{
		iter:     iter,
		moved:    true,
		released: false,
	}
}

type batch struct {
	b    *pebble.Batch
	db   *Database
	size int
}

func (b *batch) Put(key []byte, value []byte) error {
	if err := b.b.Set(key, value, nil); err != nil {
		return err
	}
	b.size += len(key) + len(value)
	return nil
}

func (b *batch) Delete(key []byte) error {
	if err := b.b.Delete(key, nil); err != nil {
		return err
	}
	b.size += len(key)
	return nil
}

func (b *batch) DeleteRange(start []byte, end []byte) error {
	if err := b.b.DeleteRange(start, end, nil); err != nil {
		return err
	}
	b.size += len(start) + len(end)
	return nil
}

func (b *batch) ValueSize() int {
	return b.size
}

func (b *batch) Write() error {
	return b.b.Commit(b.db.writeOptions)
}

func (b *batch) Reset() {
	b.b.Reset()
	b.size = 0
}

func (b *batch) Replay(w storage.KVWriter) error {
	reader := b.b.Reader()
	for {
		kind, k, v, ok, err := reader.Next()
		if !ok || err != nil {
			return err
		}
		switch kind {
		case pebble.InternalKeyKindSet:
			if err := w.Put(k, v); err != nil {
				return err
			}
		case pebble.InternalKeyKindDelete:
			if err := w.Delete(k); err != nil {
				return err
			}
		case pebble.InternalKeyKindRangeDelete:
			if rangeDeleter, ok := w.(storage.KVRangeDeleter); ok {
				if err := rangeDeleter.DeleteRange(k, v); err != nil {
					return err
				}
			} else {
				return errors.New("writer does not support range delete")
			}
		default:
			return pebble.ErrInvalidBatch
		}
	}
}

type iterator struct {
	iter     *pebble.Iterator
	moved    bool
	released bool
}

func (it *iterator) Next() bool {
	if it.moved {
		it.moved = false
		return it.iter.Valid()
	}
	return it.iter.Next()
}

func (it *iterator) Error() error {
	return it.iter.Error()
}

func (it *iterator) Key() []byte {
	return it.iter.Key()
}

func (it *iterator) Value() []byte {
	return it.iter.Value()
}

func (it *iterator) Release() {
	if !it.released {
		it.iter.Close()
		it.released = true
	}
}

// compile-time interface check
var _ storage.KVStore = (*Database)(nil)
