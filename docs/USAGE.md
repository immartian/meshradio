# MeshRadio Usage Guide

## Running MeshRadio

MeshRadio has **two interfaces**: Terminal UI (TUI) and Web GUI.

### Terminal UI (Default)

```bash
# Run with TUI
./meshradio MYCALLSIGN

# Or with environment variable
export MESHRADIO_CALLSIGN="W1TEST"
./meshradio
```

**Keyboard controls:**
- **[b]** - Start Broadcasting
- **[l]** - Listen to station
- **[i]** - Show info
- **[q]** - Quit

### Web GUI (Graphical)

```bash
# Run with web GUI
./meshradio --gui

# With custom port
./meshradio --gui --port 9000

# With callsign
./meshradio --gui MYCALLSIGN
```

Then open your browser to: **http://localhost:8080**

**Features:**
- Beautiful modern interface
- Real-time updates via WebSocket
- Signal meters and visualizations
- Activity log
- One-click controls

## Command Line Options

```
Usage: ./meshradio [OPTIONS] [CALLSIGN]

Options:
  --gui          Launch web GUI instead of terminal UI
  --port PORT    Web GUI port (default: 8080)

Arguments:
  CALLSIGN       Your station callsign (or use $MESHRADIO_CALLSIGN)

Examples:
  ./meshradio W1AW                    # TUI with callsign W1AW
  ./meshradio --gui                   # Web GUI with default callsign
  ./meshradio --gui --port 9000 W2XYZ # Web GUI on port 9000
```

## Quick Comparisonnon

### Terminal UI
- ✅ Fast and lightweight
- ✅ SSH-friendly
- ✅ Keyboard shortcuts
- ✅ No browser needed
- 👍 Best for: Remote access, terminals, old-school feel

### Web GUI
- ✅ Beautiful visuals
- ✅ Mouse-friendly
- ✅ Real-time meters
- ✅ Easy for beginners
- 👍 Best for: Local use, demonstrations, new users

**Use both!** Run TUI on your server, GUI on your desktop.

## Broadcasting

### Via TUI
```
1. Run: ./meshradio YOURCALL
2. Press: b
3. Speak into microphone
4. Press: q or ESC to stop
```

### Via GUI
```
1. Run: ./meshradio --gui YOURCALL
2. Open: http://localhost:8080
3. Click: "Start Broadcasting"
4. Speak into microphone
5. Click: "Stop Broadcasting"
```

## Listening

### Via TUI
```
1. Run: ./meshradio YOURCALL
2. Press: l
3. Enter: Station IPv6 address
4. Listen to audio
5. Press: q or ESC to stop
```

### Via GUI
```
1. Run: ./meshradio --gui YOURCALL
2. Open: http://localhost:8080
3. Enter: Station IPv6 in text field
4. Click: "Start Listening"
5. Watch signal meter
6. Click: "Stop Listening"
```

## Examples

### Local Test (Same Machine)

**Terminal 1 - Broadcaster (TUI):**
```bash
export MESHRADIO_CALLSIGN="STATION1"
/tmp/meshradio
# Press 'b' to broadcast
```

**Terminal 2 - Listener (GUI):**
```bash
/tmp/meshradio --gui STATION2
# Open browser, enter: ::1
# Click "Start Listening"
```

### Network Test (Different Machines)

**Machine A - Broadcaster:**
```bash
/tmp/meshradio --gui W1BROADCAST
# Note your IPv6 from the web page
# Click "Start Broadcasting"
```

**Machine B - Listener:**
```bash
/tmp/meshradio W2LISTEN
# Press 'l'
# Enter Machine A's IPv6
```

## Environment Variables

```bash
# Set your callsign
export MESHRADIO_CALLSIGN="W1AW"

# Force IPv6 (override detection)
export YGGDRASIL_IPV6="201:abcd:1234::1"
```

## Tips & Tricks

### Run GUI in Background
```bash
./meshradio --gui > /dev/null 2>&1 &
# Opens web interface, runs in background
```

### Access GUI from Another Machine
```bash
# On server
./meshradio --gui --port 8080

# From client (if firewall allows)
# Open: http://server-ip:8080
```

### Use Both Interfaces
```bash
# Terminal 1: Monitor with TUI
./meshradio MYCALL

# Terminal 2: Control with GUI
./meshradio --gui MYCALL --port 8081
```

### Quick Status Check
```bash
# TUI: Press 'i'
# GUI: Look at status bar (always visible)
```

## Troubleshooting

### Can't See TUI
- Make sure you're in a real terminal
- Don't pipe or redirect
- SSH works fine

### Can't Access GUI
- Check firewall: `sudo ufw allow 8080`
- Try: http://127.0.0.1:8080
- Check server is running

### Port Already in Use
```bash
# Use different port
./meshradio --gui --port 9000
```

### No Audio
- This is normal! (simulated audio for now)
- Install PortAudio + Opus for real audio
- See: AUDIO_SETUP.md

## Next Steps

- See **AUDIO_SETUP.md** for real audio
- See **GUI_README.md** for GUI details
- See **DESIGN.md** for architecture
- See **QUICKSTART.md** for quick start

---

**Choose your interface and start broadcasting!** 📻
