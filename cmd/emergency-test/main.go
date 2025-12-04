package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/meshradio/meshradio/internal/broadcaster"
	"github.com/meshradio/meshradio/internal/listener"
	"github.com/meshradio/meshradio/pkg/audio"
	"github.com/meshradio/meshradio/pkg/emergency"
	"github.com/meshradio/meshradio/pkg/yggdrasil"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	mode := os.Args[1]

	switch mode {
	case "broadcast-critical":
		broadcastCritical()
	case "broadcast-emergency":
		broadcastEmergency()
	case "broadcast-high":
		broadcastHigh()
	case "broadcast-normal":
		broadcastNormal()
	case "listen-autotune":
		listenAutoTune()
	case "listen-manual":
		listenManual()
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Emergency Priority Test Program")
	fmt.Println()
	fmt.Println("Usage: emergency-test <mode>")
	fmt.Println()
	fmt.Println("Broadcast Modes:")
	fmt.Println("  broadcast-critical   - Broadcast on emergency channel (critical priority)")
	fmt.Println("  broadcast-emergency  - Broadcast on netcontrol channel (emergency priority)")
	fmt.Println("  broadcast-high       - Broadcast on weather channel (high priority)")
	fmt.Println("  broadcast-normal     - Broadcast on community channel (normal priority)")
	fmt.Println()
	fmt.Println("Listen Modes:")
	fmt.Println("  listen-autotune      - Listen with auto-tune enabled")
	fmt.Println("  listen-manual        - Listen without auto-tune")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  Terminal 1: emergency-test broadcast-critical")
	fmt.Println("  Terminal 2: emergency-test listen-manual")
}

func broadcastCritical() {
	broadcast("emergency", "EMERGENCY-TEST", 8790)
}

func broadcastEmergency() {
	broadcast("netcontrol", "NETCONTROL-TEST", 8791)
}

func broadcastHigh() {
	broadcast("weather", "WEATHER-TEST", 8793)
}

func broadcastNormal() {
	broadcast("community", "COMMUNITY-TEST", 8795)
}

