# MeshRadio Installation & Usage

## Quick Start

### 1. Build the Binary

```bash
cd /media/im3/plus/labx/meshradio
export PATH=$PATH:/usr/local/go/bin
make build
```

This creates the `meshradio` binary (5.3MB).

### 2. Run MeshRadio

**Option A: From your terminal (recommended)**
```bash
# In a real terminal (not through automation)
./meshradio MYCALLSIGN
```

**Option B: Using go run**
```bash
go run ./cmd/meshradio MYCALLSIGN
```

**Option C: Install system-wide**
```bash
sudo make install
# Then run from anywhere:
meshradio MYCALLSIGN
```

### 3. Fix "Permission Denied" Issues

If you get permission errors, the filesystem might be mounted with `noexec`. Solutions:

**A. Copy to /tmp (executable filesystem)**
```bash
cp meshradio /tmp/
cd /tmp
./meshradio MYCALLSIGN
```

**B. Use go run instead**
```bash
cd /media/im3/plus/labx/meshradio
go run ./cmd/meshradio MYCALLSIGN
```

**C. Install to system path**
```bash
sudo cp meshradio /usr/local/bin/
meshradio MYCALLSIGN
```

## Running MeshRadio

### Basic Usage

```bash
# Set your callsign
export MESHRADIO_CALLSIGN="W1TEST"

# Run the TUI
./meshradio

# Or pass callsign as argument
./meshradio W1TEST
```

### Interactive Commands

Once running, you'll see the main menu:

- **[b]** - Start Broadcasting
  - Transmits audio to Yggdrasil multicast (ff02::1)
  - Shows real-time packet transmission
  - Sends beacons every 30 seconds

- **[l]** - Listen to Station
  - Prompts for IPv6 address
  - Receives audio stream
  - Shows packet stats and signal quality

- **[i]** - Show Info
  - Display your callsign
  - Show your Yggdrasil IPv6 address

- **[q]** - Quit

### Testing Locally

**Terminal 1 - Broadcaster:**
```bash
export MESHRADIO_CALLSIGN="STATION1"
./meshradio
# Press 'b' to start broadcasting
# Note your IPv6 address shown in the UI
```

**Terminal 2 - Listener:**
```bash
export MESHRADIO_CALLSIGN="STATION2"
./meshradio
# Press 'l' to listen
# Enter the broadcaster's IPv6 (from Terminal 1)
# Or use ::1 for localhost testing
```

### Testing Over Network

**Requirements:**
- Both machines on Yggdrasil mesh
- Both machines can ping each other

**Steps:**
1. Machine A: Start MeshRadio and begin broadcasting
2. Machine A: Note the IPv6 shown in the UI
3. Machine B: Start MeshRadio
4. Machine B: Press 'l' and enter Machine A's IPv6
5. Watch packets flow!

## What You'll See

### Broadcasting Mode
```
● BROADCASTING ...

Station:  W1TEST
Address:  201:abcd:1234::1:9001
Codec:    Opus (simulated)
Quality:  48kHz, Mono, 64kbps
Multicast: ff02::1 (all local nodes)

Audio Level: ▓▓▓▓▓░░░░░

📡 Transmitting audio frames...
   Network: UDP multicast
   Status: Active
```

### Listening Mode
```
● LISTENING ...

Station: W2XYZ
Packets: 150 | Sequence: 42

Signal:  ▓▓▓▓▓▓▓░░░
Network: Receiving on port 9002

🎧 Listening for audio...
   Codec: Opus
   Buffer: Good
```

## Current Limitations

### Audio is Simulated
- Microphone: Generates silence (no real capture)
- Speakers: Discards audio (no real playback)
- **Why**: PortAudio library not installed yet
- **Solution**: Install PortAudio (see below)

### No Audio Compression
- Codec: Pass-through (no compression)
- **Why**: Opus library not installed yet
- **Solution**: Install Opus (see below)

### What DOES Work
- ✅ Real network transmission over Yggdrasil
- ✅ UDP multicast broadcasting
- ✅ Packet encoding/decoding
- ✅ Real IPv6 detection
- ✅ Station beacons
- ✅ Packet statistics
- ✅ Beautiful interactive UI

## Installing Audio Dependencies (Optional)

To enable **real audio** capture and playback:

### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install portaudio19-dev libopus-dev
```

### Fedora
```bash
sudo dnf install portaudio-devel opus-devel
```

### Arch Linux
```bash
sudo pacman -S portaudio opus
```

### macOS
```bash
brew install portaudio opus
```

After installing, rebuild MeshRadio:
```bash
make clean
make build
```

## Troubleshooting

### "Permission denied" when running ./meshradio
- Filesystem mounted with noexec
- Solution: Use `go run` or copy to /tmp

### "could not open a new TTY"
- Not running in a real terminal
- Solution: Run in actual terminal window, not through script

### "Could not detect Yggdrasil IPv6"
- Yggdrasil not installed or not running
- Solution: Install Yggdrasil or use fallback localhost (::1)

### No audio
- This is expected! Audio I/O is simulated
- Solution: Install PortAudio/Opus libraries (optional)

### Can't connect to broadcaster
- Check both machines are on Yggdrasil mesh
- Verify IPv6 address is correct
- Try localhost (::1) for same-machine testing

## Check System Status

Run the dependency checker:
```bash
bash scripts/check-deps.sh
```

This shows:
- ✅ Go version
- ✅ Yggdrasil status
- ⚠️ PortAudio status
- ⚠️ Opus status

## Next Steps

1. **Test the UI** - Run in your terminal and explore
2. **Test Broadcasting** - Start a broadcaster
3. **Test Listening** - Connect a listener
4. **Try Local Test** - Two terminals on same machine
5. **Install Audio** - Add PortAudio for real audio (optional)

## Support

- **GitHub**: https://github.com/immartian/meshradio
- **Issues**: https://github.com/immartian/meshradio/issues
- **Docs**: See DESIGN.md, QUICKSTART.md, STATUS.md

---

**Ready to broadcast on the mesh!** 📻
