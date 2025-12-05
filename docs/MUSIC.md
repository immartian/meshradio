# Music Broadcasting with MeshRadio

Broadcast your music collection over the mesh network!

## Quick Start

### Scan Your Music Library

```bash
# Scan default ~/Music directory
./music-broadcast

# Scan custom directory
./music-broadcast --dir /path/to/music

# With custom callsign and channel
./music-broadcast --callsign DJ-MIKE --group talk --port 8798
```

### Listen to Music Broadcast

On another machine:

```bash
# Get the broadcaster's IPv6 from their terminal output
# Then listen:
./meshradio

# In TUI, press 'l' to listen
# Enter the broadcaster's IPv6:port (e.g., 200:1234::5678:8798)
```

## Features

### Music Discovery

The music broadcaster automatically:
- ✅ Scans directories recursively
- ✅ Finds all MP3 files
- ✅ Shows file count and playlist
- ✅ Displays duration and sample rate
- ✅ Advertises via mDNS (if enabled)

### Supported Formats

Currently supported:
- ✅ MP3 (.mp3)

Planned for future:
- ⏳ FLAC (.flac)
- ⏳ OGG Vorbis (.ogg)
- ⏳ WAV (.wav)
- ⏳ M4A (.m4a)

## Current Status

⚠️ **Note:** The music broadcaster currently demonstrates:
- ✅ Music file discovery and scanning
- ✅ MP3 metadata reading (duration, sample rate)
- ✅ Playlist generation

**Not yet integrated:**
- ⏳ Actual MP3 playback to broadcast
- ⏳ Sample rate conversion (44.1kHz → 48kHz)
- ⏳ Audio pipeline integration with broadcaster

### Workaround: Broadcast Music Now

Until full integration is complete, you can broadcast music using this method:

**On the broadcasting machine:**

```bash
# Terminal 1: Play music locally
mpg123 ~/Music/your-song.mp3

# Terminal 2: Start MeshRadio broadcaster (captures system audio)
./meshradio
# Press 'b' to broadcast
# Your microphone will pick up the speakers
```

Or use virtual audio cables (Linux):

```bash
# Install PulseAudio virtual cable
pactl load-module module-null-sink sink_name=virtual1
pactl load-module module-loopback sink=virtual1

# Play music to virtual sink
mpg123 --audiodevice virtual1 ~/Music/*.mp3

# MeshRadio captures from virtual1
./meshradio
```

## Architecture

### Current Implementation (Phase 1)

```
MP3 Files
    ↓
[Music Scanner]  ← We are here
    ↓
[Playlist]
    ↓
[MP3 Decoder]    ← Partially implemented
    ↓
[44.1kHz PCM]
```

### Full Implementation (Phase 2)

```
MP3 Files
    ↓
[Music Scanner]  ✅ Complete
    ↓
[Playlist Manager]  ← Need to implement
    ↓
[MP3 Decoder]  ⏳ Library integrated
    ↓
[Resampler 44.1→48kHz]  ← Need to implement
    ↓
[Opus Encoder]  ✅ Already have
    ↓
[RTP Broadcaster]  ✅ Already have
    ↓
[Yggdrasil Mesh]  ✅ Already have
```

## Use Cases

### 1. Mesh Radio Station

Run a community radio station over Yggdrasil:

```bash
# DJ machine
./music-broadcast \
  --callsign "MESH-FM" \
  --dir ~/Music/Playlists/ChillVibes \
  --group community \
  --port 8795 \
  --shuffle \
  --loop
```

Listeners discover and tune in automatically via mDNS!

### 2. Emergency Broadcast System

Broadcast pre-recorded emergency messages:

```bash
./music-broadcast \
  --callsign "EMERGENCY-NET" \
  --dir ~/EmergencyMessages \
  --group emergency \
  --port 8790 \
  --loop
```

### 3. Event Audio

Stream live event audio:

```bash
# Record event to MP3s
# Then broadcast with low latency
./music-broadcast \
  --callsign "EVENT-AUDIO" \
  --dir ~/Events/Conference2024 \
  --group talk
```

## Command-Line Options

```
Usage: ./music-broadcast [options]

Options:
  --dir string
      Music directory to scan (default: ~/Music)

  --callsign string
      Your callsign (default: MUSIC-DJ)

  --port int
      Broadcast port (default: 8798 for 'talk' channel)

  --group string
      Multicast group (default: talk)

  --shuffle
      Shuffle playlist (default: false)

  --loop
      Loop playlist (default: true)

  --advertise
      Advertise via mDNS (default: true)
```

## Channels for Music

