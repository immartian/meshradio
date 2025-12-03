package mdns

import "net"

// ServiceInfo represents a discovered MeshRadio service
type ServiceInfo struct {
	Name     string // Service instance name (e.g., "STATION1")
	Host     string // Hostname
	Port     int    // RTP port
	IPv6     net.IP // IPv6 address

	// TXT record fields
	Group    string // Multicast group label (emergency, community, talk)
	Channel  string // Channel type (emergency, community, etc.)
	Callsign string // Station identifier (e.g., W1EMERGENCY)
	Priority string // Priority level (normal, high, emergency, critical)
	Codec    string // Audio codec (opus)
	Bitrate  int    // Bitrate in kbps
}

// Priority constants
const (
	PriorityNormal    = "normal"
	PriorityHigh      = "high"
	PriorityEmergency = "emergency"
	PriorityCritical  = "critical"
)

// Channel/Group constants
const (
	ChannelEmergency  = "emergency"
	ChannelCommunity  = "community"
	ChannelTalk       = "talk"
)

// ServiceType is the mDNS service type for MeshRadio
const ServiceType = "_meshradio._udp"
