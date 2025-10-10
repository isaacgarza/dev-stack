package types

import "time"

// StartOptions defines options for starting services
type StartOptions struct {
	Build         bool
	ForceRecreate bool
	NoDeps        bool
	Detach        bool
	Timeout       time.Duration
}

// StopOptions defines options for stopping services
type StopOptions struct {
	Timeout       int
	Remove        bool
	RemoveVolumes bool
}

// ExecOptions defines options for executing commands in containers
type ExecOptions struct {
	User        string
	WorkingDir  string
	Env         []string
	Interactive bool
	TTY         bool
	Detach      bool
}

// LogOptions defines options for retrieving container logs
type LogOptions struct {
	Follow     bool
	Timestamps bool
	Tail       string
	Since      string
}

// ConnectOptions defines options for connecting to services
type ConnectOptions struct {
	User     string
	Database string
	Host     string
	Port     string
	ReadOnly bool
}

// ScaleOptions defines options for scaling services
type ScaleOptions struct {
	Detach     bool
	Timeout    time.Duration
	NoRecreate bool
}

// BackupOptions defines options for backing up service data
type BackupOptions struct {
	OutputDir string
	Compress  bool
	Format    string
	Database  string
	User      string
	NoOwner   bool
	Clean     bool
}

// RestoreOptions defines options for restoring service data
type RestoreOptions struct {
	Database          string
	User              string
	Clean             bool
	CreateDB          bool
	DropDB            bool
	SingleTransaction bool
}

// CleanupOptions defines options for cleaning up resources
type CleanupOptions struct {
	RemoveVolumes  bool
	RemoveImages   bool
	RemoveNetworks bool
	All            bool
	DryRun         bool
}
