package auditlog

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.LogDir != "./logs" {
		t.Errorf("expected LogDir './logs', got %q", cfg.LogDir)
	}
	if cfg.RetainDays != 7 {
		t.Errorf("expected RetainDays 7, got %d", cfg.RetainDays)
	}
}

func TestAuditLogConfig_CustomValues(t *testing.T) {
	cfg := AuditLogConfig{
		LogDir:     "/var/log/grpcdemo",
		RetainDays: 30,
	}

	if cfg.LogDir != "/var/log/grpcdemo" {
		t.Errorf("expected LogDir '/var/log/grpcdemo', got %q", cfg.LogDir)
	}
	if cfg.RetainDays != 30 {
		t.Errorf("expected RetainDays 30, got %d", cfg.RetainDays)
	}
}
