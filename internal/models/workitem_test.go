package models

import (
	"testing"
	"time"
)

func TestStatusIsValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"not_started", StatusNotStarted, true},
		{"in_progress", StatusInProgress, true},
		{"done", StatusDone, true},
		{"blocked", StatusBlocked, true},
		{"empty string", Status(""), false},
		{"unknown string", Status("unknown"), false},
		{"cancelled", Status("cancelled"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("Status(%q).IsValid() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{"lead", "lead", true},
		{"api_dev", "api_dev", true},
		{"ui_dev", "ui_dev", true},
		{"backend_dev", "backend_dev", true},
		{"security", "security", true},
		{"unit_test", "unit_test", true},
		{"integration", "integration", true},
		{"data_model", "data_model", true},
		{"scenario", "scenario", true},
		{"empty", "", false},
		{"unknown", "unknown", false},
		{"admin", "admin", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRole(tt.role); got != tt.want {
				t.Errorf("IsValidRole(%q) = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"30 seconds", 30 * time.Second, "30s"},
		{"5 minutes", 5 * time.Minute, "5m"},
		{"2h 15m", 2*time.Hour + 15*time.Minute, "2h 15m"},
		{"3 hours exactly", 3 * time.Hour, "3h"},
		{"1d 4h", 28 * time.Hour, "1d 4h"},
		{"2 days exactly", 48 * time.Hour, "2d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatDuration(tt.duration); got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestPhaseCompositeID(t *testing.T) {
	tests := []struct {
		name    string
		phase   Phase
		project string
		want    string
	}{
		{"basic", Phase{Number: 3, Name: "Implementation"}, "acme", "acme:3"},
		{"phase zero", Phase{Number: 0, Name: "Init"}, "proj", "proj:0"},
		{"large number", Phase{Number: 42, Name: "Finalize"}, "myapp", "myapp:42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.phase.CompositeID(tt.project); got != tt.want {
				t.Errorf("Phase.CompositeID(%q) = %q, want %q", tt.project, got, tt.want)
			}
		})
	}
}
