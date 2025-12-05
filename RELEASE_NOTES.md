# MeshRadio v0.3-alpha Release Notes

## Stable Music Streaming Achieved

We're excited to announce **production-quality continuous music broadcasting** in MeshRadio v0.3-alpha.

### What's Working

- **Multi-hour stability**: Successfully tested with 93-song playlists streaming continuously without interruption
- **Web GUI**: Browser-based control interface now functional at `http://localhost:8899`
- **High-quality audio**: Opus codec at 128kbps with 12-13x compression ratio
- **Seamless transitions**: Songs change smoothly without losing listeners

### Critical Fixes

1. **Shared Subscription Manager**: Subscribers no longer disappear when songs change. Fixed by sharing the subscription manager across all broadcaster instances.

2. **IPv6 Key Matching**: Heartbeats now reliably update subscriber last-seen timestamps. Fixed by using hex byte representation for consistent map keys instead of string-based IPv6 addresses.

3. **Ticker-Based Packet Pacing**: Eliminated audio jitter by sending exactly one packet every 20ms instead of bursting.

4. **Clean TUI Shutdown**: Terminal interface now quits properly with timeout-based audio device cleanup.

### Quick Start

```bash
# Terminal 1 - Start broadcaster
./music-broadcast --dir ~/Music --callsign DJ-AWESOME

# Terminal 2 - Start listener
./meshradio
# Select "3" for Listener, enter ::1 and port 8799

# Or use Web GUI
./meshradio --gui --callsign YOUR-CALLSIGN
# Open http://localhost:8899
```

### Technical Specs

- Sample Rate: 48kHz stereo
- Bitrate: 128kbps
- Bandwidth: ~130kbps
- Latency: <100ms typical
- Jitter Buffer: 3 seconds
- Heartbeat: Every 5 seconds

### What's Next

- Voice broadcasting (microphone input)
- Station discovery (mDNS/DHT)
- Multi-listener stress testing
- Recording and time-shifting

**Download**: https://github.com/immartian/meshradio

Built for the decentralized web. Powered by Yggdrasil mesh network.
