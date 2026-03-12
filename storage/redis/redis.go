// Package redis implements the storage.KVStore interface backed by Redis.
//
// All binary keys are stored under a configurable key prefix using Redis
// hash-map commands (HSET / HGET / HDEL / HSCAN). This keeps all data for
// one logical "store" under a single Redis hash key, making SCAN-based
// iteration and key management simple.
package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"

	goredis "github.com/redis/go-redis/v9"
	"github.com/wang900115/quant/storage"
)

var errClosed = errors.New("redis: store closed")

// Config holds connection settings for Redis.
type Config struct {
	Addr     string // host:port, defaults to "localhost:6379"
	Password string
	DB       int
	HashKey  string // Redis hash key to store all data under, defaults to "kv_store"
}

// Database implements storage.KVStore backed by Redis using a single hash.
type Database struct {
	client  *goredis.Client
	hashKey string
	mu      sync.RWMutex
	closed  bool
}

// Open connects to Redis and returns a Database.
func Open(cfg Config) (*Database, error) {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:6379"
	}
	if cfg.HashKey == "" {
		cfg.HashKey = "kv_store"
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}

	return &Database{
		client:  client,
		hashKey: cfg.HashKey,
	}, nil
}

func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	return d.client.Close()
}

// --- KVReader ---

func (d *Database) Has(key []byte) (bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed {
		return false, errClosed
	}
	n, err := d.client.HExists(context.Background(), d.hashKey, string(key)).Result()
	return n, err
}

func (d *Database) Get(key []byte) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed {
		return nil, errClosed
	}
	val, err := d.client.HGet(context.Background(), d.hashKey, string(key)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, fmt.Errorf("redis: key not found")
	}
	return val, err
}

// --- KVWriter ---

func (d *Database) Put(key []byte, value []byte) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed {
		return errClosed
	}
	return d.client.HSet(context.Background(), d.hashKey, string(key), value).Err()
}

func (d *Database) Delete(key []byte) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed {
		return errClosed
	}
	return d.client.HDel(context.Background(), d.hashKey, string(key)).Err()
}

