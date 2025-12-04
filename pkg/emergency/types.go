package emergency

import (
	"net"
	"time"
)

// Priority represents broadcast priority level
type Priority int

const (
	PriorityNormal   Priority = 0 // Normal communications
	PriorityHigh     Priority = 1 // Important but not critical
	PriorityEmergency Priority = 2 // Emergency coordination
	PriorityCritical Priority = 3 // Active emergency
)

// String returns the string representation of priority
func (p Priority) String() string {
	switch p {
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityEmergency:
		return "emergency"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ParsePriority parses a priority string
func ParsePriority(s string) Priority {
	switch s {
	case "critical":
		return PriorityCritical
	case "emergency":
		return PriorityEmergency
	case "high":
		return PriorityHigh
	case "normal":
		return PriorityNormal
	default:
		return PriorityNormal
	}
}

// AutoTuneMode represents auto-tune behavior
type AutoTuneMode int

const (
	AutoTuneNever  AutoTuneMode = 0 // Never auto-tune
	AutoTunePrompt AutoTuneMode = 1 // Prompt user
	AutoTuneAlways AutoTuneMode = 2 // Always auto-tune
)

// String returns the string representation of auto-tune mode
func (a AutoTuneMode) String() string {
	switch a {
	case AutoTuneNever:
		return "never"
	case AutoTunePrompt:
		return "prompt"
	case AutoTuneAlways:
		return "always"
	default:
		return "never"
	}
}

// ParseAutoTuneMode parses an auto-tune mode string
func ParseAutoTuneMode(s string) AutoTuneMode {
	switch s {
	case "always":
		return AutoTuneAlways
	case "prompt":
		return AutoTunePrompt
	case "never":
		return AutoTuneNever
	default:
		return AutoTuneNever
	}
}

// EmergencyNotification represents an emergency broadcast notification
type EmergencyNotification struct {
	Channel   string
	Priority  Priority
	Callsign  string
	IPv6      net.IP
	Port      int
	Timestamp time.Time
	Message   string
}

// EmergencySettings holds user preferences for emergency handling
type EmergencySettings struct {
	AutoTuneMode     AutoTuneMode // How to handle auto-tune
	CriticalChannels []string     // Which channels trigger auto-tune
	VisualAlerts     bool         // Show visual alerts/banners
	AudioAlerts      bool         // Play alert sounds
	SaveChannel      bool         // Remember channel before emergency
	AutoReturn       bool         // Return to saved channel when emergency ends
}

// DefaultSettings returns default emergency settings
func DefaultSettings() EmergencySettings {
	return EmergencySettings{
		AutoTuneMode:     AutoTunePrompt, // Prompt by default
		CriticalChannels: []string{"emergency", "netcontrol", "medical", "sar"},
		VisualAlerts:     true,
		AudioAlerts:      false, // No audio alerts by default
		SaveChannel:      true,
		AutoReturn:       false, // User must manually return
	}
}

// ShouldAutoTune determines if listener should auto-tune for this notification
func (s EmergencySettings) ShouldAutoTune(notif EmergencyNotification) bool {
	// Never auto-tune if disabled
	if s.AutoTuneMode == AutoTuneNever {
		return false
	}

	// Check if this channel is in critical channels list
	isCriticalChannel := false
	for _, ch := range s.CriticalChannels {
		if ch == notif.Channel {
			isCriticalChannel = true
			break
		}
	}

	// Only auto-tune for critical channels
	if !isCriticalChannel {
		return false
	}

	// Critical priority: auto-tune if mode is Always
	if notif.Priority >= PriorityCritical {
		return s.AutoTuneMode == AutoTuneAlways
	}

	// Emergency priority: auto-tune if mode is Always
	if notif.Priority >= PriorityEmergency {
		return s.AutoTuneMode == AutoTuneAlways
	}

	// High/Normal: never auto-tune
	return false
}

// NeedsPrompt determines if user should be prompted to switch
func (s EmergencySettings) NeedsPrompt(notif EmergencyNotification) bool {
	if s.AutoTuneMode != AutoTunePrompt {
		return false
	}

	// Prompt for critical or emergency priority
	return notif.Priority >= PriorityEmergency
}