func broadcast(group, callsign string, port int) {
	// Get local IPv6
	ipv6, err := yggdrasil.GetLocalIPv6()
	if err != nil {
		fmt.Printf("Error getting IPv6: %v\n", err)
		fmt.Println("Make sure Yggdrasil is running!")
		os.Exit(1)
	}

	// Get channel info
	registry := emergency.NewChannelRegistry()
	channel, ok := registry.GetByGroup(group)
	if !ok {
		fmt.Printf("Unknown group: %s\n", group)
		os.Exit(1)
	}

	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║ Emergency Broadcast Test                                     ║\n")
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ Channel:     %-47s ║\n", channel.Name)
	fmt.Printf("║ Group:       %-47s ║\n", group)
	fmt.Printf("║ Priority:    %-47s ║\n", channel.Priority.String())
	fmt.Printf("║ Port:        %-47d ║\n", port)
	fmt.Printf("║ Callsign:    %-47s ║\n", callsign)
	fmt.Printf("║ IPv6:        %-47s ║\n", ipv6.String())
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n")
	fmt.Println()

	// Create broadcaster config
	cfg := broadcaster.Config{
		Callsign: callsign,
		IPv6:     ipv6,
		Port:     port,
		Group:    group,
		AudioConfig: audio.StreamConfig{
			SampleRate: 48000,
			Channels:   1,
			FrameSize:  960,
			Bitrate:    24000,
		},
	}

	// Create and start broadcaster
	b, err := broadcaster.New(cfg)
	if err != nil {
		fmt.Printf("Error creating broadcaster: %v\n", err)
		os.Exit(1)
	}

	if err := b.Start(); err != nil {
		fmt.Printf("Error starting broadcaster: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Broadcasting started. Press Ctrl+C to stop.")
	fmt.Printf("Priority level: %s (%d)\n", channel.Priority.String(), channel.Priority)
	fmt.Println()

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nStopping broadcaster...")
	b.Stop()
}

func listenAutoTune() {
	listen(true)
}

func listenManual() {
	listen(false)
}

func listen(autoTune bool) {
	// Command line flags
	var targetAddr string
	var targetPort int
	var group string
	var callsign string

	fs := flag.NewFlagSet("listen", flag.ExitOnError)
	fs.StringVar(&targetAddr, "target", "", "Target broadcaster IPv6 address")
	fs.IntVar(&targetPort, "port", 8790, "Target broadcaster port")
	fs.StringVar(&group, "group", "emergency", "Multicast group to join")
	fs.StringVar(&callsign, "callsign", "LISTENER-TEST", "Your callsign")
	fs.Parse(os.Args[2:])

	if targetAddr == "" {
		fmt.Println("Error: -target flag is required")
		fmt.Println()
		fmt.Println("Usage: emergency-test listen-manual -target <ipv6> [-port <port>] [-group <group>]")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  emergency-test listen-manual -target 200:1234::5678 -port 8790 -group emergency")
		os.Exit(1)
	}

	// Get local IPv6
	localIPv6, err := yggdrasil.GetLocalIPv6()
	if err != nil {
		fmt.Printf("Error getting IPv6: %v\n", err)
		fmt.Println("Make sure Yggdrasil is running!")
		os.Exit(1)
	}

	// Parse target IPv6
	targetIPv6 := net.ParseIP(targetAddr)
	if targetIPv6 == nil {
		fmt.Printf("Error parsing target address: %s\n", targetAddr)
		os.Exit(1)
	}

	// Get channel info
	registry := emergency.NewChannelRegistry()
	channel, _ := registry.GetByGroup(group)

	autoTuneStr := "Disabled"
	if autoTune {
		autoTuneStr = "Enabled (not yet implemented)"
	}

	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║ Emergency Listener Test                                      ║\n")
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ Channel:     %-47s ║\n", channel.Name)
	fmt.Printf("║ Group:       %-47s ║\n", group)
	fmt.Printf("║ Priority:    %-47s ║\n", channel.Priority.String())
	fmt.Printf("║ Port:        %-47d ║\n", targetPort)
	fmt.Printf("║ Callsign:    %-47s ║\n", callsign)
	fmt.Printf("║ Target:      %-47s ║\n", targetAddr)
	fmt.Printf("║ Auto-tune:   %-47s ║\n", autoTuneStr)
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n")
	fmt.Println()

	// Create listener config
	cfg := listener.Config{
		Callsign:   callsign,
		LocalIPv6:  localIPv6,
		LocalPort:  targetPort + 1000, // Use different port for listener
		TargetIPv6: targetIPv6,
		TargetPort: targetPort,
		Group:      group,
		SSMSource:  targetIPv6, // SSM mode - only from this broadcaster
		AudioConfig: audio.StreamConfig{
			SampleRate: 48000,
			Channels:   1,
			FrameSize:  960,
			Bitrate:    24000,
		},
	}

	// Create and start listener
	l, err := listener.New(cfg)
	if err != nil {
		fmt.Printf("Error creating listener: %v\n", err)
		os.Exit(1)
	}

	if err := l.Start(); err != nil {
		fmt.Printf("Error starting listener: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Listening started. Press Ctrl+C to stop.")
	fmt.Println()
	fmt.Println("Watch for priority change alerts:")
	fmt.Println("  🚨 = Critical Emergency")
	fmt.Println("  ⚠️  = Emergency")
	fmt.Println("  📢 = High Priority")
	fmt.Println()

	// Show stats periodically
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			packets, seq, station := l.GetStats()
			if packets > 0 {
				fmt.Printf("Stats: received=%d packets, seq=%d, station=%s\n",
					packets, seq, station)
			}
		case <-sigChan:
			fmt.Println("\nStopping listener...")
			l.Stop()
			return
		}
	}
}
