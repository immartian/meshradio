package mdns

import (
	"fmt"

	"github.com/grandcat/zeroconf"
)

// Advertiser advertises a MeshRadio service via mDNS
type Advertiser struct {
	server *zeroconf.Server
	info   ServiceInfo
}

// NewAdvertiser creates a new mDNS advertiser
func NewAdvertiser(info ServiceInfo) (*Advertiser, error) {
	// Validate required fields
	if info.Name == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if info.Port == 0 {
		return nil, fmt.Errorf("port is required")
	}
	if info.Callsign == "" {
		return nil, fmt.Errorf("callsign is required")
	}

	// Set defaults if not provided
	if info.Group == "" {
		info.Group = ChannelCommunity
	}
	if info.Channel == "" {
		info.Channel = ChannelCommunity
	}
	if info.Priority == "" {
		info.Priority = PriorityNormal
	}
	if info.Codec == "" {
		info.Codec = "opus"
	}
	if info.Bitrate == 0 {
		info.Bitrate = 64
	}

	// Create TXT records
	txt := CreateTXTRecord(info)

	// Register mDNS service (nil = advertise on all interfaces)
	server, err := zeroconf.Register(
		info.Name,      // Instance name (e.g., "STATION1")
		ServiceType,    // Service type (_meshradio._udp)
		"local.",       // Domain
		info.Port,      // Port
		txt,            // TXT records
		nil,            // Advertise on all network interfaces
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register mDNS service: %w", err)
	}

	return &Advertiser{
		server: server,
		info:   info,
	}, nil
}

// UpdateTXT updates the TXT records of the advertised service
func (a *Advertiser) UpdateTXT(info ServiceInfo) error {
	// Note: zeroconf v1.0.0 doesn't support updating TXT records
	// This would require re-registering the service
	// For now, return an error indicating this limitation
	return fmt.Errorf("updating TXT records requires re-advertising the service")
}

// Shutdown stops advertising the service
func (a *Advertiser) Shutdown() error {
	if a.server != nil {
		a.server.Shutdown()
	}
	return nil
}

// GetInfo returns the service info
func (a *Advertiser) GetInfo() ServiceInfo {
	return a.info
}
