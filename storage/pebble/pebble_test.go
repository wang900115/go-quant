package pebble

import (
	"bytes"
	"sync"
	"testing"

	"github.com/wang900115/quant/storage"
)

func newTestDB(t *testing.T) *Database {
	t.Helper()
	db, err := NewDatabase(t.TempDir())
	if err != nil {
		t.Fatalf("NewDatabase: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// --- Constructor & Close ---

func TestNewDatabase(t *testing.T) {
	db, err := NewDatabase(t.TempDir())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestClose_DoesNotPanic(t *testing.T) {
	db := newTestDB(t)
	db.Close()
	// second Close should not panic (may return error – that's fine)
	_ = db.Close()
}

// --- KVReader ---

func TestHas_Missing(t *testing.T) {
	db := newTestDB(t)
	ok, err := db.Has([]byte("missing"))
	if err != nil {
		t.Fatalf("Has: %v", err)
	}
	if ok {
		t.Fatal("expected false for missing key")
	}
}

func TestHas_Present(t *testing.T) {
	db := newTestDB(t)
	if err := db.Put([]byte("k"), []byte("v")); err != nil {
		t.Fatal(err)
	}
	ok, err := db.Has([]byte("k"))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected true after Put")
	}
}

func TestGet_Missing(t *testing.T) {
	db := newTestDB(t)
	val, err := db.Get([]byte("missing"))
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if val != nil {
		t.Fatal("expected nil value for missing key")
	}
}

func TestGet_Present(t *testing.T) {
	db := newTestDB(t)
	if err := db.Put([]byte("key"), []byte("value")); err != nil {
		t.Fatal(err)
	}
	val, err := db.Get([]byte("key"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(val, []byte("value")) {
		t.Fatalf("expected 'value', got %q", val)
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	db := newTestDB(t)
	if err := db.Put([]byte("k"), []byte("original")); err != nil {
		t.Fatal(err)
	}
	val, _ := db.Get([]byte("k"))
	val[0] = 'X'
	val2, _ := db.Get([]byte("k"))
	if !bytes.Equal(val2, []byte("original")) {
		t.Fatal("Get should return a copy; modifying it affected stored value")
	}
}

// --- KVWriter ---

func TestPut(t *testing.T) {
	db := newTestDB(t)
	if err := db.Put([]byte("a"), []byte("1")); err != nil {
		t.Fatal(err)
	}
	val, err := db.Get([]byte("a"))
	if err != nil || !bytes.Equal(val, []byte("1")) {
		t.Fatalf("Put/Get round-trip failed: val=%q err=%v", val, err)
	}
}

func TestPut_OverwriteKey(t *testing.T) {
	db := newTestDB(t)
	db.Put([]byte("k"), []byte("first"))
	db.Put([]byte("k"), []byte("second"))
	val, _ := db.Get([]byte("k"))
	if !bytes.Equal(val, []byte("second")) {
		t.Fatalf("expected 'second', got %q", val)
	}
}

func TestDelete_Existing(t *testing.T) {
	db := newTestDB(t)
	db.Put([]byte("k"), []byte("v"))
	if err := db.Delete([]byte("k")); err != nil {
		t.Fatal(err)
	}
	ok, _ := db.Has([]byte("k"))
	if ok {
		t.Fatal("key should be gone after Delete")
	}
}

func TestDelete_Missing(t *testing.T) {
	db := newTestDB(t)
	if err := db.Delete([]byte("nonexistent")); err != nil {
		t.Fatalf("Delete of missing key should not error: %v", err)
	}
}

// --- KVRangeDeleter ---

func TestDeleteRange(t *testing.T) {
	db := newTestDB(t)
	for _, k := range []string{"a", "b", "c", "d", "e"} {
		db.Put([]byte(k), []byte("v"))
	}
	// DeleteRange removes [b, e) — so b, c, d are removed; a and e remain
	if err := db.DeleteRange([]byte("b"), []byte("e")); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"b", "c", "d"} {
		ok, _ := db.Has([]byte(k))
		if ok {
			t.Errorf("key %q should have been deleted", k)
		}
	}
	for _, k := range []string{"a", "e"} {
		ok, _ := db.Has([]byte(k))
		if !ok {
			t.Errorf("key %q should still exist", k)
		}
	}
}

// --- Batch ---

func TestBatch_PutAndCommit(t *testing.T) {
	db := newTestDB(t)
	b := db.NewBatch()
	b.Put([]byte("bk"), []byte("bv"))
	if err := b.Write(); err != nil {
		t.Fatal(err)
	}
	val, err := db.Get([]byte("bk"))
	if err != nil || !bytes.Equal(val, []byte("bv")) {
		t.Fatalf("batch Put+Write: val=%q err=%v", val, err)
	}
}

func TestBatch_DeleteAndCommit(t *testing.T) {
	db := newTestDB(t)
	db.Put([]byte("dk"), []byte("dv"))
	b := db.NewBatch()
	b.Delete([]byte("dk"))
	b.Write()
	ok, _ := db.Has([]byte("dk"))
	if ok {
		t.Fatal("key should be gone after batch Delete+Write")
	}
}

func TestBatch_Reset(t *testing.T) {
	db := newTestDB(t)
	b := db.NewBatch()
	b.Put([]byte("rk"), []byte("rv"))
	b.Reset()
	b.Write()
	ok, _ := db.Has([]byte("rk"))
	if ok {
		t.Fatal("key should not exist after batch Reset before Write")
	}
}

func TestBatch_ValueSize(t *testing.T) {
	db := newTestDB(t)
	b := db.NewBatch()
	if b.ValueSize() != 0 {
		t.Fatal("ValueSize should be 0 for new batch")
	}
	b.Put([]byte("key"), []byte("value"))
	if b.ValueSize() == 0 {
		t.Fatal("ValueSize should be > 0 after Put")
	}
	b.Reset()
	if b.ValueSize() != 0 {
		t.Fatal("ValueSize should be 0 after Reset")
	}
}

func TestBatch_Replay(t *testing.T) {
	db := newTestDB(t)
	b := db.NewBatch()
	b.Put([]byte("r1"), []byte("v1"))
	b.Put([]byte("r2"), []byte("v2"))

	target := db.NewBatch()
	if err := b.Replay(target); err != nil {
		t.Fatal(err)
	}
	target.Write()

	for k, expected := range map[string]string{"r1": "v1", "r2": "v2"} {
		val, err := db.Get([]byte(k))
		if err != nil || !bytes.Equal(val, []byte(expected)) {
			t.Errorf("Replay: key=%q val=%q err=%v", k, val, err)
		}
	}
}

// --- Iterator ---

func TestIterator_ScanAll(t *testing.T) {
	db := newTestDB(t)
	keys := []string{"a", "b", "c", "d", "e"}
	for _, k := range keys {
		db.Put([]byte(k), []byte(k+"v"))
	}
	it := db.NewIterator(nil, nil)
	defer it.Release()

	var got []string
	for it.Next() {
		got = append(got, string(it.Key()))
	}
	if err := it.Error(); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(keys) {
		t.Fatalf("expected %d keys, got %d: %v", len(keys), len(got), got)
	}
}

func TestIterator_PrefixScan(t *testing.T) {
	db := newTestDB(t)
	db.Put([]byte("foo:1"), []byte("a"))
	db.Put([]byte("foo:2"), []byte("b"))
	db.Put([]byte("bar:1"), []byte("c"))

	it := db.NewIterator([]byte("foo:"), nil)
	defer it.Release()

	var got []string
	for it.Next() {
		got = append(got, string(it.Key()))
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 keys with prefix 'foo:', got %d: %v", len(got), got)
	}
}

func TestIterator_Empty(t *testing.T) {
	db := newTestDB(t)
	it := db.NewIterator(nil, nil)
	defer it.Release()
	if it.Next() {
		t.Fatal("expected Next()=false on empty db")
	}
}

func TestIterator_Release(t *testing.T) {
	db := newTestDB(t)
	it := db.NewIterator(nil, nil)
	it.Release()
	it.Release() // should not panic
}

// --- Maintenance ---

func TestStat(t *testing.T) {
	db := newTestDB(t)
	s, err := db.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if s == "" {
		t.Fatal("Stat returned empty string")
	}
}

func TestSync(t *testing.T) {
	db := newTestDB(t)
	if err := db.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestCompact(t *testing.T) {
	db := newTestDB(t)
	db.Put([]byte("a"), []byte("v"))
	db.Put([]byte("z"), []byte("v"))
	if err := db.Compact([]byte("a"), []byte("z")); err != nil {
		t.Fatal(err)
	}
}

// --- Concurrency ---

func TestConcurrentReadWrite(t *testing.T) {
	db := newTestDB(t)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := []byte{byte(n)}
			db.Put(key, key)
			db.Get(key)
		}(i)
	}
	wg.Wait()
}

// --- Closed DB ---

func TestClosed_Put(t *testing.T) {
	db := newTestDB(t)
	db.Close()
	if err := db.Put([]byte("k"), []byte("v")); err == nil {
		t.Fatal("expected error on Put after Close")
	}
}

func TestClosed_Get(t *testing.T) {
	db := newTestDB(t)
	db.Close()
	if _, err := db.Get([]byte("k")); err == nil {
		t.Fatal("expected error on Get after Close")
	}
}

// compile-time check that *Database satisfies storage.KVStore
var _ storage.KVStore = (*Database)(nil)
