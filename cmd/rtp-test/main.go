package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/meshradio/meshradio/pkg/audio"
	"github.com/meshradio/meshradio/pkg/rtp"
)

func main() {
	mode := flag.String("mode", "send", "Mode: send or receive")
	target := flag.String("target", "[::1]:9999", "Target address for sending")
	port := flag.Int("port", 9999, "Local port")
	flag.Parse()

	if *mode == "send" {
		runSender(*target, *port)
	} else {
		runReceiver(*port)
	}
}

func runSender(targetAddr string, localPort int) {
	fmt.Printf("RTP Sender Test\n")
	fmt.Printf("===============\n")
	fmt.Printf("Sending to: %s\n", targetAddr)
	fmt.Printf("Press Ctrl+C to stop\n\n")

	// Parse target address
	udpAddr, err := net.ResolveUDPAddr("udp6", targetAddr)
	if err != nil {
		log.Fatalf("Invalid target address: %v", err)
	}

	// Create RTP sender
	sender, err := rtp.NewSender(rtp.SenderConfig{
		LocalPort:   localPort,
		PayloadType: 111, // Opus
		SampleRate:  48000,
	})
	if err != nil {
		log.Fatalf("Failed to create RTP sender: %v", err)
	}
	defer sender.Close()

	// Create audio input (dummy for now - generates silence/tone)
	audioIn := audio.NewInputStream(audio.DefaultConfig())
	err = audioIn.Start()
	if err != nil {
		log.Fatalf("Failed to start audio input: %v", err)
	}
	defer audioIn.Stop()

	// Create Opus encoder
	codec := audio.NewDummyCodec(960) // 20ms at 48kHz

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Send loop
	ticker := time.NewTicker(20 * time.Millisecond) // 20ms frames
	defer ticker.Stop()

	packetCount := 0

	for {
		select {
		case <-ticker.C:
			// Read audio frame
			pcm, err := audioIn.Read()
			if err != nil {
				continue
			}

			// Encode to Opus
			opus, err := codec.Encode(pcm)
			if err != nil {
				log.Printf("Encode error: %v", err)
				continue
			}

			// Send as RTP
			err = sender.SendOpus(opus, udpAddr)
			if err != nil {
				log.Printf("Send error: %v", err)
				continue
			}

			packetCount++

			// Print stats every 50 packets (~1 second)
			if packetCount%50 == 0 {
				stats := sender.GetStats()
				fmt.Printf("Sent %d packets | SSRC: %d | Timestamp: %d\n",
					stats.PacketsSent, stats.SSRC, stats.CurrentTimestamp)
			}

		case <-sigChan:
			fmt.Println("\nStopping sender...")
			return
		}
	}
}

func runReceiver(localPort int) {
	fmt.Printf("RTP Receiver Test\n")
	fmt.Printf("=================\n")
	fmt.Printf("Listening on port: %d\n", localPort)
	fmt.Printf("Press Ctrl+C to stop\n\n")

	// Create RTP receiver
	receiver, err := rtp.NewReceiver(rtp.ReceiverConfig{
		LocalPort:  localPort,
		BufferSize: 50,
	})
	if err != nil {
		log.Fatalf("Failed to create RTP receiver: %v", err)
	}
	defer receiver.Stop()

	// Start receiving
	err = receiver.Start()
	if err != nil {
		log.Fatalf("Failed to start receiver: %v", err)
	}

	// Create audio output
	audioOut := audio.NewOutputStream(audio.DefaultConfig())
	err = audioOut.Start()
	if err != nil {
		log.Fatalf("Failed to start audio output: %v", err)
	}
	defer audioOut.Stop()

	// Create Opus decoder
	codec := audio.NewDummyCodec(960)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Receive loop
	packetCount := 0

	for {
		select {
		case <-sigChan:
			fmt.Println("\nStopping receiver...")
			return

		default:
			// Read Opus data from RTP
			opus, err := receiver.ReadOpus()
			if err != nil {
				log.Printf("Receive error: %v", err)
				return
			}

			// Decode Opus to PCM
			pcm, err := codec.Decode(opus)
			if err != nil {
				log.Printf("Decode error: %v", err)
				continue
			}

			// Play audio
			audioOut.Write(pcm)

			packetCount++

			// Print stats every 50 packets (~1 second)
			if packetCount%50 == 0 {
				stats := receiver.GetStats()
				fmt.Printf("Received %d packets | Lost: %d (%.2f%%)\n",
					stats.PacketsReceived, stats.PacketsLost, stats.PacketLossRate)
			}
		}
	}
}