// --- KVRangeDeleter ---
// DeleteRange removes all fields whose key is in [startKey, endKey).
// Implemented via HSCAN + pipeline delete — O(N) where N is matching fields.
func (d *Database) DeleteRange(startKey []byte, endKey []byte) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed {
		return errClosed
	}
	ctx := context.Background()
	var cursor uint64
	var toDelete []string
	start := string(startKey)
	end := string(endKey)

	for {
		fields, next, err := d.client.HScan(ctx, d.hashKey, cursor, "*", 100).Result()
		if err != nil {
			return err
		}
		// HScan returns alternating field, value pairs
		for i := 0; i < len(fields)-1; i += 2 {
			field := fields[i]
			if field >= start && field < end {
				toDelete = append(toDelete, field)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	if len(toDelete) == 0 {
		return nil
	}
	return d.client.HDel(ctx, d.hashKey, toDelete...).Err()
}

// --- KVStater ---

func (d *Database) Stat() (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed {
		return "", errClosed
	}
	n, err := d.client.HLen(context.Background(), d.hashKey).Result()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("hash=%s fields=%d", d.hashKey, n), nil
}

// --- KVSyncer ---

// Sync flushes pending writes. Redis writes are synchronous; this is a no-op.
func (d *Database) Sync() error {
	if d.closed {
		return errClosed
	}
	return nil
}

// --- Compactor ---

// Compact is a no-op for Redis; memory management is handled by Redis itself.
func (d *Database) Compact(start []byte, limit []byte) error {
	if d.closed {
		return errClosed
	}
	return nil
}

// --- Batcher ---

func (d *Database) NewBatch() storage.Batch {
	return &redisBatch{db: d}
}

func (d *Database) NewBatchWithSize(_ int) storage.Batch {
	return &redisBatch{db: d}
}

// redisBatch buffers operations and executes them as a Redis pipeline on Write().
type redisBatch struct {
	db      *Database
	entries []batchEntry
	size    int
}

type batchEntry struct {
	op    string // "put", "delete", "deleterange"
	key   []byte
	value []byte
	end   []byte
}

func (b *redisBatch) Put(key []byte, value []byte) error {
	b.entries = append(b.entries, batchEntry{op: "put", key: key, value: value})
	b.size += len(key) + len(value)
	return nil
}

func (b *redisBatch) Delete(key []byte) error {
	b.entries = append(b.entries, batchEntry{op: "delete", key: key})
	b.size += len(key)
	return nil
}

func (b *redisBatch) DeleteRange(start []byte, end []byte) error {
	b.entries = append(b.entries, batchEntry{op: "deleterange", key: start, end: end})
	b.size += len(start) + len(end)
	return nil
}

func (b *redisBatch) ValueSize() int { return b.size }

func (b *redisBatch) Reset() {
	b.entries = b.entries[:0]
	b.size = 0
}

func (b *redisBatch) Write() error {
	if b.db.closed {
		return errClosed
	}
	ctx := context.Background()
	pipe := b.db.client.Pipeline()

	for _, e := range b.entries {
		switch e.op {
		case "put":
			pipe.HSet(ctx, b.db.hashKey, string(e.key), e.value)
		case "delete":
			pipe.HDel(ctx, b.db.hashKey, string(e.key))
		case "deleterange":
			// deleterange in pipeline: collect keys then delete — falls back to individual deletes
			// For simplicity, execute immediately (outside pipeline) for range ops
			if err := b.db.DeleteRange(e.key, e.end); err != nil {
				return err
			}
		}
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (b *redisBatch) Replay(w storage.KVWriter) error {
	for _, e := range b.entries {
		switch e.op {
		case "put":
			if err := w.Put(e.key, e.value); err != nil {
				return err
			}
		case "delete":
			if err := w.Delete(e.key); err != nil {
				return err
			}
		case "deleterange":
			if rd, ok := w.(storage.KVRangeDeleter); ok {
				if err := rd.DeleteRange(e.key, e.end); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// --- Iteratee ---

func (d *Database) NewIterator(prefix []byte, start []byte) storage.Iterator {
	return &redisIterator{
		db:     d,
		prefix: string(prefix),
		start:  string(append(prefix, start...)),
		pos:    -1,
	}
}

type redisIterator struct {
	db       *Database
	prefix   string
	start    string
	rows     []redisRow
	pos      int
	err      error
	released bool
}

type redisRow struct {
	key   []byte
	value []byte
}

func (it *redisIterator) load() {
	if it.db.closed {
		it.err = errClosed
		return
	}
	ctx := context.Background()
	var cursor uint64
	pattern := it.prefix + "*"
	if it.prefix == "" {
		pattern = "*"
	}

	var matched []redisRow
	for {
		fields, next, err := it.db.client.HScan(ctx, it.db.hashKey, cursor, pattern, 100).Result()
		if err != nil {
			it.err = err
			return
		}
		for i := 0; i < len(fields)-1; i += 2 {
			k := fields[i]
			if k < it.start {
				continue
			}
			v := []byte(fields[i+1])
			matched = append(matched, redisRow{key: []byte(k), value: v})
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}

	// Sort by key for deterministic iteration
	for i := 1; i < len(matched); i++ {
		for j := i; j > 0 && string(matched[j].key) < string(matched[j-1].key); j-- {
			matched[j], matched[j-1] = matched[j-1], matched[j]
		}
	}
	it.rows = matched
	it.pos = 0
}

func (it *redisIterator) Next() bool {
	if it.released {
		return false
	}
	if it.pos == -1 {
		it.load()
	} else {
		it.pos++
	}
	return it.pos < len(it.rows)
}

func (it *redisIterator) Error() error  { return it.err }
func (it *redisIterator) Key() []byte   { return it.rows[it.pos].key }
func (it *redisIterator) Value() []byte { return it.rows[it.pos].value }
func (it *redisIterator) Release()      { it.released = true }

// compile-time interface check
var _ storage.KVStore = (*Database)(nil)
