# MeshRadio - Current Status

**Last Updated**: 2025-11-25
**Version**: 0.2-alpha
**Repository**: https://github.com/immartian/meshradio

---

## ✅ What's Working NOW

### Network Layer
- ✅ **Yggdrasil Integration** - Auto-detects IPv6 from yggdrasilctl
- ✅ **Fallback Detection** - Scans network interfaces for Yggdrasil addresses
- ✅ **UDP Transport** - Full send/receive implementation
- ✅ **Multicast Broadcasting** - Sends to ff02::1 (all local nodes)
- ✅ **Packet Transmission** - Real network packets flowing
- ✅ **Beacon System** - Periodic station announcements with metadata

### Protocol Layer
- ✅ **Packet Encoding/Decoding** - Binary protocol implementation
- ✅ **Audio Packets** - Structured audio payload format
- ✅ **Beacon Packets** - Station information broadcasting
- ✅ **Metadata Support** - JSON payloads for station info

### Application Layer
- ✅ **Broadcaster** - Transmits audio frames + beacons
- ✅ **Listener** - Receives and processes packets
- ✅ **Stats Tracking** - Packet counts, sequences, station info

### User Interface
- ✅ **Cross-platform TUI** - Works on Linux/Mac/Windows
- ✅ **Real-time Updates** - UI refreshes every second
- ✅ **Animated Status** - Live connection indicators
- ✅ **Signal Visualization** - Strength bars and audio levels
- ✅ **Activity Logs** - Recent events display
- ✅ **Error Handling** - User-friendly error messages

### Developer Tools
- ✅ **Dependency Checker** - Script to verify system requirements
- ✅ **Makefile** - Build automation
- ✅ **Documentation** - Complete design spec and guides
- ✅ **Git Workflow** - Clean commit history

---

## 🔧 What's Simulated (Still To Do)

### Audio I/O
- ⏳ **Microphone Capture** - Currently generates silence
- ⏳ **Speaker Playback** - Currently discards audio
- ⏳ **PortAudio Integration** - Needs native library bindings

**Status**: Ready for implementation, waiting for PortAudio install

**Implementation Plan**:
1. Install PortAudio: `sudo apt-get install portaudio19-dev`
2. Add Go bindings: `go get github.com/gordonklaus/portaudio`
3. Replace dummy capture/playback in `pkg/audio/stream.go`
4. Test with real microphone/speakers

### Audio Codec
- ⏳ **Opus Encoding** - Currently pass-through (no compression)
- ⏳ **Opus Decoding** - Currently pass-through
- ⏳ **libopus Integration** - Needs native library

**Status**: Ready for implementation, waiting for Opus install

**Implementation Plan**:
1. Install Opus: `sudo apt-get install libopus-dev`
2. Add Go bindings: `go get gopkg.in/hraban/opus.v2`
3. Replace DummyCodec in `pkg/audio/codec.go`
4. Test compression ratios and quality

---

## 🚀 What You Can Test RIGHT NOW

### Local Testing (Same Machine)

**Terminal 1 - Broadcaster**:
```bash
export MESHRADIO_CALLSIGN="STATION1"
./meshradio
# Press 'b' to start broadcasting
```

**Terminal 2 - Listener**:
```bash
export MESHRADIO_CALLSIGN="STATION2"
./meshradio
# Press 'l', then enter your Yggdrasil IPv6
# (get it from broadcaster's screen)
```

**What You'll See**:
- Broadcaster transmitting packets every ~20ms
- Listener receiving packets (if on same network)
- Beacon announcements every 30 seconds
- Real-time packet counters
- Signal strength indicators
- Animated status dots

### Network Testing (Different Machines)

**Requirements**:
- Both machines on Yggdrasil mesh
- Both machines can ping each other's IPv6

**Setup**:
1. Machine A: Start broadcaster
2. Note Machine A's IPv6 address (shown in UI)
3. Machine B: Start listener, enter Machine A's IPv6
4. Watch packets flow!

---

## 📊 Current Capabilities

| Feature | Status | Notes |
|---------|--------|-------|
| Yggdrasil Detection | ✅ Working | Auto-detects IPv6 |
| Network Transmission | ✅ Working | UDP multicast |
| Packet Protocol | ✅ Working | Binary encoding |
| Broadcaster | ✅ Working | Transmits frames |
| Listener | ✅ Working | Receives frames |
| TUI | ✅ Working | Real-time updates |
| Audio Capture | ⏳ Simulated | Needs PortAudio |
| Audio Playback | ⏳ Simulated | Needs PortAudio |
| Audio Codec | ⏳ Pass-through | Needs Opus |
| Station Discovery | ❌ Not Started | Phase 2 |
| IPv6 Scanning | ❌ Not Started | Phase 2 |
| DHT Registry | ❌ Not Started | Phase 3 |

---

## 🎯 Next Milestones

### Milestone 1: Real Audio (High Priority)
**Goal**: Stream actual audio from mic to speakers

**Tasks**:
- [ ] Install PortAudio library
- [ ] Integrate PortAudio Go bindings
- [ ] Replace simulated capture with real capture
- [ ] Replace simulated playback with real playback
- [ ] Test with headphones to avoid feedback
- [ ] Add audio device selection in UI

**Estimated Effort**: 1-2 days
**Blocker**: Need to install system libraries

### Milestone 2: Opus Codec (High Priority)
**Goal**: Compress audio for efficient transmission

