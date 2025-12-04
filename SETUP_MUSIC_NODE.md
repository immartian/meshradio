# Setting Up a Music Broadcasting Node

Choose one machine to be your music DJ station!

## Step-by-Step Setup

### On the Music Broadcasting Node

#### 1. Make sure you have music files

```bash
# Check if you have MP3 files
ls ~/Music/*.mp3
ls -R ~/Music/*.mp3

# Or check your music directory
find ~/Music -name "*.mp3" | wc -l
```

If you don't have music in ~/Music, note where your music is located.

#### 2. Build MeshRadio

```bash
cd meshradio
./build.sh
```

#### 3. Get your IPv6 address (for others to connect)

```bash
yggdrasilctl getSelf | grep "IPv6 address"
```

**Example output:**
```
IPv6 address: 200:1234:5678::abcd
```

**Write this down!** Other nodes will use this to listen.

#### 4. Scan your music library

```bash
# Default: scans ~/Music
./music-broadcast

# Or specify your music directory
./music-broadcast --dir /path/to/your/music

# Example with custom directory
./music-broadcast --dir /media/external/MusicCollection
```

**What you'll see:**
```
╔══════════════════════════════════════════════════════════════╗
║ MeshRadio Music Broadcaster                                  ║
╠══════════════════════════════════════════════════════════════╣
║ Callsign:    MUSIC-DJ                                        ║
║ Channel:     talk                                            ║
║ Port:        8798                                            ║
║ Music Dir:   /home/user/Music                                ║
╚══════════════════════════════════════════════════════════════╝

🔍 Scanning for MP3 files...
✅ Found 247 MP3 file(s)

📻 Playlist:
  1. Artist1 - Song1.mp3 (Duration: 3m45s | 44100 Hz)
  2. Artist1 - Song2.mp3 (Duration: 4m12s | 44100 Hz)
  3. Artist2 - Song3.mp3 (Duration: 2m58s | 44100 Hz)
  ...

📡 Broadcasting on: 200:1234:5678::abcd:8798
```

#### 5. Customize your broadcast (optional)

```bash
# Custom callsign
./music-broadcast --callsign DJ-ROCK

# Custom channel (for different groups)
./music-broadcast --group community --port 8795

# Shuffle playlist
./music-broadcast --shuffle

# Don't loop (play once and stop)
./music-broadcast --loop=false

# Combine options
./music-broadcast \
  --callsign DJ-JAZZ \
  --dir ~/Music/Jazz \
  --group community \
  --port 8795 \
  --shuffle
```

---

### On the Listening Node(s)

Other machines can listen to your music broadcast:

#### Option 1: Quick Listen (Command Line)

```bash
# Replace with broadcasting node's IPv6
./emergency-test listen-manual \
  -target 200:1234:5678::abcd \
  -port 8798 \
  -group talk
```

#### Option 2: Use Main Program (TUI)

```bash
./meshradio

# In the TUI:
# 1. Press 'l' to listen
# 2. Enter target IPv6: 200:1234:5678::abcd
# 3. Enter port: 8798
# 4. Enjoy the music!
```

#### Option 3: Use Web GUI

```bash
./meshradio --gui --port 8080

# Open browser: http://localhost:8080
# Click "Listen"
# Enter the broadcaster's IPv6 and port
```

#### Option 4: Auto-Discovery (If mDNS works)

```bash
# Browse for available stations
./mdns-test browse

# Should show something like:
# 🎵 DJ-ROCK._meshradio._udp.local.
#    IPv6: 200:1234:5678::abcd
#    Port: 8798
#    Group: talk
```

---

## Complete Example: Two-Node Music Station

### Music Node Setup

**Machine: `dj-server` (has music library)**

```bash
# 1. Build
cd meshradio
./build.sh

# 2. Get IPv6
yggdrasilctl getSelf | grep "IPv6 address"
# Output: 200:1234:5678::abcd

# 3. Scan and prepare music
./music-broadcast \
  --callsign DJ-STATION \
  --dir ~/Music/Favorites \
  --group talk \
  --shuffle

# Output shows:
# ✅ Found 150 MP3 file(s)
# 📡 Broadcasting on: 200:1234:5678::abcd:8798
```

### Listener Node Setup

**Machine: `listener-client`**

```bash
# 1. Build
cd meshradio
./build.sh

# 2. Listen to DJ-STATION
./emergency-test listen-manual \
  -target 200:1234:5678::abcd \
  -port 8798 \
  -group talk

# Should see:
# Connected to station: DJ-STATION
# Received: packets=50, seq=49, from=DJ-STATION
```

---

## Current Status & Workarounds

### ⚠️ Current Limitation

The music-broadcast tool currently:
- ✅ **Scans** your music files
- ✅ **Lists** what it found
- ✅ **Shows** metadata (duration, sample rate)
- ⏳ **Does NOT yet** actually broadcast the audio

**Full MP3 playback integration is coming in Phase 2.**

### Workaround: Broadcast Music NOW

Until Phase 2 is complete, here's how to broadcast music right now:

#### Method 1: Microphone Capture (Simple)

**On Music Node:**
```bash
# Terminal 1: Play music locally
mpg123 ~/Music/*.mp3

# Terminal 2: Broadcast (microphone picks up speakers)
./meshradio
# Press 'b' to broadcast
# Your mic captures the speaker output
```

**On Listener Node:**
```bash
./meshradio
# Press 'l' to listen
# Enter DJ node's IPv6 and port
```

