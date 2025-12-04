# MeshRadio Quick Start Guide

Get MeshRadio running in 5 minutes!

## On Each Node

### 1. Prerequisites (One-time setup)

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y git golang-go portaudio19-dev avahi-daemon

# Install Yggdrasil
wget https://github.com/yggdrasil-network/yggdrasil-go/releases/download/v0.5.5/yggdrasil-0.5.5-amd64.deb
sudo dpkg -i yggdrasil-0.5.5-amd64.deb
sudo systemctl start yggdrasil
```

### 2. Build MeshRadio

```bash
# Clone and build
git clone https://github.com/immartian/meshradio.git
cd meshradio
./build.sh
```

You should see:
```
✅ All binaries built successfully!

Available programs:
  ./meshradio        - Main program (TUI and Web GUI)
  ./rtp-test         - Test RTP streaming
  ./mdns-test        - Test mDNS discovery
  ./multicast-test   - Test multicast overlay
  ./emergency-test   - Test emergency features
  ./music-broadcast  - Broadcast music files (MP3)
```

### 3. Get Your IPv6 Address

```bash
yggdrasilctl getSelf | grep "IPv6 address"
```

You'll see something like: `IPv6 address: 200:1234:5678::abcd`

**Write this down!** You'll need it for the other node.

## Quick Tests

### Test 1: Audio Streaming (Simplest)

**Node 1 (Broadcaster):**
```bash
./rtp-test broadcast -callsign STATION-A -port 8799
```

**Node 2 (Listener):**
```bash
# Use Node 1's IPv6 address
./rtp-test listen -target 200:1234:5678::abcd -port 8799
```

✅ **Success:** You should see "Received packet" messages!

---

### Test 2: Service Discovery (Auto-find stations)

**Node 1 (Advertise):**
```bash
./mdns-test advertise -callsign BEACON-1 -port 8790 -group emergency
```

**Node 2 (Browse):**
```bash
./mdns-test browse
```

✅ **Success:** You should see `BEACON-1._meshradio._udp.local.` appear!

---

### Test 3: Emergency Priority (Cool alerts!)

**Node 1 (Critical Broadcast):**
```bash
./emergency-test broadcast-critical
```

**Node 2 (Listen):**
```bash
# Use Node 1's IPv6 address
./emergency-test listen-manual -target 200:1234:5678::abcd -port 8790 -group emergency
```

✅ **Success:** You should see 🚨 **CRITICAL EMERGENCY BROADCAST** alerts!

---

### Test 4: Music Broadcasting (Your MP3s!)

**Node 1 (Music):**
```bash
# Scans ~/Music for MP3 files
./music-broadcast

