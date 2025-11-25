package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/meshradio/meshradio/pkg/gui"
	"github.com/meshradio/meshradio/pkg/yggdrasil"
)

func main() {
	// Get local Yggdrasil IPv6
	localIPv6 := getLocalIPv6()

	// Get or prompt for callsign
	callsign := getCallsign()

	// Print startup info
	fmt.Printf("🚀 Starting MeshRadio Web GUI\n\n")
	fmt.Printf("Callsign: %s\n", callsign)
	fmt.Printf("IPv6: %s\n\n", localIPv6.String())

	// Create and start GUI server
	server := gui.NewServer(8080, callsign, localIPv6)

	fmt.Printf("🌐 Web GUI: http://localhost:8080\n")
	fmt.Printf("📱 Open in your browser to control MeshRadio\n\n")

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

// getLocalIPv6 gets the local Yggdrasil IPv6 address
func getLocalIPv6() net.IP {
	// Try to get real Yggdrasil IPv6
	ipv6, err := yggdrasil.GetLocalIPv6()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not detect Yggdrasil IPv6: %v\n", err)
		fmt.Fprintf(os.Stderr, "Using fallback address. Install Yggdrasil for full functionality.\n\n")
		// Fallback to localhost for testing
		ipv6 = net.IPv6loopback
	}
	return ipv6
}

// getCallsign gets or generates a callsign
func getCallsign() string {
	// Check environment variable
	if callsign := os.Getenv("MESHRADIO_CALLSIGN"); callsign != "" {
		return callsign
	}

	// Check command line args
	if len(os.Args) > 1 {
		return os.Args[1]
	}

	// Default
	return "STATION"
}
