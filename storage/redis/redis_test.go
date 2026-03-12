package redis

import (
	"bytes"
	"context"
	"os"
	"sort"
	"testing"
)

// --- Unit tests (no Redis required) ---

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}
	if cfg.Addr != "" {
		t.Errorf("want empty addr, got %q", cfg.Addr)
	}
	if cfg.HashKey != "" {
		t.Errorf("want empty hashkey, got %q", cfg.HashKey)
	}
}

func TestRedisBatch_PutValueSize(t *testing.T) {
	b := &redisBatch{}
	_ = b.Put([]byte("hello"), []byte("world"))
	if b.size != 10 {
		t.Errorf("want size 10, got %d", b.size)
	}
	if len(b.entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(b.entries))
	}
	if b.entries[0].op != "put" {
		t.Errorf("want op=put, got %q", b.entries[0].op)
	}
}

func TestRedisBatch_DeleteValueSize(t *testing.T) {
	b := &redisBatch{}
	_ = b.Delete([]byte("key"))
	if b.size != 3 {
		t.Errorf("want size 3, got %d", b.size)
	}
	if b.entries[0].op != "delete" {
		t.Errorf("want op=delete, got %q", b.entries[0].op)
	}
}

func TestRedisBatch_DeleteRange(t *testing.T) {
	b := &redisBatch{}
	_ = b.DeleteRange([]byte("a"), []byte("z"))
	if b.size != 2 {
		t.Errorf("want size 2, got %d", b.size)
	}
	if b.entries[0].op != "deleterange" {
		t.Errorf("want op=deleterange, got %q", b.entries[0].op)
	}
	if !bytes.Equal(b.entries[0].key, []byte("a")) {
		t.Errorf("want start=a, got %q", b.entries[0].key)
	}
	if !bytes.Equal(b.entries[0].end, []byte("z")) {
		t.Errorf("want end=z, got %q", b.entries[0].end)
	}
}

func TestRedisBatch_Reset(t *testing.T) {
	b := &redisBatch{}
	_ = b.Put([]byte("k"), []byte("v"))
	b.Reset()
	if b.size != 0 {
		t.Errorf("want size 0 after reset, got %d", b.size)
	}
	if len(b.entries) != 0 {
		t.Errorf("want 0 entries after reset, got %d", len(b.entries))
	}
}

func TestRedisBatch_ValueSize(t *testing.T) {
	b := &redisBatch{}
	_ = b.Put([]byte("abc"), []byte("de"))
	if b.ValueSize() != 5 {
		t.Errorf("want ValueSize 5, got %d", b.ValueSize())
	}
}

func TestRedisBatch_Replay(t *testing.T) {
	b := &redisBatch{}
	_ = b.Put([]byte("k1"), []byte("v1"))
	_ = b.Delete([]byte("k2"))

	w := &mockKVWriter{}
	if err := b.Replay(w); err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if len(w.puts) != 1 {
		t.Errorf("want 1 put replayed, got %d", len(w.puts))
	}
	if len(w.deletes) != 1 {
		t.Errorf("want 1 delete replayed, got %d", len(w.deletes))
	}
}

func TestRedisBatch_ReplayDeleteRange(t *testing.T) {
	b := &redisBatch{}
	_ = b.DeleteRange([]byte("a"), []byte("z"))

	w := &mockRangeWriter{}
	if err := b.Replay(w); err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if len(w.ranges) != 1 {
		t.Errorf("want 1 deleterange replayed, got %d", len(w.ranges))
	}
}

func TestRedisIterator_SortOrder(t *testing.T) {
	// Verify insertion sort used in load() produces ascending order.
	rows := []redisRow{
		{key: []byte("c"), value: []byte("3")},
		{key: []byte("a"), value: []byte("1")},
		{key: []byte("b"), value: []byte("2")},
	}
	// Apply the same sort logic from load()
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && string(rows[j].key) < string(rows[j-1].key); j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
	keys := make([]string, len(rows))
	for i, r := range rows {
		keys[i] = string(r.key)
	}
	if !sort.StringsAreSorted(keys) {
		t.Errorf("expected sorted keys, got %v", keys)
	}
}

// --- Mocks ---

type mockKVWriter struct {
	puts    [][]byte
	deletes [][]byte
}

func (m *mockKVWriter) Put(key, value []byte) error {
	m.puts = append(m.puts, key)
	return nil
}

func (m *mockKVWriter) Delete(key []byte) error {
	m.deletes = append(m.deletes, key)
	return nil
}

type mockRangeWriter struct {
	mockKVWriter
	ranges [][2][]byte
}

func (m *mockRangeWriter) DeleteRange(start, end []byte) error {
	m.ranges = append(m.ranges, [2][]byte{start, end})
	return nil
}

// --- Integration tests (require live Redis) ---

func integrationAddr(t *testing.T) string {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set; skipping integration test")
	}
	return addr
}

func openTestDB(t *testing.T) *Database {
	t.Helper()
	addr := integrationAddr(t)
	db, err := Open(Config{
		Addr:    addr,
		HashKey: "test_kv_" + t.Name(),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		// clean up test hash key
		db.client.Del(context.Background(), db.hashKey)
		db.Close()
	})
	return db
}

