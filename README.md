# MeshRadio

Decentralized radio broadcasting over Yggdrasil mesh network.

## What is MeshRadio?

MeshRadio brings HAM radio-style broadcasting to the Yggdrasil mesh network. Use IPv6 addresses as "frequencies" to broadcast and listen to audio streams across the mesh.

**Architecture:** Subscription-based unicast streaming. Listeners "tune" to a broadcaster's IPv6:port, establishing a direct connection. No multicast required.

## Features (v0.4-alpha MVP)

- **Broadcast** audio from your microphone to your Yggdrasil IPv6
- **Listen** to stations by dialing their IPv6:port (subscription-based)
- **Listener tracking** - broadcasters see who's tuned in
- **Automatic heartbeat** - connections maintained automatically
- **Cross-platform** terminal UI and Web GUI
- **Real audio** - PortAudio capture + Opus codec encoding

## Prerequisites

1. **Yggdrasil** must be installed and running
   - Install from: https://yggdrasil-network.github.io/
   - Ensure you're connected to at least one peer

2. **PortAudio** library
   ```bash
   # macOS
   brew install portaudio

   # Ubuntu/Debian
   sudo apt-get install portaudio19-dev

   # Fedora
   sudo dnf install portaudio-devel

   # Arch
   sudo pacman -S portaudio
   ```

## Installation

```bash
go install github.com/meshradio/meshradio@latest
```

Or build from source:
```bash
git clone https://github.com/meshradio/meshradio
cd meshradio
go build -o meshradio ./cmd/meshradio
```

## Quick Start

### 1. Check your Yggdrasil IPv6
```bash
yggdrasilctl getSelf
```

### 2. Start MeshRadio
```bash
meshradio
```

### 3. Broadcast
- Press `b` to enter broadcast mode
- Select your microphone
- Share your IPv6:port with listeners (e.g., `201:abcd:1234::1:8799`)
- See connected listeners in real-time

### 4. Listen
- Press `l` to enter listen mode
- Enter the broadcaster's IPv6:port (get this from the broadcaster)
- Your listener automatically subscribes and sends heartbeats
- Enjoy the stream!

## Usage

```bash
# Interactive TUI mode (default)
meshradio

# Broadcast mode (headless)
meshradio broadcast --ipv6 <your-ipv6>

# Listen mode (headless)
meshradio listen --ipv6 <station-ipv6>

# Show your IPv6
meshradio info
```

## Architecture

See [DESIGN.md](DESIGN.md) for full technical specification.

## Development Status

**Current**: v0.4-alpha (Subscription-based MVP)
- ✅ Subscription-based streaming (unicast over Yggdrasil)
- ✅ Listener tracking with heartbeats
- ✅ Real audio (PortAudio + Opus)
- ✅ Terminal UI and Web GUI
- ⏳ Station discovery (manual OOB for MVP)
- ⏳ DHT registry (future)
- ⏳ Scanning (future)

## Contributing

MeshRadio is open source! Contributions welcome.

## License

GPL-3.0 License - see LICENSE file

## Community

- Documentation: [DESIGN.md](DESIGN.md)
- Issues: GitHub Issues
- Discussions: GitHub Discussions

---

**Note**: This is alpha software. Expect bugs and breaking changes.
