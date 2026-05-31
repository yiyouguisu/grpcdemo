package auditlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// createTestDir creates a temporary directory for testing and returns its path
// and a cleanup function.
func createTestDir(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	return dir, func() { os.RemoveAll(dir) }
}

// writeTestLogFile creates a log file with the given date string in the service directory.
func writeTestLogFile(t *testing.T, baseDir, serviceName, dateStr string) {
	t.Helper()
	dir := filepath.Join(baseDir, serviceName)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		t.Fatalf("failed to create service dir: %v", err)
	}
	path := filepath.Join(dir, dateStr+".log")
	if err := os.WriteFile(path, []byte("test log entry\n"), filePerm); err != nil {
		t.Fatalf("failed to write test log file: %v", err)
	}
}

// fixedClock returns a NowFunc that always returns the given time.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// newTestFileManager creates a FileManager configured for testing with a fixed
// clock and short cleanup interval. It calls Start() and registers cleanup via t.Cleanup.
func newTestFileManager(t *testing.T, cfg AuditLogConfig, now time.Time) *FileManager {
	t.Helper()
	fm := NewFileManager(cfg)
	fm.NowFunc = fixedClock(now)
	fm.CleanupInterval = time.Hour
	fm.Start()
	t.Cleanup(func() { fm.Close() })
	return fm
}

func TestFileManager_AutoCreateDirAndFile(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	fm := newTestFileManager(t, cfg, now)

	writer, err := fm.GetLogWriter("ChatService")
	if err != nil {
		t.Fatalf("GetLogWriter failed: %v", err)
	}

	if writer == nil {
		t.Fatal("expected non-nil writer")
	}

	expectedPath := filepath.Join(dir, "ChatService", "2024-01-15.log")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected log file %s to exist", expectedPath)
	}

	expectedDir := filepath.Join(dir, "ChatService")
	if info, err := os.Stat(expectedDir); os.IsNotExist(err) || !info.IsDir() {
		t.Errorf("expected directory %s to exist", expectedDir)
	}
}

func TestFileManager_MultipleServicesIndependentDirs(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	fm := newTestFileManager(t, cfg, now)

	services := []string{"ChatService", "StreamService", "AdminService"}
	for _, svc := range services {
		writer, err := fm.GetLogWriter(svc)
		if err != nil {
			t.Fatalf("GetLogWriter(%s) failed: %v", svc, err)
		}
		if writer == nil {
			t.Fatalf("GetLogWriter(%s) returned nil writer", svc)
		}
	}

	for _, svc := range services {
		expectedPath := filepath.Join(dir, svc, "2024-01-15.log")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected log file %s to exist", expectedPath)
		}
	}
}

func TestFileManager_DailyFileRotation(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	day1 := time.Date(2024, 1, 14, 10, 0, 0, 0, time.UTC)
	fm := newTestFileManager(t, cfg, day1)

	writer, err := fm.GetLogWriter("ChatService")
	if err != nil {
		t.Fatalf("GetLogWriter failed: %v", err)
	}
	if writer == nil {
		t.Fatal("expected non-nil writer")
	}

	oldPath := filepath.Join(dir, "ChatService", "2024-01-14.log")
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		t.Errorf("expected old log file %s to exist", oldPath)
	}

	day2 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fm.NowFunc = fixedClock(day2)

	writer2, err := fm.GetLogWriter("ChatService")
	if err != nil {
		t.Fatalf("GetLogWriter after date change failed: %v", err)
	}
	if writer2 == nil {
		t.Fatal("expected non-nil writer after date change")
	}

	newPath := filepath.Join(dir, "ChatService", "2024-01-15.log")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("expected new log file %s to exist", newPath)
	}
}

