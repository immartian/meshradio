package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/meshradio/meshradio/internal/broadcaster"
	"github.com/meshradio/meshradio/pkg/audio"
	"github.com/meshradio/meshradio/pkg/yggdrasil"
)

var (
	musicDir  = flag.String("dir", "", "Music directory to scan (default: ~/Music)")
	callsign  = flag.String("callsign", "MUSIC-DJ", "Your callsign")
	port      = flag.Int("port", 8798, "Broadcast port (default: 8798 for 'talk' channel)")
	group     = flag.String("group", "talk", "Multicast group")
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

	// Show playlist
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
		ipv6 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	}

	fmt.Printf("📡 Broadcasting on: %s:%d\n", formatIPv6(ipv6), *port)
	fmt.Println()

	// Create broadcaster
	cfg := broadcaster.Config{
		Callsign: *callsign,
		IPv6:     ipv6,
		Port:     *port,
		Group:    *group,
		AudioConfig: audio.StreamConfig{
			SampleRate: 48000,
			Channels:   2, // Stereo for music
			FrameSize:  960,
			Bitrate:    128000, // Higher bitrate for music quality
		},
	}

	b, err := broadcaster.New(cfg)
	if err != nil {
		fmt.Printf("Error creating broadcaster: %v\n", err)
		os.Exit(1)
	}

	// Note: We won't actually start the broadcaster yet
	// because we need to replace its audio input with MP3 decoder
	// For now, this is a placeholder that shows the concept

	fmt.Println("🎵 Starting music broadcast...")
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  Ctrl+C  - Stop broadcasting")
	fmt.Println()

	// TODO: Implement actual MP3 playback integration
	// This requires refactoring broadcaster to accept custom audio source
	// For now, just show what would be played

	fmt.Println("⚠️  NOTE: MP3 playback not yet integrated")
	fmt.Println("    This is a demonstration of the music discovery feature.")
	fmt.Println("    Full MP3 playback integration requires:")
	fmt.Println("    1. MP3 decoder (go-mp3)")
	fmt.Println("    2. Sample rate conversion (44.1kHz → 48kHz)")
	fmt.Println("    3. Audio pipeline refactoring")
	fmt.Println()
	fmt.Println("    For now, use the main 'meshradio' program to broadcast")
	fmt.Println("    live audio from your microphone while playing music locally.")
	fmt.Println()

	// Simulate playback
	for i, file := range playlist.files {
		fmt.Printf("▶️  Now playing [%d/%d]: %s\n", i+1, len(playlist.files), filepath.Base(file))

		// Open MP3 file to get duration
		f, err := os.Open(file)
		if err != nil {
			fmt.Printf("   Error opening file: %v\n", err)
			continue
		}

		decoder, err := mp3.NewDecoder(f)
		if err != nil {
			fmt.Printf("   Error decoding MP3: %v\n", err)
			f.Close()
			continue
		}

		// Calculate approximate duration
		sampleRate := decoder.SampleRate()
		length := decoder.Length()
		duration := time.Duration(length) * time.Second / time.Duration(sampleRate) / 4 // 4 = 2 channels * 2 bytes per sample

		fmt.Printf("   Duration: %s | Sample Rate: %d Hz\n", duration.Round(time.Second), sampleRate)

		f.Close()

		// Sleep to simulate playback
		time.Sleep(3 * time.Second)

		// Check for interrupt
		select {
		case <-make(chan struct{}):
		default:
		}
	}

	_ = b // Silence unused variable warning

	fmt.Println()
	fmt.Println("✅ Playlist complete!")
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

func formatIPv6(ipv6 []byte) string {
	if len(ipv6) != 16 {
		return "invalid"
	}
	return fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
		ipv6[0], ipv6[1], ipv6[2], ipv6[3], ipv6[4], ipv6[5], ipv6[6], ipv6[7],
		ipv6[8], ipv6[9], ipv6[10], ipv6[11], ipv6[12], ipv6[13], ipv6[14], ipv6[15])
}

// playMP3 demonstrates how MP3 playback would work
// This is a simplified example - full integration requires more work
func playMP3(filepath string, output chan []int16) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}

	// Read samples
	buf := make([]byte, 4096)
	for {
		n, err := decoder.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Convert bytes to int16 samples
		samples := make([]int16, n/2)
		for i := 0; i < n/2; i++ {
			samples[i] = int16(buf[i*2]) | int16(buf[i*2+1])<<8
		}

		// Send to output channel
		// (In real implementation, this would go through resampling and encoding)
		_ = samples

		// Check for interrupt
		select {
		case <-make(chan struct{}):
			return nil
		default:
		}
	}

	return nil
}