func TestIntegration_PutGet(t *testing.T) {
	db := openTestDB(t)
	if err := db.Put([]byte("foo"), []byte("bar")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	val, err := db.Get([]byte("foo"))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(val, []byte("bar")) {
		t.Errorf("want bar, got %s", val)
	}
}

func TestIntegration_Has(t *testing.T) {
	db := openTestDB(t)
	_ = db.Put([]byte("exists"), []byte("1"))

	ok, err := db.Has([]byte("exists"))
	if err != nil || !ok {
		t.Errorf("Has existing key: ok=%v err=%v", ok, err)
	}
	ok, err = db.Has([]byte("missing"))
	if err != nil || ok {
		t.Errorf("Has missing key: ok=%v err=%v", ok, err)
	}
}

func TestIntegration_Delete(t *testing.T) {
	db := openTestDB(t)
	_ = db.Put([]byte("del"), []byte("v"))
	if err := db.Delete([]byte("del")); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	ok, _ := db.Has([]byte("del"))
	if ok {
		t.Error("key should be deleted")
	}
}

func TestIntegration_GetNotFound(t *testing.T) {
	db := openTestDB(t)
	_, err := db.Get([]byte("nonexistent"))
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestIntegration_DeleteRange(t *testing.T) {
	db := openTestDB(t)
	for _, k := range []string{"a", "b", "c", "d"} {
		_ = db.Put([]byte(k), []byte("v"))
	}
	if err := db.DeleteRange([]byte("b"), []byte("d")); err != nil {
		t.Fatalf("DeleteRange: %v", err)
	}
	// "a" and "d" should still exist
	if ok, _ := db.Has([]byte("a")); !ok {
		t.Error("a should still exist")
	}
	if ok, _ := db.Has([]byte("d")); !ok {
		t.Error("d should still exist")
	}
	// "b" and "c" should be gone
	if ok, _ := db.Has([]byte("b")); ok {
		t.Error("b should be deleted")
	}
	if ok, _ := db.Has([]byte("c")); ok {
		t.Error("c should be deleted")
	}
}

func TestIntegration_Stat(t *testing.T) {
	db := openTestDB(t)
	_ = db.Put([]byte("x"), []byte("1"))
	stat, err := db.Stat()
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if stat == "" {
		t.Error("expected non-empty stat")
	}
}

func TestIntegration_SyncCompact(t *testing.T) {
	db := openTestDB(t)
	if err := db.Sync(); err != nil {
		t.Errorf("Sync: %v", err)
	}
	if err := db.Compact(nil, nil); err != nil {
		t.Errorf("Compact: %v", err)
	}
}

func TestIntegration_Batch(t *testing.T) {
	db := openTestDB(t)
	b := db.NewBatch()
	_ = b.Put([]byte("b1"), []byte("v1"))
	_ = b.Put([]byte("b2"), []byte("v2"))
	_ = b.Delete([]byte("b1"))
	if err := b.Write(); err != nil {
		t.Fatalf("batch Write: %v", err)
	}
	if ok, _ := db.Has([]byte("b1")); ok {
		t.Error("b1 should be deleted via batch")
	}
	if val, err := db.Get([]byte("b2")); err != nil || !bytes.Equal(val, []byte("v2")) {
		t.Errorf("b2 get: val=%q err=%v", val, err)
	}
}

func TestIntegration_Iterator(t *testing.T) {
	db := openTestDB(t)
	pairs := map[string]string{
		"pfx/a": "1",
		"pfx/b": "2",
		"pfx/c": "3",
		"other": "4",
	}
	for k, v := range pairs {
		_ = db.Put([]byte(k), []byte(v))
	}

	it := db.NewIterator([]byte("pfx/"), nil)
	defer it.Release()

	var got []string
	for it.Next() {
		got = append(got, string(it.Key()))
	}
	if err := it.Error(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("want 3 keys, got %d: %v", len(got), got)
	}
	if !sort.StringsAreSorted(got) {
		t.Errorf("expected sorted keys, got %v", got)
	}
}

func TestIntegration_ClosedDB(t *testing.T) {
	addr := integrationAddr(t)
	db, err := Open(Config{Addr: addr, HashKey: "closed_test"})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_ = db.Close()

	if _, err := db.Has([]byte("k")); err != errClosed {
		t.Errorf("Has after close: want errClosed, got %v", err)
	}
	if _, err := db.Get([]byte("k")); err != errClosed {
		t.Errorf("Get after close: want errClosed, got %v", err)
	}
	if err := db.Put([]byte("k"), []byte("v")); err != errClosed {
		t.Errorf("Put after close: want errClosed, got %v", err)
	}
	if err := db.Delete([]byte("k")); err != errClosed {
		t.Errorf("Delete after close: want errClosed, got %v", err)
	}
	if err := db.Sync(); err != errClosed {
		t.Errorf("Sync after close: want errClosed, got %v", err)
	}
	if err := db.Compact(nil, nil); err != errClosed {
		t.Errorf("Compact after close: want errClosed, got %v", err)
	}
}
