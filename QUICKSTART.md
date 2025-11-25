# MeshRadio Quick Start Guide

## 🚀 Run the Web GUI

```bash
# Navigate to project
cd /media/im3/plus/labx/meshradio

# Run with your callsign
/tmp/meshradio --gui YourCallsign

# Open in browser
# → http://localhost:7999
```

## 📻 Basic Usage

### Broadcasting
1. Open http://localhost:7999
2. Click **"Start Broadcasting"**
3. Your station is now LIVE on the mesh!
4. Share your IPv6 with listeners

### Listening
1. Get broadcaster's IPv6 address
2. Open http://localhost:7999
3. Enter IPv6 in the "Station IPv6" field
4. Click **"Start Listening"**

## 🎛️ Command Options

```bash
# Web GUI with callsign
./meshradio --gui Martian

# Terminal UI (default)
./meshradio Martian

# Custom port
./meshradio --gui --port 8000 Martian

# Using environment variable
MESHRADIO_CALLSIGN=Martian ./meshradio --gui

# Using explicit flag
./meshradio --gui --callsign Martian
```

## 🔧 Port Numbers

All ports use Yggdrasil theme ("799"):

- **7999** - Web GUI
- **8799** - Broadcaster
- **9799** - Listener

## 📋 Features

### ✅ Working Now
- Web GUI with real-time updates
- Broadcasting to mesh network
- Listening to remote stations
- Activity logging
- Status visualization
- Both TUI and GUI modes

### ⏳ Coming Soon
- Real audio (currently simulated)
- Station scanning
- Audio device selection
- Recording functionality

## 🐛 Troubleshooting

### Port Already in Use
Change the GUI port:
```bash
./meshradio --gui --port 8000 Martian
```

### Can't Execute Binary
Copy to /tmp/:
```bash
cp meshradio /tmp/
/tmp/meshradio --gui Martian
```

### Yggdrasil Not Detected
Check if Yggdrasil is running:
```bash
yggdrasilctl getSelf
```

Install if needed:
```bash
# Ubuntu/Debian
sudo apt-get install yggdrasil

# Or follow: https://yggdrasil-network.github.io/installation.html
```

## 📚 More Documentation

- `DESIGN.md` - System architecture
- `STATUS.md` - Current implementation status
- `COMPLETION_SUMMARY.md` - Full feature documentation
- `AUDIO_SETUP.md` - Real audio integration guide

## 🎯 Next Steps

1. **Test Broadcasting**
   - Start broadcaster on one machine
   - Start listener on another
   - Verify packets are received

2. **Explore the GUI**
   - Try all buttons
   - Watch activity log
   - Monitor meters

3. **Add Real Audio** (Advanced)
   - See `AUDIO_SETUP.md`
   - Install PortAudio and Opus
   - Build with audio support

## 🔗 Links

- **Repository**: https://github.com/immartian/meshradio
- **Yggdrasil**: https://yggdrasil-network.github.io/
- **Issues**: https://github.com/immartian/meshradio/issues

---

**Ready to broadcast on the mesh!** 📻

*Version: 0.3-alpha*