func TestFileManager_ConcurrentWritesNoCorruption(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fm := newTestFileManager(t, cfg, now)

	const goroutines = 50
	const writesPerGoroutine = 20
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				writer, err := fm.GetLogWriter("ChatService")
				if err != nil {
					t.Errorf("goroutine %d: GetLogWriter failed: %v", id, err)
					return
				}
				line := fmt.Sprintf("goroutine-%d-write-%d\n", id, j)
				if _, err := writer.Write([]byte(line)); err != nil {
					t.Errorf("goroutine %d: Write failed: %v", id, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	logPath := filepath.Join(dir, "ChatService", "2024-01-15.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	expectedLines := goroutines * writesPerGoroutine
	if len(lines) != expectedLines {
		t.Errorf("expected %d log lines, got %d", expectedLines, len(lines))
	}

	for _, line := range lines {
		if !strings.HasPrefix(line, "goroutine-") {
			t.Errorf("corrupted log line: %q", line)
		}
	}
}

func TestFileManager_CleanupRemovesExpiredFiles(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	writeTestLogFile(t, dir, "ChatService", "2024-01-05")
	writeTestLogFile(t, dir, "ChatService", "2024-01-06")
	writeTestLogFile(t, dir, "ChatService", "2024-01-07")
	writeTestLogFile(t, dir, "ChatService", "2024-01-08")
	writeTestLogFile(t, dir, "StreamService", "2024-01-01")
	writeTestLogFile(t, dir, "StreamService", "2024-01-10")

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fm := newTestFileManager(t, cfg, now)

	fm.cleanup()

	// cutoff = 2024-01-15 - 7 days = 2024-01-08; files before cutoff are deleted.
	expiredFiles := []string{
		filepath.Join(dir, "ChatService", "2024-01-05.log"),
		filepath.Join(dir, "ChatService", "2024-01-06.log"),
		filepath.Join(dir, "ChatService", "2024-01-07.log"),
		filepath.Join(dir, "StreamService", "2024-01-01.log"),
	}
	for _, path := range expiredFiles {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("expected expired file %s to be deleted", path)
		}
	}

	// 2024-01-08 is the cutoff boundary; it should be kept.
	keptFiles := []string{
		filepath.Join(dir, "ChatService", "2024-01-08.log"),
		filepath.Join(dir, "StreamService", "2024-01-10.log"),
	}
	for _, path := range keptFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected recent file %s to be preserved", path)
		}
	}
}

func TestFileManager_CleanupPreservesRecentFiles(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	writeTestLogFile(t, dir, "ChatService", "2024-01-08")
	writeTestLogFile(t, dir, "ChatService", "2024-01-09")
	writeTestLogFile(t, dir, "ChatService", "2024-01-15")

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fm := newTestFileManager(t, cfg, now)

	fm.cleanup()

	recentFiles := []string{
		filepath.Join(dir, "ChatService", "2024-01-08.log"),
		filepath.Join(dir, "ChatService", "2024-01-09.log"),
		filepath.Join(dir, "ChatService", "2024-01-15.log"),
	}
	for _, path := range recentFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected recent file %s to be preserved", path)
		}
	}
}

func TestFileManager_CleanupRemovesEmptyServiceDir(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	writeTestLogFile(t, dir, "OldService", "2024-01-01")

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fm := newTestFileManager(t, cfg, now)

	fm.cleanup()

	oldDir := filepath.Join(dir, "OldService")
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Errorf("expected empty service directory %s to be removed", oldDir)
	}
}

func TestFileManager_CloseFlushesAndClosesFiles(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fm := NewFileManager(cfg)
	fm.NowFunc = fixedClock(now)
	fm.CleanupInterval = 50 * time.Millisecond
	fm.Start()

	writer, err := fm.GetLogWriter("ChatService")
	if err != nil {
		t.Fatalf("GetLogWriter failed: %v", err)
	}

	if _, err := writer.Write([]byte("test entry\n")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := fm.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	logPath := filepath.Join(dir, "ChatService", "2024-01-15.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file after Close: %v", err)
	}
	if string(data) != "test entry\n" {
		t.Errorf("expected 'test entry\\n', got %q", string(data))
	}
}

func TestFileManager_SameDayReturnsSameWriter(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fm := newTestFileManager(t, cfg, now)

	writer1, err := fm.GetLogWriter("ChatService")
	if err != nil {
		t.Fatalf("first GetLogWriter failed: %v", err)
	}

	writer2, err := fm.GetLogWriter("ChatService")
	if err != nil {
		t.Fatalf("second GetLogWriter failed: %v", err)
	}

	if writer1 != writer2 {
		t.Error("expected same writer for same service on same day")
	}
}

func TestFileManager_CleanupTickerRuns(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	// Create an expired log file.
	writeTestLogFile(t, dir, "ChatService", "2024-01-01")

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	fm := NewFileManager(cfg)
	fm.NowFunc = fixedClock(now)
	fm.CleanupInterval = 50 * time.Millisecond
	fm.Start()
	defer fm.Close()

	// Wait for at least one cleanup tick.
	time.Sleep(200 * time.Millisecond)

	expiredPath := filepath.Join(dir, "ChatService", "2024-01-01.log")
	if _, err := os.Stat(expiredPath); !os.IsNotExist(err) {
		t.Errorf("expected expired file to be cleaned up by ticker")
	}
}

func TestFileManager_StartIdempotent(t *testing.T) {
	dir, _ := createTestDir(t)
	cfg := AuditLogConfig{LogDir: dir, RetainDays: 7}

	fm := NewFileManager(cfg)
	fm.CleanupInterval = time.Hour
	// Calling Start multiple times should not panic or spawn extra goroutines.
	fm.Start()
	fm.Start()
	fm.Start()
	fm.Close()
}