**Tasks**:
- [ ] Install Opus library
- [ ] Integrate Opus Go bindings
- [ ] Replace DummyCodec with OpusCodec
- [ ] Configure bitrate/quality settings
- [ ] Test compression ratios
- [ ] Measure CPU usage

**Estimated Effort**: 1 day
**Blocker**: Need to install system libraries

### Milestone 3: Production Ready (Medium Priority)
**Goal**: Stable release with all core features

**Tasks**:
- [ ] Handle packet loss gracefully
- [ ] Implement jitter buffer
- [ ] Add audio level metering
- [ ] Support multiple listeners
- [ ] Error recovery
- [ ] Performance optimization
- [ ] Comprehensive testing

**Estimated Effort**: 3-5 days

### Milestone 4: Discovery (Medium Priority)
**Goal**: Find stations automatically

**Tasks**:
- [ ] IPv6 range scanner
- [ ] Station database
- [ ] Beacon listener
- [ ] DHT implementation
- [ ] Bookmark system

**Estimated Effort**: 5-7 days

---

## 🔬 Technical Details

### Current Architecture

```
Application Layer
├── Broadcaster (transmits)
│   ├── Audio Input (simulated)
│   ├── Codec (pass-through)
│   ├── Protocol (packets)
│   └── Network (UDP multicast)
│
└── Listener (receives)
    ├── Network (UDP receive)
    ├── Protocol (parse)
    ├── Codec (pass-through)
    └── Audio Output (simulated)

Network Layer
├── Yggdrasil Detection ✅
├── UDP Transport ✅
├── Multicast Support ✅
└── Packet Send/Receive ✅

UI Layer
├── Bubbletea TUI ✅
├── Real-time Updates ✅
├── Status Visualization ✅
└── Error Handling ✅
```

### Performance Metrics

**Current**:
- Frame Size: 960 samples (20ms @ 48kHz)
- Frame Rate: ~50 frames/second
- Data Rate: ~1.8 Mbps uncompressed (simulated)
- Network: UDP multicast (ff02::1)
- Latency: <100ms (theoretical)

**With Opus** (expected):
- Compressed Rate: ~64 kbps
- Compression: ~28x reduction
- Quality: Good (voice optimized)
- CPU: <5% on modern hardware

---

## 🐛 Known Issues

### Current
1. **Audio is simulated** - No real mic/speaker support yet
2. **No compression** - Bandwidth usage high (if real audio)
3. **Multicast only** - No direct station-to-station yet
4. **Single connection** - Can't broadcast to multiple specific listeners
5. **No discovery** - Must know IPv6 address manually

### Non-Issues
- ✅ Network transmission works
- ✅ Yggdrasil detection works
- ✅ Packets flow correctly
- ✅ UI updates in real-time
- ✅ Protocol encoding solid

---

## 📦 Dependencies Status

### Installed & Working
- ✅ Go 1.24.10
- ✅ Yggdrasil (daemon running)
- ✅ Bubbletea (Go package)
- ✅ Lipgloss (Go package)

### Needed for Real Audio
- ❌ PortAudio (system library)
- ❌ Opus (system library)

**Install Command (Ubuntu/Debian)**:
```bash
sudo apt-get install portaudio19-dev libopus-dev
```

**Install Command (Fedora)**:
```bash
sudo dnf install portaudio-devel opus-devel
```

**Install Command (Arch)**:
```bash
sudo pacman -S portaudio opus
```

---

## 🎓 How to Contribute

### For Testing
1. Clone the repo
2. Build: `make build`
3. Run: `./meshradio YOUR_CALLSIGN`
4. Test broadcasting and listening
5. Report issues on GitHub

### For Development
1. Pick a task from Milestones above
2. Fork the repository
3. Create a feature branch
4. Implement and test
5. Submit a pull request

### Priority Areas
1. **PortAudio Integration** - Most needed!
2. **Opus Codec** - Second priority
3. **Discovery System** - Phase 2
4. **Documentation** - Always welcome
5. **Testing** - Write tests!

---

## 🚦 Roadmap

### v0.2-alpha (Current)
- ✅ Yggdrasil integration
- ✅ Network transmission
- ✅ Enhanced UI

### v0.3-alpha (Next - Real Audio)
- ⏳ PortAudio integration
- ⏳ Opus codec
- ⏳ Real audio streaming

### v0.4-alpha (Discovery)
- ⏳ IPv6 scanning
- ⏳ Station discovery
- ⏳ Beacon listening

### v0.5-beta (Polish)
- ⏳ Multiple listeners
- ⏳ Jitter buffer
- ⏳ Error recovery
- ⏳ Performance tuning

### v1.0 (Production)
- ⏳ All features stable
- ⏳ Full documentation
- ⏳ Comprehensive tests
- ⏳ Package releases

---

## 🎉 Success So Far

We've achieved:
- ✅ Complete protocol implementation
- ✅ Real network transmission
- ✅ Yggdrasil integration
- ✅ Working broadcaster/listener
- ✅ Beautiful cross-platform UI
- ✅ Real-time status updates
- ✅ Modular, extensible architecture
- ✅ Clean codebase with documentation

**MeshRadio is real and working!** 🚀

The core concept is proven. Now we just need real audio to make it truly useful.

---

**Ready to help?** Start with Milestone 1 (Real Audio) - it's the biggest impact item!

**Questions?** Open an issue on GitHub: https://github.com/immartian/meshradio/issues
