# MeshRadio Quick Start Guide

## What You've Built

A working MVP of MeshRadio - a HAM radio-style broadcasting system over Yggdrasil mesh network!

## Features (MVP v0.1)

✅ **Broadcaster** - Stream audio from your station
✅ **Listener** - Tune into other stations by IPv6
✅ **Cross-platform TUI** - Beautiful terminal interface
✅ **Protocol** - Full packet encoding/decoding
✅ **Audio Pipeline** - Ready for Opus codec integration

## Running MeshRadio

### 1. Set Your Callsign

```bash
export MESHRADIO_CALLSIGN="W1AW"
```

Or pass as argument:
```bash
./meshradio MYCALLSIGN
```

### 2. Launch the TUI

```bash
./meshradio
```

### 3. Main Menu Options

- **[b]** - Start Broadcasting
  - Broadcasts on your Yggdrasil IPv6 address
  - Port: 9001
  - Simulated audio (silence) for MVP

- **[l]** - Listen to Station
  - Enter target IPv6 address
  - Receives and "plays" audio stream

- **[i]** - Show Info
  - Display your callsign and IPv6

- **[q]** - Quit

### 4. Broadcasting

Press `b` to start broadcasting:
```
● BROADCASTING

Station: W1AW
Address: 200::1:9001
Codec:   Opus (simulated)
Quality: 48kHz, Mono, 64kbps

Press 'q' or ESC to stop broadcasting
```

Your station is now live! Others can connect to your IPv6 address.

### 5. Listening

Press `l` and enter a station's IPv6:
```
Enter station IPv6 to listen:
200:1234:5678:abcd::1

(Press ESC to cancel)
```

Once connected:
```
● LISTENING

Station: W2XYZ
Packets: 150 | Last Seq: 42

Press 'q' or ESC to stop listening
```

## Testing Locally

### Terminal 1 (Broadcaster)
```bash
export MESHRADIO_CALLSIGN="STATION1"
./meshradio
# Press 'b' to broadcast
```

### Terminal 2 (Listener)
```bash
export MESHRADIO_CALLSIGN="STATION2"
./meshradio
# Press 'l' and enter: ::1 (localhost)
```

## What's Simulated (MVP)

🔧 **Audio Capture** - Currently generates silence
- Replace with PortAudio in production
- Location: `pkg/audio/stream.go`

🔧 **Audio Playback** - Currently discards received audio
- Replace with PortAudio in production
- Location: `pkg/audio/stream.go`

🔧 **Codec** - Pass-through (no compression)
- Integrate libopus for real codec
- Location: `pkg/audio/codec.go`

🔧 **Yggdrasil Integration** - Uses placeholder IPv6
- Query yggdrasilctl for real address
- Location: `cmd/meshradio/main.go:getLocalIPv6()`

## Next Steps

### Phase 1: Real Audio
1. Install PortAudio bindings
2. Implement real capture/playback
3. Integrate Opus codec

### Phase 2: Yggdrasil Integration
1. Query yggdrasilctl for IPv6
2. Auto-detect Yggdrasil daemon
3. Use real mesh routing

### Phase 3: Discovery
1. Implement scanner
2. Add station database
3. DHT for station registry

### Phase 4: Features
1. Bookmarks
2. Signal quality metrics
3. Station metadata
4. Recording

## Architecture

```
meshradio/
├── cmd/meshradio/           # Main entry point
├── pkg/
│   ├── protocol/            # Packet format & encoding
│   ├── audio/               # Audio streaming & codecs
│   ├── network/             # UDP transport layer
│   └── ui/                  # Bubbletea TUI
└── internal/
    ├── broadcaster/         # Broadcasting logic
    └── listener/            # Listening logic
```

## Troubleshooting

### Build Errors

```bash
# Ensure Go is in PATH
export PATH=$PATH:/usr/local/go/bin

# Tidy modules
go mod tidy

# Rebuild
go build -o meshradio ./cmd/meshradio
```

### Runtime Issues

1. **Cannot bind to port 9001**
   - Port might be in use
   - Try: `sudo lsof -i :9001`

2. **No audio**
   - This is expected in MVP!
   - Audio is simulated (silence)

3. **Connection refused**
   - Check target IPv6 is correct
   - Ensure broadcaster is running
   - Test with localhost (::1) first

## Code Tour

### Broadcasting Flow
```
main.go → ui/model.go → broadcaster/broadcaster.go
                      ↓
                audio/stream.go (capture)
                      ↓
                audio/codec.go (encode)
                      ↓
                protocol/packet.go (packetize)
                      ↓
                network/transport.go (send UDP)
```

### Listening Flow
```
main.go → ui/model.go → listener/listener.go
                      ↓
                network/transport.go (receive UDP)
                      ↓
                protocol/packet.go (parse)
                      ↓
                audio/codec.go (decode)
                      ↓
                audio/stream.go (playback)
```

## Contributing

Want to help? Here are great starting points:

1. **Add real audio I/O** - Integrate PortAudio
2. **Opus codec** - Add libopus bindings
3. **Yggdrasil query** - Parse yggdrasilctl output
4. **Scanner** - Implement IPv6 range scanning
5. **Better UI** - Add more visualizations

## License

GPL-3.0 - See LICENSE file

---

**Status**: MVP Complete ✅
**Version**: 0.1-alpha
**Date**: 2025-11-25

Ready to broadcast on the mesh! 📻
