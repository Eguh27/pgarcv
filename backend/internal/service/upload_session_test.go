package service

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// runtimeGoroutineCount mengembalikan jumlah goroutine yang sedang berjalan.
func runtimeGoroutineCount() int { return runtime.NumGoroutine() }

// TestSessionStorePutAndGet memverifikasi sesi tersimpan dan bisa diambil.
func TestSessionStorePutAndGet(t *testing.T) {
	store := NewSessionStoreWithTTL(time.Hour, time.Hour)
	defer store.Close()

	session, err := store.Put("upload-1", "video.mp4", 3, 1024, t.TempDir())
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if _, err := os.Stat(session.ChunksDir); err != nil {
		t.Fatalf("direktori chunk tidak dibuat: %v", err)
	}

	got, ok := store.Get("upload-1")
	if !ok {
		t.Fatal("sesi tidak ditemukan")
	}
	if got.TotalChunks != 3 || got.TotalSize != 1024 || got.Filename != "video.mp4" {
		t.Errorf("sesi tidak sesuai: %+v", got)
	}

	if _, ok := store.Get("tidak-ada"); ok {
		t.Error("sesi yang tidak ada seharusnya tidak ditemukan")
	}
}

// TestSessionMarkReceivedAndComplete memverifikasi pelacakan progres chunk.
func TestSessionMarkReceivedAndComplete(t *testing.T) {
	store := NewSessionStoreWithTTL(time.Hour, time.Hour)
	defer store.Close()

	session, err := store.Put("upload-2", "video.mp4", 3, 1024, t.TempDir())
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	if session.IsComplete() {
		t.Error("sesi baru seharusnya belum lengkap")
	}

	if got := session.MarkReceived(0); got != 1 {
		t.Errorf("received = %d want 1", got)
	}
	// Chunk duplikat tidak boleh dihitung dua kali.
	if got := session.MarkReceived(0); got != 1 {
		t.Errorf("received setelah chunk duplikat = %d want 1", got)
	}

	session.MarkReceived(1)
	session.MarkReceived(2)
	if !session.IsComplete() {
		t.Error("sesi seharusnya lengkap setelah semua chunk masuk")
	}
}

// TestSessionStoreRemoveCleansChunks memastikan Remove menghapus direktori chunk.
func TestSessionStoreRemoveCleansChunks(t *testing.T) {
	store := NewSessionStoreWithTTL(time.Hour, time.Hour)
	defer store.Close()

	session, err := store.Put("upload-3", "video.mp4", 1, 128, t.TempDir())
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	chunkPath := filepath.Join(session.ChunksDir, "chunk_0")
	if err := os.WriteFile(chunkPath, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	store.Remove("upload-3")

	if _, ok := store.Get("upload-3"); ok {
		t.Error("sesi masih ada setelah Remove")
	}
	if _, err := os.Stat(session.ChunksDir); !os.IsNotExist(err) {
		t.Error("direktori chunk masih ada setelah Remove")
	}
}

// TestSessionStoreSweepExpired memastikan sweep membuang sesi yang lewat TTL
// beserta direktori chunknya. Ini yang mencegah sesi dan disk menumpuk.
func TestSessionStoreSweepExpired(t *testing.T) {
	store := NewSessionStoreWithTTL(time.Hour, time.Hour)
	defer store.Close()

	// TTL sangat pendek supaya sesi langsung dianggap basi.
	expiredStore := NewSessionStoreWithTTL(time.Nanosecond, time.Hour)
	defer expiredStore.Close()

	session, err := expiredStore.Put("basah", "video.mp4", 1, 128, t.TempDir())
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	time.Sleep(2 * time.Millisecond)

	expiredStore.sweep(time.Now())

	if _, ok := expiredStore.Get("basah"); ok {
		t.Error("sesi basi seharusnya dibuang oleh sweep")
	}
	if _, err := os.Stat(session.ChunksDir); !os.IsNotExist(err) {
		t.Error("direktori chunk sesi basi seharusnya dihapus")
	}

	// Store dengan TTL panjang tidak boleh ikut membuang sesi.
	if _, err := store.Put("segar", "video.mp4", 1, 128, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	store.sweep(time.Now())
	if _, ok := store.Get("segar"); !ok {
		t.Error("sesi yang belum kedaluwarsa ikut dibuang")
	}
}

// TestSessionStoreCloseIdempotent memastikan Close aman dipanggil berulang.
func TestSessionStoreCloseIdempotent(t *testing.T) {
	store := NewSessionStoreWithTTL(time.Hour, time.Hour)
	store.Close()
	store.Close()
}

// TestSessionStoreSweepLoopStops memastikan goroutine sweep berhenti setelah
// Close, sehingga tidak ada goroutine yatim saat server shutdown.
func TestSessionStoreSweepLoopStops(t *testing.T) {
	store := NewSessionStoreWithTTL(time.Hour, time.Nanosecond)
	store.Close()

	before := runtimeGoroutineCount()
	// Beri waktu bagi goroutine sweep untuk keluar setelah Close.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if runtimeGoroutineCount() <= before {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("goroutine sweep tidak berhenti (sebelum=%d, sesudah=%d)", before, runtimeGoroutineCount())
}
