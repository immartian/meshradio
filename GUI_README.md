# MeshRadio Web GUI

## 🎨 Beautiful Web Interface for MeshRadio

MeshRadio now includes a gorgeous web-based graphical user interface! Control your mesh radio station from your browser.

## Quick Start

### Run the GUI

```bash
# From the repository
/tmp/meshradio-gui

# Or build and run
make build-gui
./meshradio-gui

# Or use make run-gui
make run-gui
```

Then open your browser to: **http://localhost:8080**

## What You Get

### 🌟 Features

1. **Beautiful Modern UI**
   - Gradient backgrounds
   - Glassmorphism design
   - Smooth animations
   - Responsive layout

2. **Real-Time Status**
   - WebSocket updates every second
   - Live packet counters
   - Signal strength meters
   - Audio level visualization

3. **Easy Controls**
   - One-click broadcasting
   - Simple station tuning
   - Status at a glance
   - Activity log

4. **Live Monitoring**
   - Animated status indicators
   - Signal quality bars
   - Packet statistics
   - Connection status

## Screenshots

### Main Interface
```
┌─────────────────────────────────────────────────┐
│        📻 MeshRadio                             │
│   Decentralized Broadcasting over Yggdrasil    │
├─────────────────────────────────────────────────┤
│ Callsign: Martian                               │
│ IPv6: 201:e8c5:3538:87a3:aa54:7dfb:8008:fb2e   │
│ Mode: [● Broadcasting] Network: [● Connected]   │
├─────────────────────────────────────────────────┤
│                                                 │
│  🎙️ Broadcast          🎧 Listen              │
│  ┌─────────────┐        ┌─────────────┐        │
│  │ ● LIVE      │        │ Station:    │        │
│  │ Audio: ▓▓▓▓ │        │ W2XYZ       │        │
│  │ [Stop]      │        │ Signal: ▓▓▓ │        │
│  └─────────────┘        └─────────────┘        │
│                                                 │
│  📋 Recent Activity                            │
│  [14:30] Broadcasting started                   │
│  [14:29] Connected to server                    │
└─────────────────────────────────────────────────┘
```

## How to Use

### 1. Start Broadcasting

1. Click **"Start Broadcasting"** button
2. Watch the status change to **● BROADCASTING**
3. See audio meters animating
4. Monitor transmission in activity log

### 2. Listen to a Station

1. Enter station's IPv6 address
2. Click **"Start Listening"**
3. Watch signal strength meter
4. See packet count increasing

### 3. Monitor Status

- **Top bar** shows your callsign and IPv6
- **Network indicator** shows connection status
- **Mode badge** shows current operation
- **Activity log** shows recent events

## Technical Details

### Architecture

```
Browser (Port 8080)
    ↕ HTTP + WebSocket
Web GUI Server (Go)
    ↕
Broadcaster/Listener
    ↕
Yggdrasil Network
```

### API Endpoints

- `GET /` - Web interface
- `GET /ws` - WebSocket for real-time updates
- `POST /api/broadcast/start` - Start broadcasting
- `POST /api/broadcast/stop` - Stop broadcasting
- `POST /api/listen/start` - Start listening
- `POST /api/listen/stop` - Stop listening
- `GET /api/status` - Get current status

### WebSocket Updates

Status updates sent every second:
```json
{
  "timestamp": 1732557845,
  "callsign": "Martian",
  "ipv6": "201:e8c5:3538:87a3:aa54:7dfb:8008:fb2e",
  "mode": "broadcasting",
  "packetCount": 1234,
  "signalQuality": 245
}
```

## Features

### Live Status Dashboard
- Real-time updates via WebSocket
- No page refresh needed
- Instant status changes

### Visual Feedback
- Animated pulse on broadcasting
- Signal strength bars
- Audio level meters
- Connection indicators

### Activity Log
- Timestamps for all events
- Color-coded messages (success/error/info)
- Scrollable history
- Auto-updates

### Responsive Design
- Works on desktop
- Works on tablet
- Works on mobile
- Adapts to screen size

## Comparison: TUI vs GUI

### Terminal UI (meshradio)
- ✅ Lightweight
- ✅ SSH-friendly
- ✅ No browser needed
- ✅ Keyboard shortcuts
- ❌ Text-only interface

### Web GUI (meshradio-gui)
- ✅ Beautiful visuals
- ✅ Mouse-friendly
- ✅ Modern design
- ✅ Easy for beginners
- ❌ Requires browser

**Use both!** They're complementary.

## Customization

### Change Port

```bash
# Edit cmd/meshradio-gui/main.go
server := gui.NewServer(8080, callsign, localIPv6)
// Change 8080 to your preferred port
```

### Modify UI Colors

Edit `pkg/gui/web/style.css`:
```css
:root {
    --primary: #7c3aed;  /* Change this */
    --secondary: #06b6d4;  /* And this */
}
```

## Troubleshooting

### Port Already in Use

```bash
# Check what's using port 8080
lsof -i :8080

# Use a different port or kill the process
```

### Can't Connect to GUI

1. Check firewall settings
2. Ensure meshradio-gui is running
3. Try http://127.0.0.1:8080 instead
4. Check browser console for errors

### WebSocket Disconnects

- Check network stability
- Ensure server is still running
- Refresh the page
- Check browser console

## Development

### File Structure

```
pkg/gui/
├── server.go           # HTTP + WebSocket server
└── web/
    ├── index.html      # Main page
    ├── style.css       # Beautiful CSS
    └── app.js          # WebSocket client

cmd/meshradio-gui/
└── main.go            # Entry point
```

### Adding Features

1. Add API endpoint in `server.go`
2. Add UI controls in `index.html`
3. Add styling in `style.css`
4. Add logic in `app.js`

### Building

```bash
# Standard build
go build -o meshradio-gui ./cmd/meshradio-gui

# With audio support
go build -tags "portaudio opus" -o meshradio-gui ./cmd/meshradio-gui
```

## Future Enhancements

Planned features:
- [ ] Audio device selection
- [ ] Volume controls
- [ ] Spectrum analyzer visualization
- [ ] Station bookmarks in UI
- [ ] Dark/light theme toggle
- [ ] Mobile-optimized layout
- [ ] Multi-language support
- [ ] Export activity log

## Credits

Built with:
- **Go** - Backend server
- **Gorilla WebSocket** - Real-time communication
- **Vanilla JavaScript** - No framework bloat
- **CSS3** - Modern styling
- **HTML5** - Semantic markup

## License

GPL-3.0 (same as MeshRadio)

---

**Enjoy your beautiful MeshRadio GUI!** 🎨📻

Run it: `/tmp/meshradio-gui`

Then browse to: http://localhost:8080
