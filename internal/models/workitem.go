package models

import (
	"fmt"
	"time"
)

type Status string

const (
	StatusNotStarted Status = "not_started"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusBlocked    Status = "blocked"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusNotStarted, StatusInProgress, StatusDone, StatusBlocked:
		return true
	}
	return false
}

var ValidRoles = map[string]string{
	"lead":        "human",
	"api_dev":     "agent",
	"ui_dev":      "agent",
	"backend_dev": "agent",
	"security":    "agent",
	"unit_test":   "agent",
	"integration": "agent",
	"data_model":  "agent",
	"scenario":    "agent",
}

func IsValidRole(role string) bool {
	_, ok := ValidRoles[role]
	return ok
}

type WorkItem struct {
	ID              string     `bson:"_id" json:"id" yaml:"id"`
	Title           string     `bson:"title" json:"title" yaml:"title"`
	Phase           int        `bson:"phase" json:"phase" yaml:"phase"`
	Owner           string     `bson:"owner" json:"owner" yaml:"owner"`
	Role            string     `bson:"role" json:"role" yaml:"role"`
	Status          Status     `bson:"status" json:"status"`
	Dependencies    []string   `bson:"dependencies" json:"dependencies" yaml:"dependencies"`
	ClaimedBy       *string    `bson:"claimed_by,omitempty" json:"claimed_by"`
	StartedAt       *time.Time `bson:"started_at,omitempty" json:"started_at"`
	CompletedAt     *time.Time `bson:"completed_at,omitempty" json:"completed_at"`
	DurationSeconds *int64     `bson:"duration_seconds,omitempty" json:"duration_seconds"`
	BlockerNote     *string    `bson:"blocker_note,omitempty" json:"blocker_note"`
	Project         string     `bson:"project" json:"project"`
	UpdatedAt       time.Time  `bson:"updated_at" json:"updated_at"`
}

type Phase struct {
	ID      string `bson:"_id" json:"id"`
	Number  int    `bson:"number" json:"number" yaml:"number"`
	Name    string `bson:"name" json:"name" yaml:"name"`
	Target  string `bson:"target" json:"target" yaml:"target"`
	Project string `bson:"project" json:"project"`
}

func (p Phase) CompositeID(project string) string {
	return fmt.Sprintf("%s:%d", project, p.Number)
}

type SeedFile struct {
	Project   string         `yaml:"project"`
	Phases    []SeedPhase    `yaml:"phases"`
	WorkItems []SeedWorkItem `yaml:"work_items"`
}

type SeedPhase struct {
	Number int    `yaml:"number"`
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
}

type SeedWorkItem struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	Phase        int      `yaml:"phase"`
	Owner        string   `yaml:"owner"`
	Role         string   `yaml:"role"`
	Dependencies []string `yaml:"dependencies"`
}

func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", minutes)
}
