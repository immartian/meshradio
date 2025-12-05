package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/meshradio/meshradio/pkg/gui"
	"github.com/meshradio/meshradio/pkg/ui"
	"github.com/meshradio/meshradio/pkg/yggdrasil"
)

func main() {
	// Parse command line flags
	guiMode := flag.Bool("gui", false, "Launch web GUI instead of TUI")
	port := flag.Int("port", 8799, "Port for audio broadcast/listen (default: 8799)")
	webPort := flag.Int("web-port", 8899, "Port for web GUI (default: 8899, only used with --gui)")
	callsign := flag.String("callsign", "", "Station callsign (or use MESHRADIO_CALLSIGN env var)")
	flag.Parse()

	// Get local Yggdrasil IPv6
	localIPv6 := getLocalIPv6()

	// Get callsign (priority: flag > env > args > default)
	stationCallsign := getCallsign(*callsign)

	// Launch appropriate interface
	if *guiMode {
		runGUI(stationCallsign, localIPv6, *port, *webPort)
	} else {
		runTUI(stationCallsign, localIPv6, *port)
	}
}

func runTUI(callsign string, ipv6 net.IP, port int) {
	// Create and run TUI
	model := ui.NewModel(callsign, ipv6, port)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runGUI(callsign string, ipv6 net.IP, audioPort int, webPort int) {
	// Print startup info
	fmt.Printf("🚀 Starting MeshRadio Web GUI\n\n")
	fmt.Printf("Callsign: %s\n", callsign)
	fmt.Printf("IPv6: %s\n", ipv6.String())
	fmt.Printf("Audio Port: %d\n\n", audioPort)
	fmt.Printf("🌐 Web GUI: http://localhost:%d\n", webPort)
	fmt.Printf("📱 Open in your browser to control MeshRadio\n\n")

	// Create and start GUI server
	server := gui.NewServer(webPort, callsign, ipv6)
	server.SetAudioPort(audioPort)

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
		fmt.Fprintf(os.Stderr, "Using fallback address. Install Yggdrasil for full functionality.\n")
		// Fallback to localhost for testing
		ipv6 = net.IPv6loopback
	}
	return ipv6
}

// getCallsign gets or generates a callsign
// Priority: flag parameter > env variable > remaining args > default
func getCallsign(flagCallsign string) string {
	// Priority 1: explicit flag
	if flagCallsign != "" {
		return flagCallsign
	}

	// Priority 2: environment variable
	if envCallsign := os.Getenv("MESHRADIO_CALLSIGN"); envCallsign != "" {
		return envCallsign
	}

	// Priority 3: remaining args after flags
	if flag.NArg() > 0 {
		return flag.Arg(0)
	}

	// Default
	return "STATION"
}
