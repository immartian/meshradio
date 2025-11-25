package yggdrasil

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

// SelfInfo contains information about the local Yggdrasil node
type SelfInfo struct {
	IPv6      net.IP
	PublicKey string
	Coords    string
}

// Client interacts with Yggdrasil daemon
type Client struct {
	adminSocket string
}

// NewClient creates a new Yggdrasil client
func NewClient() *Client {
	return &Client{
		adminSocket: "/var/run/yggdrasil/yggdrasil.sock",
	}
}

// GetSelf retrieves information about the local Yggdrasil node
func (c *Client) GetSelf() (*SelfInfo, error) {
	// Try yggdrasilctl command first
	cmd := exec.Command("yggdrasilctl", "getSelf")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run yggdrasilctl: %w (is Yggdrasil installed?)", err)
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse yggdrasilctl output: %w", err)
	}

	// Extract IPv6 address
	ipv6Str, ok := result["address"].(string)
	if !ok {
		return nil, fmt.Errorf("no address field in yggdrasilctl output")
	}

	// Parse IPv6
	ipv6 := net.ParseIP(ipv6Str)
	if ipv6 == nil {
		return nil, fmt.Errorf("invalid IPv6 address: %s", ipv6Str)
	}

	info := &SelfInfo{
		IPv6: ipv6,
	}

	// Optional fields
	if pubkey, ok := result["key"].(string); ok {
		info.PublicKey = pubkey
	}
	if coords, ok := result["coords"].(string); ok {
		info.Coords = coords
	}

	return info, nil
}

// IsAvailable checks if Yggdrasil is installed and running
func (c *Client) IsAvailable() bool {
	_, err := c.GetSelf()
	return err == nil
}

// GetPeers retrieves list of connected peers
func (c *Client) GetPeers() ([]string, error) {
	cmd := exec.Command("yggdrasilctl", "getPeers")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get peers: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse peers output: %w", err)
	}

	peers := make([]string, 0)
	if peersMap, ok := result["peers"].(map[string]interface{}); ok {
		for peer := range peersMap {
			peers = append(peers, peer)
		}
	}

	return peers, nil
}

// GetNodeInfo retrieves node information for a given IPv6
func (c *Client) GetNodeInfo(ipv6 net.IP) (map[string]interface{}, error) {
	cmd := exec.Command("yggdrasilctl", "getNodeInfo", ipv6.String())
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get node info: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse node info: %w", err)
	}

	return result, nil
}

// DetectYggdrasilIPv6 attempts to detect Yggdrasil IPv6 from network interfaces
// Fallback method if yggdrasilctl is not available
func DetectYggdrasilIPv6() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// Look for Yggdrasil interface (usually tun0, tun1, etc.)
	for _, iface := range ifaces {
		// Check common Yggdrasil interface names
		if !strings.HasPrefix(iface.Name, "tun") && !strings.HasPrefix(iface.Name, "ygg") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipnet.IP
			if ip.To4() != nil {
				continue // Skip IPv4
			}

			// Yggdrasil addresses start with 0x02 or 0x03
			if ip[0] == 0x02 || ip[0] == 0x03 {
				return ip, nil
			}
		}
	}

	return nil, fmt.Errorf("no Yggdrasil IPv6 address found")
}

// GetLocalIPv6 tries multiple methods to get the Yggdrasil IPv6
func GetLocalIPv6() (net.IP, error) {
	// Method 1: Try yggdrasilctl
	client := NewClient()
	if info, err := client.GetSelf(); err == nil {
		return info.IPv6, nil
	}

	// Method 2: Try interface detection
	if ip, err := DetectYggdrasilIPv6(); err == nil {
		return ip, nil
	}

	// Method 3: Check environment variable
	// if ipStr := os.Getenv("YGGDRASIL_IPV6"); ipStr != "" {
	// 	if ip := net.ParseIP(ipStr); ip != nil {
	// 		return ip, nil
	// 	}
	// }

	return nil, fmt.Errorf("could not detect Yggdrasil IPv6 address")
}
