package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/meshradio/meshradio/internal/broadcaster"
	"github.com/meshradio/meshradio/pkg/audio"
	"github.com/meshradio/meshradio/pkg/yggdrasil"
)

var (
	musicDir  = flag.String("dir", "", "Music directory to scan (default: ~/Music)")
	callsign  = flag.String("callsign", "MUSIC-DJ", "Your callsign")
	port      = flag.Int("port", 8799, "Broadcast port (default: 8799 standard broadcaster port)")
	group     = flag.String("group", "default", "Multicast group")
	shuffle   = flag.Bool("shuffle", false, "Shuffle playlist")
	loop      = flag.Bool("loop", true, "Loop playlist")
	advertise = flag.Bool("advertise", true, "Advertise via mDNS")
)

type Playlist struct {
	files   []string
	current int
}

func main() {
	flag.Parse()

	// Determine music directory
	dir := *musicDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		dir = filepath.Join(home, "Music")
	}

	// Check directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Error: Music directory not found: %s\n", dir)
		fmt.Println("Use --dir to specify a different directory")
		os.Exit(1)
	}

	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║ MeshRadio Music Broadcaster                                  ║\n")
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ Callsign:    %-47s ║\n", *callsign)
	fmt.Printf("║ Channel:     %-47s ║\n", *group)
	fmt.Printf("║ Port:        %-47d ║\n", *port)
	fmt.Printf("║ Music Dir:   %-47s ║\n", dir)
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n")
	fmt.Println()

	// Scan for MP3 files
	fmt.Println("🔍 Scanning for MP3 files...")
	playlist, err := scanMusicDir(dir)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(playlist.files) == 0 {
		fmt.Println("No MP3 files found!")
		fmt.Printf("Checked directory: %s\n", dir)
		os.Exit(1)
	}

	fmt.Printf("✅ Found %d MP3 file(s)\n", len(playlist.files))
	fmt.Println()

	// Show playlist preview
	fmt.Println("📻 Playlist:")
	for i, file := range playlist.files {
		name := filepath.Base(file)
		if i < 10 {
			fmt.Printf("  %d. %s\n", i+1, name)
		}
	}
	if len(playlist.files) > 10 {
		fmt.Printf("  ... and %d more\n", len(playlist.files)-10)
	}
	fmt.Println()

	// Get local IPv6
	ipv6, err := yggdrasil.GetLocalIPv6()
	if err != nil {
		fmt.Printf("Warning: Could not get Yggdrasil IPv6: %v\n", err)
		fmt.Println("Using localhost for testing")
		ipv6 = net.IPv6loopback
	}

	fmt.Printf("📡 Broadcasting on: %s:%d\n", ipv6.String(), *port)
	fmt.Println()

	fmt.Println("🎵 Starting music broadcast...")
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  Ctrl+C  - Stop broadcasting")
	fmt.Println()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Play through playlist
	for {
		for i, file := range playlist.files {
			// Check for interrupt
			select {
			case <-sigChan:
				fmt.Println("\n\nStopping music broadcast...")
				return
			default:
			}

			fmt.Printf("▶️  Now playing [%d/%d]: %s\n", i+1, len(playlist.files), filepath.Base(file))

			// Get file info for duration display
			duration, sampleRate := getMP3Info(file)
			fmt.Printf("   Duration: %s | Sample Rate: %d Hz\n", duration.Round(time.Second), sampleRate)

			// Broadcast this file
			if err := broadcastFile(file, ipv6, *port, *group, *callsign, sigChan); err != nil {
				if err == io.EOF {
					// File finished normally
					fmt.Printf("   ✅ Completed\n\n")
				} else {
					fmt.Printf("   ❌ Error: %v\n\n", err)
					// If interrupted, exit main loop
					if err.Error() == "interrupted" {
						fmt.Println("\n✅ Stopped by user")
						return
					}
				}
			}
		}

		// Check if we should loop
		if !*loop {
			break
		}

		fmt.Println("🔄 Looping playlist...\n")
	}

	fmt.Println("✅ Playlist complete!")
}

func broadcastFile(filepath string, ipv6 net.IP, port int, group, callsign string, sigChan chan os.Signal) error {
	// Create audio config for music - use high quality settings
	audioConfig := audio.DefaultConfig()

	// Create MP3 source
	mp3Source, err := audio.NewMP3Source(filepath, audioConfig)
	if err != nil {
		return fmt.Errorf("failed to create MP3 source: %w", err)
	}

	// Create broadcaster with MP3 source
	cfg := broadcaster.Config{
		Callsign:    callsign,
		IPv6:        ipv6,
		Port:        port,
		Group:       group,
		AudioConfig: audioConfig,
		AudioSource: mp3Source, // Use MP3 as audio source!
	}

	b, err := broadcaster.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create broadcaster: %w", err)
	}

	// Start broadcasting
	if err := b.Start(); err != nil {
		return fmt.Errorf("failed to start broadcaster: %w", err)
	}

	// Wait for file to finish or interrupt
	// The broadcaster will automatically stop when MP3 source returns EOF
	// We'll check periodically for signals
	for {
		select {
		case <-sigChan:
			b.Stop()
			return fmt.Errorf("interrupted")
		case <-time.After(1 * time.Second):
			// Check if source is still running
			if !mp3Source.IsRunning() {
				b.Stop()
				return io.EOF
			}
		}
	}
}

func scanMusicDir(dir string) (*Playlist, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".mp3" {
				files = append(files, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Playlist{
		files:   files,
		current: 0,
	}, nil
}

func getMP3Info(filepath string) (time.Duration, int) {
	f, err := os.Open(filepath)
	if err != nil {
		return 0, 0
	}
	defer f.Close()

	decoder, err := mp3.NewDecoder(f)
	if err != nil {
		return 0, 0
	}

	sampleRate := decoder.SampleRate()
	length := decoder.Length()
	duration := time.Duration(length) * time.Second / time.Duration(sampleRate) / 4 // 4 = 2 channels * 2 bytes per sample

	return duration, sampleRate
}
