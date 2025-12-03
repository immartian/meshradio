package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/meshradio/meshradio/pkg/mdns"
)

func main() {
	mode := flag.String("mode", "browse", "Mode: advertise, browse, or query")
	callsign := flag.String("callsign", "TEST1", "Station callsign")
	port := flag.Int("port", 8790, "RTP port")
	channel := flag.String("channel", "community", "Channel type (emergency, community, talk)")
	priority := flag.String("priority", "normal", "Priority level (normal, high, emergency, critical)")
	filter := flag.String("filter", "", "Filter by channel when browsing")
	timeout := flag.Int("timeout", 3, "Browse timeout in seconds")
	flag.Parse()

	switch *mode {
	case "advertise":
		runAdvertise(*callsign, *port, *channel, *priority)
	case "browse":
		runBrowse(*filter, *timeout)
	case "query":
		runQuery(*callsign, *timeout)
	default:
		log.Fatalf("Invalid mode: %s (use: advertise, browse, or query)", *mode)
	}
}

func runAdvertise(callsign string, port int, channel string, priority string) {
	fmt.Printf("mDNS Advertiser Test\n")
	fmt.Printf("====================\n")
	fmt.Printf("Callsign: %s\n", callsign)
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Channel: %s\n", channel)
	fmt.Printf("Priority: %s\n", priority)
	fmt.Printf("Press Ctrl+C to stop\n\n")

	// Create service info
	info := mdns.ServiceInfo{
		Name:     callsign,
		Port:     port,
		Group:    channel,
		Channel:  channel,
		Callsign: callsign,
		Priority: priority,
		Codec:    "opus",
		Bitrate:  64,
	}

	// Create advertiser
	advertiser, err := mdns.NewAdvertiser(info)
	if err != nil {
		log.Fatalf("Failed to create advertiser: %v", err)
	}
	defer advertiser.Shutdown()

	fmt.Printf("✓ Advertising service: %s\n", mdns.FormatServiceInfo(info))
	fmt.Printf("✓ Service type: %s.local.\n", mdns.ServiceType)
	fmt.Printf("\nService is now discoverable via mDNS/Avahi\n")
	fmt.Printf("Test with: avahi-browse -r %s\n\n", mdns.ServiceType)

	// Wait for signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nStopping advertiser...")
}

func runBrowse(filterChannel string, timeoutSec int) {
	fmt.Printf("mDNS Browser Test\n")
	fmt.Printf("=================\n")
	if filterChannel != "" {
		fmt.Printf("Filter: channel=%s\n", filterChannel)
	}
	fmt.Printf("Timeout: %d seconds\n", timeoutSec)
	fmt.Printf("Browsing for MeshRadio services...\n\n")

	// Create browser
	browser, err := mdns.NewBrowser()
	if err != nil {
		log.Fatalf("Failed to create browser: %v", err)
	}

	// Browse for services
	opts := mdns.BrowseOptions{
		Timeout:       time.Duration(timeoutSec) * time.Second,
		FilterChannel: filterChannel,
	}

	services, err := browser.Browse(opts)
	if err != nil {
		log.Fatalf("Browse failed: %v", err)
	}

	// Display results
	if len(services) == 0 {
		fmt.Println("No services found.")
		fmt.Println("\nTips:")
		fmt.Println("- Make sure a broadcaster is running (./mdns-test -mode advertise)")
		fmt.Println("- Check firewall settings (port 5353 UDP)")
		fmt.Println("- Ensure you're on the same network segment")
		return
	}

	fmt.Printf("Found %d service(s):\n\n", len(services))
	for i, info := range services {
		fmt.Printf("%d. %s\n", i+1, mdns.FormatServiceInfo(info))
		if info.IPv6 != nil {
			fmt.Printf("   → Connect with: ./rtp-test -mode receive -port %d\n", info.Port)
			fmt.Printf("   → (Broadcaster should use: -target %s)\n", info.GetIPv6Addr())
		}
		fmt.Println()
	}
}

func runQuery(callsign string, timeoutSec int) {
	fmt.Printf("mDNS Query Test\n")
	fmt.Printf("===============\n")
	fmt.Printf("Searching for: %s\n", callsign)
	fmt.Printf("Timeout: %d seconds\n\n", timeoutSec)

	// Create browser
	browser, err := mdns.NewBrowser()
	if err != nil {
		log.Fatalf("Failed to create browser: %v", err)
	}

	// Browse for services
	opts := mdns.BrowseOptions{
		Timeout: time.Duration(timeoutSec) * time.Second,
	}

	services, err := browser.Browse(opts)
	if err != nil {
		log.Fatalf("Browse failed: %v", err)
	}

	// Find matching callsign
	for _, info := range services {
		if info.Callsign == callsign || info.Name == callsign {
			fmt.Printf("✓ Found: %s\n\n", mdns.FormatServiceInfo(info))
			if info.IPv6 != nil {
				fmt.Printf("Connect with:\n")
				fmt.Printf("  ./rtp-test -mode receive -port %d\n", info.Port)
				fmt.Printf("  (Broadcaster: -target %s)\n", info.GetIPv6Addr())
			}
			return
		}
	}

	fmt.Printf("✗ Service not found: %s\n", callsign)
	fmt.Printf("\nAvailable services:\n")
	for _, info := range services {
		fmt.Printf("  - %s (%s)\n", info.Callsign, info.Name)
	}
}
