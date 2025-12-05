# MeshRadio

Decentralized radio broadcasting over Yggdrasil mesh network.

## What is MeshRadio?

MeshRadio brings HAM radio-style broadcasting to the Yggdrasil mesh network. Use IPv6 addresses as "frequencies" to broadcast and listen to audio streams across the mesh.

**Architecture:** Subscription-based unicast streaming. Listeners "tune" to a broadcaster's IPv6:port, establishing a direct connection. The system uses Opus compression for high-quality audio at efficient bitrates.

## Current Status: v0.3-alpha

**STABLE MUSIC STREAMING ACHIEVED**

Recent development has delivered stable, continuous music broadcasting with the following improvements:
- Smooth, jitter-free audio playback
- Continuous streaming without interruption
- Playlist cycling support
- High-quality Opus compression (12-13x, 128kbps)

The system has been tested with multi-hour broadcasts of 93-song playlists with excellent stability.

## Features

### Working Now
- **Music Broadcasting** - Stream MP3 files from your music library
- **Playlist Management** - Automatic cycling through your music collection
- **High-Quality Audio** - Opus codec at 128kbps for music quality
- **Subscription System** - Listeners subscribe to broadcasters with automatic heartbeats
- **Emergency Priority** - Priority levels for critical broadcasts
- **Web GUI** - Browser-based station management interface
- **Cross-platform** - Linux, macOS, Windows support

### In Development
- Voice broadcasting (microphone input)
- Multi-listener stress testing
- Station discovery and browsing
- Enhanced metadata (now playing, artist info)

## Prerequisites

1. **Go 1.21 or later**
   ```bash
   go version  # Should be 1.21+
   ```

2. **FFmpeg** (for MP3 decoding)
   ```bash
   # macOS
   brew install ffmpeg

   # Ubuntu/Debian
   sudo apt-get install ffmpeg

   # Fedora
   sudo dnf install ffmpeg

   # Arch
   sudo pacman -S ffmpeg
   ```

3. **Yggdrasil** (optional - works on localhost for testing)
   - Install from: https://yggdrasil-network.github.io/
   - For production use over mesh networks

## Installation

### Build from Source

```bash
git clone https://github.com/immartian/meshradio
cd meshradio
go build -o meshradio ./cmd/meshradio
```

## Quick Start

### Option 1: Interactive Menu (Recommended)

```bash
./meshradio
```

Select from the menu:
1. Music Broadcaster - Scan and broadcast MP3 files
2. Voice Broadcaster - Broadcast from microphone (experimental)
3. Listener - Listen to broadcasts
4. Emergency Test - Test emergency priorities
5. Discovery Test - Test mDNS service discovery
6. Integration Test - Full two-node test

### Option 2: Web GUI

```bash
./meshradio --gui
```

Then open http://localhost:8799 in your browser.

### Testing Locally (No Yggdrasil Required)

For testing, the system works on localhost (::1):

**Terminal 1 - Broadcaster:**
```bash
./meshradio
# Select "1" for Music Broadcaster
# Choose your music directory
```

**Terminal 2 - Listener:**
```bash
./meshradio
# Select "3" for Listener
# Enter IPv6: ::1
# Enter port: 8799
```

### Broadcasting Over Yggdrasil

1. Check your Yggdrasil IPv6:
   ```bash
   yggdrasilctl getSelf
   ```

2. Start broadcaster with your Yggdrasil address

3. Share your IPv6:port with listeners (e.g., `200:1234:5678::1:8799`)

## Technical Details

### Audio Pipeline

**Broadcaster:**
```
MP3 File → FFmpeg Decoder → PCM Audio → Opus Encoder → UDP Packets → Network
```

**Listener:**
```
Network → UDP Packets → Opus Decoder → PCM Audio → Audio Device (Speakers)
```

### Performance Characteristics

- **Compression**: 12-13x (3840 bytes → ~300 bytes per frame)
- **Frame Size**: 20ms @ 48kHz stereo
- **Bitrate**: 128kbps for music quality
- **Buffer**: 3 seconds (150 frames) for jitter tolerance
- **Transport**: UDP over IPv6
- **Codec**: Opus (RFC 6716)

### Network Architecture

- **Protocol**: Custom protocol over UDP
- **Addressing**: IPv6 (Yggdrasil or localhost)
- **Delivery**: Unicast fan-out (broadcaster sends to each subscriber)
- **Reliability**: Heartbeat mechanism (5s interval)
- **Subscription**: Explicit subscribe/unsubscribe packets

## Recent Fixes (v0.3-alpha)

### Audio Stability Improvements
- Fixed audio callback race condition causing spurious underruns
- Implemented ticker-based packet pacing (exactly 20ms intervals)
- Changed audio callback to use 5ms timeout instead of immediate fallback
- Increased buffer from 50 to 150 frames for better jitter tolerance

### Playlist Management
- Removed infinite loop to allow proper playlist cycling
- Songs now advance automatically when finished
- Playlist loops back to start after last track

### Performance Tuning
- Reduced logging verbosity (5-second intervals)
- Added diagnostic logging for network troubleshooting
- Optimized FFmpeg realtime decoding with `-re` flag

## Architecture

See [DESIGN.md](DESIGN.md) for full technical specification.

## Command Line Options

```bash
# Interactive TUI mode (default)
./meshradio

# Web GUI mode
./meshradio --gui

# Show version
./meshradio --version

# Get help
./meshradio --help
```

## Troubleshooting

### Audio playback stops after 1-2 minutes
**Status**: Fixed in v0.3-alpha. Update to latest version.

### No audio output
- Check audio device permissions
- Verify FFmpeg is installed: `ffmpeg -version`
- Check audio buffer isn't full (should see "buffer=X/150" in logs)

### Can't connect to broadcaster
- Verify IPv6 address is correct
- Check firewall allows UDP traffic
- Test on localhost first (::1)
- Ensure Yggdrasil is running (for mesh networking)

### High memory usage
- Normal buffer usage: ~50MB for 150-frame buffer
- Check for memory leaks if usage grows continuously
- Monitor with: `ps aux | grep meshradio`

## Development Status

**v0.3-alpha (Current)**
- Stable MP3 broadcasting
- Playlist support
- Web GUI
- Emergency priority system
- Subscription/heartbeat mechanism

**v0.4 (Planned)**
- Multi-listener stress testing
- Station discovery (mDNS/DHT)
- Voice broadcasting refinement
- Metadata broadcasting (now playing, artist)
- Recording/time-shifting
- Mobile clients

## Contributing

MeshRadio is open source and welcomes contributions!

### Areas Needing Help
- Testing on real mesh networks
- Multi-listener scalability testing
- Documentation improvements
- UI/UX enhancements
- Audio quality optimization

### Getting Started
1. Read [DESIGN.md](DESIGN.md) for architecture overview
2. Check GitHub Issues for open tasks
3. Submit PRs with clear descriptions
4. Report bugs with detailed reproduction steps

## License

GPL-3.0 License - see LICENSE file

## Links

- **Repository**: https://github.com/immartian/meshradio
- **Issues**: https://github.com/immartian/meshradio/issues
- **Discussions**: https://github.com/immartian/meshradio/discussions
- **Documentation**: [DESIGN.md](DESIGN.md)

## Credits

Built with:
- Yggdrasil Network (mesh networking)
- Opus Codec (audio compression)
- FFmpeg (media decoding)
- Malgo (audio I/O)
- Go (implementation language)

---

**Note**: This is alpha software under active development. Expect bugs and breaking changes. Feedback and testing reports are highly appreciated!
