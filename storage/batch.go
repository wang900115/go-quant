package storage

const (
	DefaultBatchSize = 100 * 1024
)

type Batch interface {
	KVWriter
	KVRangeDeleter

	ValueSize() int
	Write() error
	Reset()
	Replay(w KVWriter) error
}

type Batcher interface {
	NewBatch() Batch

	NewBatchWithSize(size int) Batch
}

type HookBatch struct {
	Batch

	OnPut         func(key []byte, value []byte)
	OnDelete      func(key []byte) error
	OnDeleteRange func(start []byte, end []byte)
}

func (b HookBatch) Put(key []byte, value []byte) error {
	if b.OnPut != nil {
		b.OnPut(key, value)
	}
	return b.Batch.Put(key, value)
}

func (b HookBatch) Delete(key []byte) error {
	if b.OnDelete != nil {
		if err := b.OnDelete(key); err != nil {
			return err
		}
	}
	return b.Batch.Delete(key)
}

func (b HookBatch) DeleteRange(start []byte, end []byte) error {
	if b.OnDeleteRange != nil {
		b.OnDeleteRange(start, end)
	}
	return b.Batch.DeleteRange(start, end)
}
