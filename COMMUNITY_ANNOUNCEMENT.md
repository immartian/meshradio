# MeshRadio v0.3-alpha: Stable Music Streaming Over Mesh Networks

## We Need Your Help Testing!

After intensive development work, MeshRadio has achieved a major milestone: **stable, continuous music streaming** over mesh networks. The system now broadcasts and plays MP3 files smoothly without interruption.

We're reaching out to the community for testing, feedback, and collaboration.

## What is MeshRadio?

MeshRadio is a **decentralized internet radio system** designed for mesh networks like Yggdrasil. Think of it as community radio without centralized servers:

- No central infrastructure required
- Peer-to-peer encrypted audio streaming
- High-quality audio using Opus codec (128kbps)
- Open source (GPL-3.0)
- Cross-platform (Linux, macOS, Windows)

## What Works Now

The v0.3-alpha release includes:

**Music Broadcasting**
- Stream MP3 files from your music library
- Automatic playlist cycling
- High-quality Opus compression (12-13x ratio)
- Stable multi-hour broadcasts

**Listening**
- Subscribe to broadcasters by IPv6:port
- 3-second buffer for network jitter tolerance
- Continuous playback without interruption
- Automatic heartbeat maintenance

**Management**
- Interactive terminal UI
- Web-based GUI (http://localhost:8799)
- Emergency priority system for critical broadcasts
- Real-time listener tracking

## Recent Technical Achievements

This week's development session resolved critical stability issues:

**Audio Playback Stability**
- Fixed race condition in audio callback causing spurious buffer underruns
- Implemented precise packet pacing (20ms intervals)
- Changed audio callback to use timeout-based select
- Result: Continuous playback tested with multi-hour sessions

**Smooth Streaming**
- Eliminated packet bursting through ticker-based pacing
- Increased buffer from 1 to 3 seconds (50 to 150 frames)
- Optimized FFmpeg realtime decoding
- Result: Smooth, jitter-free audio delivery

**Playlist Management**
- Songs advance automatically when finished
- Playlist loops back to beginning after last track
- Tested with 93-song jazz collection

## Technical Stack

For those interested in the internals:

**Audio Pipeline**
```
Broadcaster: MP3 → FFmpeg → PCM → Opus → UDP/IPv6
Listener: UDP/IPv6 → Opus → PCM → Audio Device
```

**Key Technologies**
- Language: Go
- Audio Codec: Opus (RFC 6716)
- Decoder: FFmpeg
- Audio I/O: Malgo (miniaudio bindings)
- Network: UDP over IPv6
- Transport: Custom protocol with subscribe/heartbeat

**Performance**
- Compression: 12-13x (3840 bytes → ~300 bytes per frame)
- Bitrate: 128kbps for music quality
- Frame Rate: 50 fps (20ms per frame)
- Buffer: 3 seconds latency
- Sample Rate: 48kHz stereo

## Why This Matters

MeshRadio enables scenarios that traditional internet radio cannot:

**Resilience**
- Operates during internet outages
- No dependency on centralized services
- Community-owned infrastructure

**Privacy**
- No tracking or analytics
- Encrypted peer-to-peer communication
- No third-party intermediaries

**Community**
- Local music scenes can share without platforms
- Emergency broadcasts during disasters
- Neighborhood radio stations
- Educational mesh networking demonstrations

## We Need Your Help

The project is at a critical stage where community testing is essential. Here's what we need:

### Testers Wanted

**Geographic Diversity**
- Different regions and network topologies
- Various ISPs and connection types
- Urban vs rural environments
- Different Yggdrasil peer configurations

**Platform Testing**
- Linux distributions (Ubuntu, Fedora, Arch, etc.)
- macOS (Intel and Apple Silicon)
- Windows (native and WSL)
- Different audio hardware

**Stress Testing**
- Multiple simultaneous listeners
- Long-duration broadcasts (24+ hours)
- Various music formats and bitrates
- Network interruption recovery

### Contributors Needed

**Development**
- Go developers for core functionality
- Audio engineers for codec optimization
- Network protocol designers
- UI/UX designers for GUI improvements

**Documentation**
- Technical writers for user guides
- Video tutorial creators
- Translation to other languages
- Architecture documentation

**Testing and QA**
- Systematic bug reporting
- Performance benchmarking
- Security auditing
- Accessibility testing

### Specific Questions We Need Answered

1. **Scalability**: How many listeners can one broadcaster support?
2. **Reliability**: Does it work across different mesh topologies?
3. **Compatibility**: Which platforms have issues?
4. **Usability**: What features are missing or confusing?
5. **Performance**: What's the CPU/memory/bandwidth usage in production?

## Getting Started

### Quick Test (No Yggdrasil Required)

You can test on localhost without installing Yggdrasil:

```bash
# Clone and build
git clone https://github.com/immartian/meshradio.git
cd meshradio
go build -o meshradio ./cmd/meshradio

# Terminal 1 - Start broadcaster
./meshradio
# Select "1" for Music Broadcaster

# Terminal 2 - Start listener
./meshradio
# Select "3" for Listener
# Enter IPv6: ::1
# Enter port: 8799
```

### Testing Over Yggdrasil

If you're already on Yggdrasil:

1. Get your IPv6: `yggdrasilctl getSelf`
2. Start broadcaster with your music directory
3. Share your IPv6:port with test listeners
4. Report results via GitHub Issues

### Prerequisites

- Go 1.21 or later
- FFmpeg installed
- Yggdrasil (for mesh testing)
- Music files (MP3 format)

## Reporting Issues and Feedback

**GitHub Repository**
https://github.com/immartian/meshradio

**What to Include in Bug Reports**
- Operating system and version
- Go version (`go version`)
- FFmpeg version (`ffmpeg -version`)
- Yggdrasil version (if applicable)
- Detailed steps to reproduce
- Log output (full or relevant excerpts)
- Expected vs actual behavior

**Feature Requests**
- Use GitHub Discussions for feature ideas
- Explain the use case and benefit
- Consider contributing a pull request

## Roadmap

**v0.3 (Current)**
- Stable MP3 broadcasting
- Playlist support
- Basic subscription system
- Web GUI

**v0.4 (Next, Q1 2025)**
- Multi-listener stress testing results
- Station discovery (mDNS/DHT)
- Voice broadcasting improvements
- Metadata broadcasting (now playing info)
- Enhanced Web GUI

**Future**
- Mobile clients (Android/iOS)
- Recording and time-shifting
- DJ mode with live mixing
- Multi-stream support
- Podcast support
- Federation features

## Join the Community

We're building this in the open and welcome all contributors:

**GitHub**
- Repository: https://github.com/immartian/meshradio
- Issues: Bug reports and feature requests
- Discussions: Questions and ideas
- Pull Requests: Code contributions welcome

**Communication**
- GitHub Discussions for async communication
- Issues for bug tracking
- Pull requests with clear descriptions

**Documentation**
- README.md: Getting started guide
- DESIGN.md: Architecture documentation
- Issues: Known problems and solutions

## How to Contribute

1. **Test It**: Download, build, run, report results
2. **Break It**: Find bugs and edge cases
3. **Fix It**: Submit pull requests
4. **Document It**: Improve guides and tutorials
5. **Share It**: Tell others about the project

## Technical Deep Dive

For developers interested in contributing, here are key areas:

**Audio Processing** (`pkg/audio/`)
- FFmpeg integration for MP3 decoding
- Opus codec integration
- Audio buffer management
- Playback callback timing

**Network Layer** (`pkg/network/`)
- UDP socket management
- IPv6 addressing
- Packet serialization
- Transport reliability

**Broadcasting** (`internal/broadcaster/`)
- Subscription management
- Heartbeat handling
- Listener tracking
- Packet pacing and timing

**Listening** (`internal/listener/`)
- Subscription protocol
- Audio decoding pipeline
- Buffer management
- Playback synchronization

## Acknowledgments

This project builds on excellent open source technologies:

- **Yggdrasil Network**: Mesh networking layer
- **Opus Codec**: High-quality audio compression
- **FFmpeg**: Media decoding and processing
- **Malgo**: Cross-platform audio I/O
- **Go**: Implementation language

Special thanks to all early testers and contributors.

## License

MeshRadio is licensed under GPL-3.0. See LICENSE file for details.

## Call to Action

If you believe in decentralized community broadcasting, we need your help:

1. **Star the repository** on GitHub
2. **Test the software** and report results
3. **Share with others** interested in mesh networking
4. **Contribute** code, documentation, or testing
5. **Provide feedback** on features and usability

Together we can build the future of community radio.

---

**Repository**: https://github.com/immartian/meshradio

**Try it. Test it. Break it. Fix it. Share it.**

Built with passion for mesh networks and community broadcasting.
