# MeshRadio Real Audio Setup

## Current Status

✅ **MeshRadio is working with simulated audio**
🎯 **This guide adds REAL microphone and speaker support**

## Quick Start - Enable Real Audio

### Step 1: Install System Libraries

Run the installation script:

```bash
cd /media/im3/plus/labx/meshradio
bash scripts/install-audio-deps.sh
```

Or install manually:

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y portaudio19-dev libopus-dev
```

**Fedora/RHEL:**
```bash
sudo dnf install -y portaudio-devel opus-devel
```

**Arch Linux:**
```bash
sudo pacman -S portaudio opus
```

### Step 2: Install Go Audio Bindings

```bash
cd /media/im3/plus/labx/meshradio

# Install PortAudio Go bindings
go get github.com/gordonklaus/portaudio

# Install Opus Go bindings
go get gopkg.in/hraban/opus.v2
```

### Step 3: Rebuild with Audio Support

```bash
# Clean previous build
make clean

# Build with audio tags
go build -tags "portaudio opus" -o meshradio ./cmd/meshradio

# Or use make (will auto-detect libraries)
make build-audio
```

### Step 4: Test Real Audio

```bash
# Copy to executable location
cp meshradio /tmp/

# Run
/tmp/meshradio MYCALLSIGN
```

You should now see:
```
🎤 Using real microphone (PortAudio)
🔊 Using real speakers (PortAudio)
🎵 Using Opus codec
```

## What Changes With Real Audio

### Before (Simulated)
- ❌ Silence generated instead of mic input
- ❌ Audio discarded instead of played to speakers
- ❌ No compression (would use ~1.8 Mbps)

### After (Real Audio)
- ✅ Real microphone capture
- ✅ Real speaker playback
- ✅ Opus compression (~64 kbps)
- ✅ 28x bandwidth reduction
- ✅ End-to-end voice communication

## Testing Real Audio

### Test 1: Loopback (Same Machine)

**Terminal 1 - Broadcaster:**
```bash
export MESHRADIO_CALLSIGN="STATION1"
/tmp/meshradio
# Press 'b' to broadcast
# Speak into your microphone
```

**Terminal 2 - Listener:**
```bash
export MESHRADIO_CALLSIGN="STATION2"
/tmp/meshradio
# Press 'l'
# Enter: ::1 (localhost)
# You should hear Terminal 1's audio!
```

**Important**: Use headphones to avoid feedback!

### Test 2: Network (Different Machines)

Both machines need:
- Yggdrasil installed and running
- MeshRadio built with audio support
- Reachable over Yggdrasil mesh

**Machine A:**
```bash
/tmp/meshradio STATION_A
# Press 'b' to broadcast
# Note your IPv6 address shown in UI
```

**Machine B:**
```bash
/tmp/meshradio STATION_B
# Press 'l'
# Enter Machine A's IPv6
# Listen to the broadcast!
```

## Audio Configuration

### Default Settings (Voice Optimized)

- **Sample Rate**: 48 kHz
- **Channels**: Mono
- **Frame Size**: 960 samples (20ms)
- **Opus Bitrate**: 64 kbps
- **Opus Mode**: VoIP optimized
- **FEC**: Enabled (forward error correction)
- **DTX**: Enabled (discontinuous transmission)

### Performance Metrics

**Uncompressed PCM:**
- Data rate: 48000 Hz × 16 bit × 1 channel = 768 kbps
- Frame size: 1920 bytes (960 samples × 2 bytes)

**With Opus Compression:**
- Data rate: ~64 kbps (configured)
- Frame size: ~160 bytes (typical)
- Compression ratio: ~12:1
- Network bandwidth: ~80 kbps with overhead

## Troubleshooting

### No Audio Devices Found

```bash
# List available devices
pactl list sources short   # Input devices
pactl list sinks short     # Output devices
```

### Permission Denied (Audio Devices)

Add your user to audio group:
```bash
sudo usermod -a -G audio $USER
# Log out and back in
```

### Build Errors

**Error**: `portaudio.h: No such file or directory`
- **Solution**: Install portaudio19-dev

**Error**: `opus.h: No such file or directory`
- **Solution**: Install libopus-dev

**Error**: `cannot find package`
- **Solution**: Run `go get` commands for Go bindings

### Audio Quality Issues

**Choppy Audio:**
- Network latency too high
- Try increasing buffer size
- Check packet loss

**Echo/Feedback:**
- Use headphones
- Or mute speakers when broadcasting

**Low Volume:**
- Adjust system volume
- Check microphone levels: `alsamixer`

## Advanced Configuration

### Custom Audio Settings

Edit `pkg/audio/stream.go`:

```go
func DefaultConfig() StreamConfig {
    return StreamConfig{
        SampleRate: 48000,  // Change sample rate
        Channels:   2,      // Stereo instead of mono
        Bitrate:    96000,  // Higher quality
        FrameSize:  960,    // Keep at 20ms
    }
}
```

### Select Specific Devices

Future feature - will add to UI:
- List available devices
- Choose input device
- Choose output device

## Build Options

### Build Without Audio (Simulated)

```bash
go build -o meshradio ./cmd/meshradio
```

Uses simulated audio (current default).

### Build With PortAudio Only

```bash
go build -tags "portaudio" -o meshradio ./cmd/meshradio
```

Real audio I/O, but no compression.

### Build With Opus Only

```bash
go build -tags "opus" -o meshradio ./cmd/meshradio
```

Compression, but simulated I/O (not very useful).

### Build With Full Audio (Recommended)

```bash
go build -tags "portaudio opus" -o meshradio ./cmd/meshradio
```

Real audio + compression.

### Auto-Detect Build

```bash
make build-audio
```

Automatically detects available libraries and builds with support.

## Makefile Targets

```bash
make build              # Standard build (simulated audio)
make build-audio        # Build with audio support (auto-detect)
make build-full         # Force build with portaudio+opus
make test-audio         # Test audio devices
make clean              # Clean build artifacts
```

## Verifying Audio Support

Run the dependency checker:

```bash
bash scripts/check-deps.sh
```

Should show:
```
✅ Go: go1.24.10
✅ Yggdrasil: Installed
   ✅ Yggdrasil daemon: Running
✅ PortAudio: v19.x.x
✅ Opus: v1.x.x
```

## Next Steps After Audio Works

1. **Test locally** with loopback
2. **Test over network** with another machine
3. **Adjust audio quality** in settings
4. **Report issues** on GitHub
5. **Help others** set up audio

## Getting Help

- **GitHub Issues**: https://github.com/immartian/meshradio/issues
- **Check logs**: Enable verbose logging
- **Test devices**: Use `arecord` and `aplay` to verify hardware

---

**Ready to broadcast with REAL audio!** 🎙️📻
