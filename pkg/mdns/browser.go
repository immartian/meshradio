package mdns

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/grandcat/zeroconf"
)

// Browser browses for MeshRadio services via mDNS
type Browser struct {
	resolver *zeroconf.Resolver
	services map[string]ServiceInfo // Key: instance name
}

// BrowseOptions configures browsing behavior
type BrowseOptions struct {
	Timeout       time.Duration // How long to browse (default: 3 seconds)
	FilterChannel string        // Filter by channel (empty = all)
	FilterPriority string       // Filter by priority (empty = all)
}

// NewBrowser creates a new mDNS browser
func NewBrowser() (*Browser, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create mDNS resolver: %w", err)
	}

	return &Browser{
		resolver: resolver,
		services: make(map[string]ServiceInfo),
	}, nil
}

// Browse browses for MeshRadio services
func (b *Browser) Browse(opts BrowseOptions) ([]ServiceInfo, error) {
	// Set default timeout
	if opts.Timeout == 0 {
		opts.Timeout = 3 * time.Second
	}

	// Reset services map
	b.services = make(map[string]ServiceInfo)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// Channel to receive service entries
	entries := make(chan *zeroconf.ServiceEntry)

	// Start browsing
	go func() {
		err := b.resolver.Browse(ctx, ServiceType, "local.", entries)
		if err != nil {
			fmt.Printf("Browse error: %v\n", err)
		}
	}()

	// Collect entries
	for {
		select {
		case entry := <-entries:
			if entry == nil {
				continue
			}

			// Parse service info
			info := b.parseEntry(entry)

			// Apply filters
			if opts.FilterChannel != "" && info.Channel != opts.FilterChannel {
				continue
			}
			if opts.FilterPriority != "" && info.Priority != opts.FilterPriority {
				continue
			}

			// Add to services map
			b.services[info.Name] = info

		case <-ctx.Done():
			// Timeout reached, return collected services
			return b.GetServices(), nil
		}
	}
}

// Subscribe browses continuously and calls callback for each discovered service
func (b *Browser) Subscribe(callback func(ServiceInfo), opts BrowseOptions) error {
	// Set default timeout (longer for continuous browsing)
	if opts.Timeout == 0 {
		opts.Timeout = 60 * time.Second
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// Channel to receive service entries
	entries := make(chan *zeroconf.ServiceEntry)

	// Start browsing
	go func() {
		err := b.resolver.Browse(ctx, ServiceType, "local.", entries)
		if err != nil {
			fmt.Printf("Browse error: %v\n", err)
		}
	}()

	// Process entries
	for {
		select {
		case entry := <-entries:
			if entry == nil {
				continue
			}

			// Parse service info
			info := b.parseEntry(entry)

			// Apply filters
			if opts.FilterChannel != "" && info.Channel != opts.FilterChannel {
				continue
			}
			if opts.FilterPriority != "" && info.Priority != opts.FilterPriority {
				continue
			}

			// Call callback
			callback(info)

		case <-ctx.Done():
			return nil
		}
	}
}

// GetServices returns all discovered services
func (b *Browser) GetServices() []ServiceInfo {
	services := make([]ServiceInfo, 0, len(b.services))
	for _, info := range b.services {
		services = append(services, info)
	}
	return services
}

// parseEntry parses a zeroconf ServiceEntry into ServiceInfo
func (b *Browser) parseEntry(entry *zeroconf.ServiceEntry) ServiceInfo {
	info := ServiceInfo{
		Name: entry.Instance,
		Host: entry.HostName,
		Port: entry.Port,
	}

	// Get IPv6 address (prefer IPv6 for Yggdrasil)
	for _, ip := range entry.AddrIPv6 {
		if ip != nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
			info.IPv6 = ip
			break
		}
	}

	// Fallback to any IPv6 if no global unicast found
	if info.IPv6 == nil && len(entry.AddrIPv6) > 0 {
		info.IPv6 = entry.AddrIPv6[0]
	}

	// Parse TXT records
	if len(entry.Text) > 0 {
		parsed, err := ParseTXTRecord(entry.Text)
		if err == nil {
			info.Group = parsed.Group
			info.Channel = parsed.Channel
			info.Callsign = parsed.Callsign
			info.Priority = parsed.Priority
			info.Codec = parsed.Codec
			info.Bitrate = parsed.Bitrate
		}
	}

	return info
}

// FormatServiceInfo returns a human-readable string for a ServiceInfo
func FormatServiceInfo(info ServiceInfo) string {
	ipv6Str := "<none>"
	if info.IPv6 != nil {
		ipv6Str = info.IPv6.String()
	}

	return fmt.Sprintf(
		"Name: %s | Callsign: %s | IPv6: [%s]:%d | Channel: %s | Priority: %s | Codec: %s @ %dkbps",
		info.Name,
		info.Callsign,
		ipv6Str,
		info.Port,
		info.Channel,
		info.Priority,
		info.Codec,
		info.Bitrate,
	)
}

// GetIPv6Addr returns the IPv6 address in [ip]:port format
func (info ServiceInfo) GetIPv6Addr() string {
	if info.IPv6 == nil {
		return ""
	}
	return net.JoinHostPort(info.IPv6.String(), fmt.Sprintf("%d", info.Port))
}
