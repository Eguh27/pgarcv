package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DefaultSessionTTL adalah batas umur sesi upload. Sesi yang sudah lewat TTL
// dibersihkan supaya tidak menumpuk di memori dan meninggalkan direktori chunk
// yatim di disk.
const DefaultSessionTTL = 6 * time.Hour

// sweepInterval adalah jarak antar-pembersihan otomatis.
const sweepInterval = 5 * time.Minute

// Session adalah state satu upload chunked yang sedang berjalan.
type Session struct {
	UploadID    string
	Filename    string
	TotalChunks int
	TotalSize   int64
	ChunksDir   string

	mu         sync.Mutex
	chunks     map[int]bool
	lastActive time.Time
}

// newSession membuat sesi baru yang siap menyimpan chunk.
func newSession(uploadID, filename string, totalChunks int, totalSize int64, chunksDir string) *Session {
	return &Session{
		UploadID:    uploadID,
		Filename:    filename,
		TotalChunks: totalChunks,
		TotalSize:   totalSize,
		ChunksDir:   chunksDir,
		chunks:      make(map[int]bool, totalChunks),
		lastActive:  time.Now(),
	}
}

// MarkReceived menandai sebuah chunk sudah tersimpan dan mengembalikan jumlah
// chunk yang sudah diterima.
func (s *Session) MarkReceived(chunkIndex int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chunks[chunkIndex] = true
	s.lastActive = time.Now()
	return len(s.chunks)
}

// Received mengembalikan jumlah chunk yang sudah diterima.
func (s *Session) Received() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.chunks)
}

// IsComplete melapor apakah semua chunk sudah masuk.
func (s *Session) IsComplete() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.chunks) == s.TotalChunks
}

// expired melapor apakah sesi sudah tidak aktif selama ttl.
func (s *Session) expired(now time.Time, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return now.Sub(s.lastActive) > ttl
}

// SessionStore menyimpan sesi upload chunked dan membersihkan sesi yang sudah
// kedaluwarsa di latar belakang.
type SessionStore struct {
	sessions sync.Map // uploadID -> *Session

	ttl         time.Duration
	sweepPeriod time.Duration

	stopOnce sync.Once
	stopCh   chan struct{}
}

// NewSessionStore membuat store yang menjalankan pembersih otomatis.
func NewSessionStore() *SessionStore {
	return NewSessionStoreWithTTL(DefaultSessionTTL, sweepInterval)
}

// NewSessionStoreWithTTL membuat store dengan TTL dan interval sweep
// eksplisit. Nilai non-positif memakai default.
func NewSessionStoreWithTTL(ttl, sweep time.Duration) *SessionStore {
	if ttl <= 0 {
		ttl = DefaultSessionTTL
	}
	if sweep <= 0 {
		sweep = sweepInterval
	}

	s := &SessionStore{
		ttl:         ttl,
		sweepPeriod: sweep,
		stopCh:      make(chan struct{}),
	}
	go s.sweepLoop()
	return s
}

// Put mendaftarkan sesi baru dan mengembalikan direktori tempat chunk disimpan.
func (s *SessionStore) Put(uploadID, filename string, totalChunks int, totalSize int64, uploadRoot string) (*Session, error) {
	chunksDir := filepath.Join(uploadRoot, ".chunks", uploadID)
	if err := os.MkdirAll(chunksDir, 0o755); err != nil {
		return nil, fmt.Errorf("gagal menyiapkan direktori chunk: %w", err)
	}

	session := newSession(uploadID, filename, totalChunks, totalSize, chunksDir)
	s.sessions.Store(uploadID, session)
	return session, nil
}

// Get mengambil sesi berdasarkan uploadID.
func (s *SessionStore) Get(uploadID string) (*Session, bool) {
	v, ok := s.sessions.Load(uploadID)
	if !ok {
		return nil, false
	}
	session, ok := v.(*Session)
	return session, ok
}

// Remove menghapus sesi dari store dan membersihkan direktori chunk-nya.
// Aman dipanggil untuk sesi yang sudah tidak ada.
func (s *SessionStore) Remove(uploadID string) {
	if v, ok := s.sessions.LoadAndDelete(uploadID); ok {
		if session, ok := v.(*Session); ok {
			_ = os.RemoveAll(session.ChunksDir)
		}
	}
}

// Len menghitung jumlah sesi yang aktif — dipakai oleh test dan monitoring.
func (s *SessionStore) Len() int {
	n := 0
	s.sessions.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}

// Close menghentikan goroutine pembersih. Aman dipanggil lebih dari sekali.
func (s *SessionStore) Close() {
	s.stopOnce.Do(func() { close(s.stopCh) })
}

// sweepLoop menghapus sesi yang lewat TTL secara berkala sampai Close() dipanggil.
func (s *SessionStore) sweepLoop() {
	ticker := time.NewTicker(s.sweepPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.sweep(time.Now())
		case <-s.stopCh:
			return
		}
	}
}

// sweep membuang sesi yang sudah kedaluwarsa beserta direktori chunknya.
func (s *SessionStore) sweep(now time.Time) {
	s.sessions.Range(func(key, value any) bool {
		session, ok := value.(*Session)
		if !ok {
			s.sessions.Delete(key)
			return true
		}
		if session.expired(now, s.ttl) {
			s.Remove(key.(string))
		}
		return true
	})
}