# Or specify custom directory
./music-broadcast --dir /path/to/your/music
```

✅ **Success:** You should see a list of MP3 files found!

**Note:** Full playback integration coming soon. For now it shows what would be played.

---

## Full Integration Test

Want to test everything at once? Use the automated test suite:

**Node 1:**
```bash
./test-integration.sh node1
```

**Node 2:**
```bash
./test-integration.sh node2
# When prompted, enter Node 1's IPv6 address
```

This runs 8 comprehensive tests covering all layers:
1. RTP Streaming
2. mDNS Discovery
3. Regular Multicast
4. SSM (Source-Specific Multicast)
5. Normal Priority
6. High Priority (📢)
7. Emergency Priority (⚠️)
8. Critical Priority (🚨)

Takes about 10-15 minutes with interactive pauses.

---

## Using the Main Program (TUI)

**Node 1 (Broadcaster):**
```bash
./meshradio
```

In the TUI:
- Press **`b`** to start broadcasting
- Enter your callsign: `STATION-A`
- Press Enter to use defaults
- Speak into your microphone!

**Node 2 (Listener):**
```bash
./meshradio
```

In the TUI:
- Press **`l`** to listen
- Enter target IPv6: `200:1234:5678::abcd` (Node 1's address)
- Enter port: `8799`
- You should hear Node 1's audio!

---

## Using the Web GUI

**Start Web GUI:**
```bash
./meshradio --gui --port 8080
```

**Access from browser:**
```
http://localhost:8080
```

Or from another machine on Yggdrasil:
```
http://[200:1234:5678::abcd]:8080
```

---

## Troubleshooting

### "Could not get Yggdrasil IPv6"

Check Yggdrasil is running:
```bash
sudo systemctl status yggdrasil
sudo systemctl start yggdrasil
yggdrasilctl getSelf
```

### "No packets received"

1. Check both nodes are on Yggdrasil network
2. Test with localhost first: use `::1` as target
3. Check firewall:
   ```bash
   sudo ufw allow 8790:8799/udp
   ```

### "mDNS services not found"

1. Check Avahi is running:
   ```bash
   sudo systemctl status avahi-daemon
   sudo systemctl start avahi-daemon
   ```

2. Test Avahi directly:
   ```bash
   avahi-browse -a
   ```

### "PortAudio error"

Make sure PortAudio is installed:
```bash
sudo apt-get install portaudio19-dev
```

Then rebuild:
```bash
./build.sh
```

---

## What's Happening Under the Hood?

When you run a broadcaster and listener:

```
[Node 1: Broadcaster]
  Microphone
     ↓
  PortAudio (capture)
     ↓
  Opus Encoder (compress)
     ↓
  RTP Packets (with priority & metadata)
     ↓
  UDP over Yggdrasil IPv6 (encrypted mesh)
     ↓
[Node 2: Listener]
     ↓
  RTP Receiver (with priority detection)
     ↓
  Opus Decoder
     ↓
  PortAudio (playback)
     ↓
  Speakers
```

**Plus:**
- mDNS for automatic station discovery
- Multicast overlay for group broadcasts
- Emergency priority signaling (4 levels)
- Subscription management with heartbeats

---

## Emergency Channels

MeshRadio has pre-defined emergency channels:

| Channel | Port | Priority | Use Case |
|---------|------|----------|----------|
| emergency | 8790 | 🚨 Critical | Active emergency |
| netcontrol | 8791 | ⚠️ Emergency | Coordination |
| medical | 8792 | ⚠️ Emergency | Medical |
| weather | 8793 | 📢 High | Weather alerts |
| sar | 8794 | ⚠️ Emergency | Search & rescue |
| community | 8795 | Normal | Community info |
| talk | 8798 | Normal | General chat |
| test | 8799 | Normal | Testing |

**Broadcast on emergency channel:**
```bash
./emergency-test broadcast-critical
```

**Listen to emergency channel:**
```bash
./emergency-test listen-manual -target <ipv6> -port 8790 -group emergency
```

---

## Next Steps

Once you have basic audio working:

1. **Read TESTING.md** - Comprehensive testing guide
2. **Read MUSIC.md** - Music broadcasting guide
3. **Read BUILD.md** - Advanced build options
4. **Check TODO.md** - See what's coming next

---

## Getting Help

- **GitHub Issues:** https://github.com/immartian/meshradio/issues
- **Documentation:** See *.md files in repository
- **Logs:** Check /tmp/test-*.log for test output

---

## Quick Reference

```bash
# Build everything
./build.sh

# Quick automated tests (single machine)
./test-quick.sh

# Full integration test (two machines)
./test-integration.sh node1  # Terminal 1
./test-integration.sh node2  # Terminal 2

# Main programs
./meshradio                    # TUI
./meshradio --gui --port 8080  # Web GUI

# Test programs
./rtp-test broadcast           # Broadcast audio
./rtp-test listen -target <ipv6> -port <port>  # Listen
./mdns-test advertise          # Advertise service
./mdns-test browse             # Find services
./emergency-test broadcast-critical  # Emergency broadcast
./music-broadcast              # Scan music files

# Get your IPv6
yggdrasilctl getSelf | grep IPv6

# Check services
sudo systemctl status yggdrasil
sudo systemctl status avahi-daemon
```

---

**You're ready to go! Start with Test 1 above and work your way through.** 🚀

Have fun building your mesh radio network! 📡✨