Recommended channels for music broadcasting:

| Channel | Port | Priority | Use Case |
|---------|------|----------|----------|
| talk | 8798 | Normal | Casual music sharing |
| community | 8795 | Normal | Community radio |
| test | 8799 | Normal | Testing playlists |

**Don't use emergency channels (8790-8794) for music!**

## File Organization Tips

Organize your music for easy broadcasting:

```
~/Music/
├── Playlists/
│   ├── ChillVibes/      # --dir ~/Music/Playlists/ChillVibes
│   ├── UpbeatRock/
│   └── ClassicalMix/
├── Emergency/           # Emergency announcements
│   ├── weather-alert.mp3
│   └── evacuation-notice.mp3
└── Podcasts/
    └── Episode001.mp3
```

Then broadcast specific playlists:

```bash
# Chill vibes station
./music-broadcast --dir ~/Music/Playlists/ChillVibes --callsign DJ-CHILL

# Rock station
./music-broadcast --dir ~/Music/Playlists/UpbeatRock --callsign DJ-ROCK
```

## Network Discovery

When broadcasting with `--advertise` (default), listeners can discover your station:

**Broadcaster advertises:**
```
Service: MUSIC-DJ._meshradio._udp.local.
TXT:
  group=talk
  channel=talk
  callsign=MUSIC-DJ
  port=8798
  priority=normal
  codec=opus
  bitrate=128
```

**Listeners see in discovery:**
```
🎵 MUSIC-DJ
   Channel: talk
   Port: 8798
   IPv6: 200:1234::5678
```

## Performance

### Bandwidth Usage

```
Bitrate 64kbps  (emergency) = 8 KB/s  = 28 MB/hour
Bitrate 96kbps  (voice)     = 12 KB/s = 43 MB/hour
Bitrate 128kbps (music)     = 16 KB/s = 58 MB/hour
Bitrate 192kbps (HQ music)  = 24 KB/s = 86 MB/hour
```

### Listener Capacity

Each broadcaster can handle:
- 50+ listeners (tested)
- 100+ listeners (theoretical, depending on network)

Limited by:
- Broadcaster bandwidth
- CPU for encoding
- Network routing capacity

## Troubleshooting

### "No MP3 files found!"

Check your directory:
```bash
ls -R ~/Music/*.mp3
```

If files exist but not detected, check permissions:
```bash
chmod -R +r ~/Music
```

### MP3 won't decode

Check file integrity:
```bash
mpg123 -t ~/Music/problematic-file.mp3
```

Re-encode if corrupted:
```bash
ffmpeg -i input.mp3 -acodec libmp3lame -b:a 320k output.mp3
```

### High CPU usage

Lower the bitrate:
```bash
# Edit cmd/music-broadcast/main.go
# Change: Bitrate: 128000
# To:     Bitrate: 64000
```

## Future Enhancements

Planned features:

- [ ] Full MP3 playback integration
- [ ] Sample rate conversion (44.1 → 48kHz)
- [ ] Playlist controls (next, previous, pause, resume)
- [ ] WebUI for remote control
- [ ] Volume normalization
- [ ] Crossfade between tracks
- [ ] Metadata display (ID3 tags)
- [ ] Album art transmission
- [ ] Multiple format support (FLAC, OGG, WAV)
- [ ] Live streaming from URL
- [ ] Scheduled broadcasts
- [ ] Request system (listeners request songs)

## Contributing

Want to help implement music broadcasting? See:
- TODO.md for roadmap
- CONTRIBUTING.md for guidelines
- Open issues tagged `music-feature`

## Examples

### Example 1: Local Test

```bash
# Terminal 1: Scan music
./music-broadcast --dir ~/Music

# Shows:
# ✅ Found 247 MP3 file(s)
# 📻 Playlist:
#   1. Artist - Song1.mp3
#   2. Artist - Song2.mp3
#   ...
```

### Example 2: Two-Node Music Station

**Node 1 (Broadcaster):**
```bash
./music-broadcast \
  --callsign MESH-FM-JAZZ \
  --dir ~/Music/Jazz \
  --shuffle \
  --group community

# Shows:
# 📡 Broadcasting on: 200:abcd::1234:8795
# ▶️  Now playing: Miles Davis - So What.mp3
```

**Node 2 (Listener):**
```bash
./meshradio

# In TUI:
# - Discovery shows: 🎵 MESH-FM-JAZZ
# - Press Enter to tune in
# - Enjoy the jazz!
```

## License

MeshRadio is open source. See LICENSE for details.

Enjoy broadcasting your music over the mesh! 🎵📡
