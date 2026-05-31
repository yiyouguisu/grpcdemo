package auditlog

const (
	defaultLogDir     = "./logs"
	defaultRetainDays = 7
)

// AuditLogConfig holds configuration for the audit log system.
type AuditLogConfig struct {
	// LogDir is the root directory for audit log files.
	LogDir string
	// RetainDays is the number of days to retain log files before cleanup.
	RetainDays int
}

// DefaultConfig returns the default audit log configuration.
// Default log directory is "./logs" and default retention is 7 days.
func DefaultConfig() AuditLogConfig {
	return AuditLogConfig{
		LogDir:     defaultLogDir,
		RetainDays: defaultRetainDays,
	}
}