#### Method 2: Virtual Audio Cable (Better Quality)

**On Music Node (Linux with PulseAudio):**

```bash
# Setup virtual audio device
pactl load-module module-null-sink sink_name=music_sink sink_properties=device.description="MusicSink"
pactl load-module module-loopback sink=music_sink

# Play music to virtual sink
mpg123 --audiodevice music_sink ~/Music/*.mp3 &

# Broadcast from virtual sink
./meshradio
# Press 'b' to broadcast
# Select "MusicSink" as input device
```

**On Listener Node:**
```bash
./meshradio
# Press 'l' to listen
# Enter DJ node's IPv6 and port
```

---

## Recommended Channel Settings

Choose the right channel for your broadcast:

| Use Case | Channel | Port | Command |
|----------|---------|------|---------|
| Music Radio | talk | 8798 | `--group talk --port 8798` |
| Community Radio | community | 8795 | `--group community --port 8795` |
| Testing | test | 8799 | `--group test --port 8799` |

**Example:**
```bash
# Community music station
./music-broadcast \
  --callsign COMMUNITY-FM \
  --group community \
  --port 8795 \
  --dir ~/Music/PublicDomain
```

---

## Tips for Organizing Music

### Create Playlists with Directories

```bash
~/Music/
├── Stations/
│   ├── Jazz/           # Jazz station
│   ├── Rock/           # Rock station
│   ├── Classical/      # Classical station
│   └── ChillVibes/     # Chill station
└── Emergency/
    └── alerts/         # Emergency announcements
```

**Broadcast specific playlist:**
```bash
# Jazz station
./music-broadcast --dir ~/Music/Stations/Jazz --callsign DJ-JAZZ

# Rock station
./music-broadcast --dir ~/Music/Stations/Rock --callsign DJ-ROCK
```

### Run Multiple Stations

You can run multiple broadcasters on different ports:

```bash
# Terminal 1: Jazz station
./music-broadcast --dir ~/Music/Jazz --port 8795 --callsign DJ-JAZZ

# Terminal 2: Rock station
./music-broadcast --dir ~/Music/Rock --port 8796 --callsign DJ-ROCK

# Terminal 3: Chill station
./music-broadcast --dir ~/Music/Chill --port 8797 --callsign DJ-CHILL
```

Listeners can choose which station to tune into!

---

## Troubleshooting

### "No MP3 files found!"

**Check your directory:**
```bash
# See if MP3 files exist
ls -R ~/Music/*.mp3

# Count MP3 files
find ~/Music -name "*.mp3" | wc -l
```

**Solutions:**
- Make sure files have `.mp3` extension (case-sensitive)
- Check file permissions: `chmod -R +r ~/Music`
- Use `--dir` to specify correct path
- Try absolute path: `--dir /home/username/Music`

### "Could not get Yggdrasil IPv6"

```bash
# Check Yggdrasil is running
sudo systemctl status yggdrasil

# Start if needed
sudo systemctl start yggdrasil

# Get IPv6
yggdrasilctl getSelf
```

### Listener can't connect

1. **Get correct IPv6 from broadcaster:**
   ```bash
   # On broadcasting node
   yggdrasilctl getSelf | grep "IPv6 address"
   ```

2. **Test with localhost first:**
   ```bash
   # If both broadcaster and listener on same machine
   # Use ::1 as target
   ./emergency-test listen-manual -target ::1 -port 8798 -group talk
   ```

3. **Check firewall:**
   ```bash
   sudo ufw allow 8790:8799/udp
   ```

4. **Verify Yggdrasil connectivity:**
   ```bash
   # On listener node, ping broadcaster
   ping6 200:1234:5678::abcd
   ```

---

## Advanced: Automated Music Station Script

Create a script to start your music station:

```bash
#!/bin/bash
# start-music-station.sh

# Get IPv6
IPV6=$(yggdrasilctl getSelf | grep "IPv6 address" | awk '{print $3}')

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║ Starting Music Station                                       ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo
echo "Station: DJ-ROCK"
echo "Channel: community"
echo "Port: 8795"
echo "IPv6: $IPV6"
echo
echo "Listeners connect to: $IPV6:8795"
echo

# Scan music
./music-broadcast \
  --callsign DJ-ROCK \
  --dir ~/Music/Rock \
  --group community \
  --port 8795 \
  --shuffle \
  --loop
```

Make it executable:
```bash
chmod +x start-music-station.sh
./start-music-station.sh
```

---

## What's Next?

Once Phase 2 (full MP3 playback) is implemented, the music-broadcast program will:

- ✅ Actually decode and play MP3 files
- ✅ Broadcast the audio over the mesh
- ✅ Support playback controls (next, prev, pause)
- ✅ Show now-playing metadata
- ✅ Support multiple audio formats (FLAC, OGG, etc.)

For now, use the workarounds above to broadcast music!

---

## Quick Reference

```bash
# Scan music (default ~/Music)
./music-broadcast

# Scan custom directory
./music-broadcast --dir /path/to/music

# Custom callsign and channel
./music-broadcast --callsign DJ-ROCK --group talk

# Get your IPv6 (share with listeners)
yggdrasilctl getSelf | grep "IPv6 address"

# Listen to music broadcast
./emergency-test listen-manual -target <ipv6> -port 8798 -group talk

# Or use TUI
./meshradio
# Press 'l', enter IPv6 and port
```

---

**Ready to broadcast? Pick your music node, scan your library, and let others tune in!** 🎵📡
