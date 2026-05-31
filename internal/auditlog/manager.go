package auditlog

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	filePerm = 0644
	dirPerm  = 0755
)

// serviceEntry holds the file handle and metadata for a single service's log.
type serviceEntry struct {
	mu   sync.Mutex
	file *os.File
	date string
}

// FileManager manages audit log files organized by service name and date.
// It handles automatic directory creation, daily file rotation, concurrent-safe
// writes, and background cleanup of expired log files.
//
// Usage:
//
//	fm := NewFileManager(config)
//	fm.NowFunc = time.Now       // optional, defaults to time.Now
//	fm.CleanupInterval = time.Hour // optional, defaults to time.Hour
//	fm.Start()                   // starts background cleanup goroutine
//	defer fm.Close()
type FileManager struct {
	config    AuditLogConfig
	entries   map[string]*serviceEntry
	entriesMu sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc

	// NowFunc returns the current time. Defaults to time.Now().
	// Override in tests for deterministic behavior. Must be set before Start().
	NowFunc func() time.Time

	// CleanupInterval controls how often the cleanup goroutine runs.
	// Defaults to 1 hour. Override in tests for faster cleanup cycles.
	// Must be set before Start().
	CleanupInterval time.Duration

	started bool
}

// NewFileManager creates a new FileManager with the given config.
// Call Start() to begin the background cleanup goroutine, and Close() to
// stop it and release all file handles.
func NewFileManager(config AuditLogConfig) *FileManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &FileManager{
		config:          config,
		entries:         make(map[string]*serviceEntry),
		ctx:             ctx,
		cancel:          cancel,
		NowFunc:         time.Now,
		CleanupInterval: time.Hour,
	}
}

// Start begins the background cleanup goroutine. It must be called after
// setting NowFunc and CleanupInterval, and before any concurrent use of
// the FileManager. Calling Start multiple times is safe; only the first
// call takes effect.
func (fm *FileManager) Start() {
	if fm.started {
		return
	}
	fm.started = true
	go fm.cleanupLoop()
}

// GetLogWriter returns an io.Writer for the given service name. It creates
// the service directory and log file if they don't exist, and automatically
// rotates to a new file when the date changes. The returned writer is safe
// for concurrent use.
func (fm *FileManager) GetLogWriter(serviceName string) (io.Writer, error) {
	entry := fm.getOrCreateEntry(serviceName)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	today := fm.NowFunc().Format("2006-01-02")
	if entry.file != nil && entry.date == today {
		return entry.file, nil
	}

	if entry.file != nil {
		entry.file.Close()
	}

	dir := filepath.Join(fm.config.LogDir, serviceName)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	path := filepath.Join(dir, today+".log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePerm)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	entry.file = f
	entry.date = today
	return f, nil
}

// getOrCreateEntry returns the serviceEntry for the given service,
// creating one if it doesn't exist. Uses double-check locking pattern.
func (fm *FileManager) getOrCreateEntry(serviceName string) *serviceEntry {
	fm.entriesMu.RLock()
	entry, ok := fm.entries[serviceName]
	fm.entriesMu.RUnlock()
	if ok {
		return entry
	}

	fm.entriesMu.Lock()
	defer fm.entriesMu.Unlock()
	entry, ok = fm.entries[serviceName]
	if ok {
		return entry
	}
	entry = &serviceEntry{}
	fm.entries[serviceName] = entry
	return entry
}

// Close stops the background cleanup goroutine and closes all open file handles.
// Returns the first error encountered while closing files, if any.
func (fm *FileManager) Close() error {
	fm.cancel()

	fm.entriesMu.Lock()
	defer fm.entriesMu.Unlock()

	var firstErr error
	for _, entry := range fm.entries {
		entry.mu.Lock()
		if entry.file != nil {
			if err := entry.file.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
			entry.file = nil
		}
		entry.mu.Unlock()
	}
	return firstErr
}

// cleanupLoop runs the cleanup ticker until the context is cancelled.
func (fm *FileManager) cleanupLoop() {
	ticker := time.NewTicker(fm.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			fm.cleanup()
		}
	}
}

// cleanup removes log files older than RetainDays and deletes empty service directories.
func (fm *FileManager) cleanup() {
	now := fm.NowFunc()
	// Normalize to midnight so cutoff aligns with file date boundaries.
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	cutoff := today.AddDate(0, 0, -fm.config.RetainDays)

	serviceDirs, err := os.ReadDir(fm.config.LogDir)
	if err != nil {
		return
	}

	for _, serviceDir := range serviceDirs {
		if !serviceDir.IsDir() {
			continue
		}
		fm.cleanupServiceDir(filepath.Join(fm.config.LogDir, serviceDir.Name()), cutoff)
	}
}

// cleanupServiceDir removes expired log files in a single service directory
// and deletes the directory if it becomes empty.
func (fm *FileManager) cleanupServiceDir(servicePath string, cutoff time.Time) {
	files, err := os.ReadDir(servicePath)
	if err != nil {
		return
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !strings.HasSuffix(name, ".log") {
			continue
		}
		dateStr := strings.TrimSuffix(name, ".log")
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		if fileDate.Before(cutoff) {
			os.Remove(filepath.Join(servicePath, name))
		}
	}

	remaining, err := os.ReadDir(servicePath)
	if err == nil && len(remaining) == 0 {
		os.Remove(servicePath)
	}
}
