package event

import "time"

// Type mirrors Aurora Sysmon IDs + extended
type Type string

const (
	TypeProcessCreation   Type = "process_creation"   // Sysmon 1
	TypeFileTime          Type = "file_time"          // Sysmon 2
	TypeNetworkConnection Type = "network_connection" // Sysmon 3
	TypeFileEvent         Type = "file_event"         // Sysmon 11
	TypeBPFEvent          Type = "bpf_event"          // Sysmon 100
	TypePrivEsc           Type = "privilege_change"
	TypeFileRename        Type = "file_rename"
	TypeChmod             Type = "file_chmod"
	TypeMemFD             Type = "memfd_create"
	TypeModule            Type = "module_load"
	TypeFSNotify          Type = "fsnotify"
)

// Event is normalized cross-provider. Compatible with Aurora distributor fields.
type Event struct {
	Timestamp       time.Time `json:"timestamp"`
	Host            string    `json:"host"`
	Type            Type      `json:"type"`
	EventID         int       `json:"event_id"`

	Image           string `json:"image"`
	CommandLine     string `json:"command_line"`
	CurrentDirectory string `json:"current_directory"`
	User            string `json:"user"`
	LogonID         string `json:"logon_id"`

	ParentImage       string `json:"parent_image"`
	ParentCommandLine string `json:"parent_command_line"`
	ParentPID         int    `json:"parent_pid"`

	ProcessID       int    `json:"process_id"`
	ParentProcessID int    `json:"parent_process_id"`

	TargetFilename string `json:"target_filename,omitempty"`
	FileAction     string `json:"file_action,omitempty"`

	DestinationIP   string `json:"destination_ip,omitempty"`
	DestinationPort int    `json:"destination_port,omitempty"`
	Protocol        string `json:"protocol,omitempty"`
	Initiated       bool   `json:"initiated,omitempty"`

	BPFCommand      string `json:"bpf_command,omitempty"`
	Cgroup          string `json:"cgroup,omitempty"`
	Container       string `json:"container,omitempty"`

	RawAudit map[string]string `json:"raw_audit,omitempty"`
}

// Alert is sigma/ioc match
type Alert struct {
	Timestamp time.Time `json:"timestamp"`
	Host      string    `json:"host"`
	RuleID    string    `json:"rule_id"`
	RuleName  string    `json:"rule_name"`
	Level     string    `json:"level"` // info/low/medium/high/critical
	MitreID   string    `json:"mitre_id,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	Event     Event     `json:"event"`
}
