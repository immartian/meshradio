# MeshRadio Implementation TODO

**Version:** v2.0-standards (SSM/RTP/mDNS based)
**Last Updated:** 2025-11-30
**Goal:** Build emergency mesh radio on proven standards

---

## Layer 1: Transport Layer (Yggdrasil IPv6)

**Status:** ✅ Already works (Yggdrasil handles this)

**What we have:**
- [x] Yggdrasil IPv6 connectivity
- [x] Encrypted unicast routing
- [x] NAT traversal (Yggdrasil handles)

**What we need:**
- [ ] Verify Yggdrasil performance for real-time audio
- [ ] Test latency/jitter across multi-hop routes
- [ ] Measure bandwidth capacity per connection

**Dependencies:** None (Yggdrasil core)

---

## Layer 2: Streaming Layer (RTP)

**Status:** 🚧 To implement

### 2.1 RTP Library Integration

**Goal:** Use standard RTP for audio packets

- [ ] Choose RTP library
  - Option A: `github.com/pion/rtp` (pure Go, good docs)
  - Option B: `github.com/pion/webrtc/v3` (includes RTP + RTCP)
  - Option C: GStreamer bindings (heavy, full-featured)
  - **Decision:** _____________________

- [ ] Basic RTP sender
  - [ ] Create RTP session
  - [ ] Set payload type (111 = Opus)
  - [ ] Generate sequence numbers
  - [ ] Add timestamps
  - [ ] Send packets via UDP

- [ ] Basic RTP receiver
  - [ ] Receive RTP packets
  - [ ] Parse header
  - [ ] Extract Opus payload
  - [ ] Detect packet loss (sequence gaps)
  - [ ] Handle jitter

- [ ] RTCP support
  - [ ] Sender reports
  - [ ] Receiver reports
  - [ ] Statistics tracking

**Test:** Send RTP stream locally, receive with VLC/ffmpeg

### 2.2 Opus Codec Integration

**Goal:** Encode/decode audio with Opus

- [ ] Choose Opus library
  - Option A: `github.com/hrfee/gopus` (current, working)
  - Option B: `gopkg.in/hraban/opus.v2`
  - **Decision:** Keep current (gopus)

- [ ] Opus encoder
  - [ ] Configure: 48kHz, mono/stereo
  - [ ] Bitrate: 64kbps (emergency), 128kbps (regular)
  - [ ] Frame size: 20ms
  - [ ] Complexity vs. quality tradeoff

- [ ] Opus decoder
  - [ ] Decode frames
  - [ ] Handle packet loss concealment (PLC)
  - [ ] FEC (Forward Error Correction) support

**Test:** Encode → RTP → Decode, measure quality

### 2.3 Audio Backend

**Goal:** Capture/playback real audio

- [ ] Keep PortAudio integration (desktop)
  - [x] Already working in v0.4
  - [ ] Verify latency settings
  - [ ] Buffer size tuning

- [ ] Add mobile audio backends
  - [ ] Android: AAudio/OpenSL ES
  - [ ] iOS: CoreAudio/AVAudioEngine
  - **Defer to Phase 2 (mobile)**

**Test:** Microphone → Opus → RTP → Speaker, measure latency

---

## Layer 3: Discovery Layer (mDNS/Avahi)

**Status:** 🚧 To implement

### 3.1 mDNS Library Integration

**Goal:** Advertise and discover stations

- [ ] Choose mDNS library
  - Option A: `github.com/grandcat/zeroconf` (cross-platform, pure Go)
  - Option B: `github.com/hashicorp/mdns` (pure Go, simpler)
  - Option C: `github.com/holoplot/go-avahi` (Linux only, native Avahi)
  - **Decision:** Start with zeroconf (cross-platform)

- [ ] Service advertisement (broadcaster)
  - [ ] Define service type: `_meshradio._udp.local.`
  - [ ] Advertise when broadcast starts
  - [ ] Withdraw when broadcast stops
  - [ ] Set TXT records (see below)

- [ ] Service discovery (listener)
  - [ ] Browse for `_meshradio._udp` services
  - [ ] Parse TXT records
  - [ ] Track discovered stations
  - [ ] Handle service removal

**Test:** Advertise on machine A, discover on machine B (same LAN)

### 3.2 TXT Record Convention

**Goal:** Define standard metadata format

- [ ] Define required TXT fields
  - [ ] `group` - Multicast group label (e.g., "emergency")
  - [ ] `channel` - Channel type (emergency/community/talk)
  - [ ] `callsign` - Station identifier
  - [ ] `port` - RTP port (8790-8799)

- [ ] Define optional TXT fields
  - [ ] `priority` - normal/high/emergency/critical
  - [ ] `codec` - opus/flac
  - [ ] `bitrate` - kbps
  - [ ] `source` - Broadcaster IPv6 (for SSM)
  - [ ] `description` - Human-readable

- [ ] Example TXT record:
  ```
  _meshradio._udp.local.
    group=emergency
    channel=emergency
    callsign=W1EMERGENCY
    port=8790
    priority=critical
    codec=opus
    bitrate=64
  ```

- [ ] Document in specification

**Test:** Advertise with TXT, query and parse correctly

### 3.3 Emergency Fallback (No mDNS)

**Goal:** Work when mDNS fails

- [ ] Manual station entry
  - [ ] Config file: `stations.yaml`
  - [ ] CLI: `meshradio tune 203::1:8790`
  - [ ] UI: Direct IPv6:port input

- [ ] Emergency port convention
  - [ ] Document port range: 8790-8799
  - [ ] 8790 = General emergency
  - [ ] 8791 = Net control
  - [ ] 8792 = Medical
  - [ ] etc.

- [ ] Quick connect shortcuts
  - [ ] `meshradio emergency` → auto-tunes to 8790
  - [ ] `meshradio netcontrol` → auto-tunes to 8791

**Test:** Connect without mDNS discovery

---

## Layer 4: Multicast Overlay (SSM/Regular Multicast)

**Status:** 🚧 To implement

### 4.1 Regular Multicast (Emergency Channels)

**Goal:** Any-source multicast for emergency

- [ ] Group management
  - [ ] Track subscribers per group (e.g., "emergency")
  - [ ] Multiple broadcasters can send to same group
  - [ ] Listeners receive from ALL sources in group

- [ ] JOIN handling
  - [ ] Listener sends JOIN("emergency")
  - [ ] Discover all broadcasters for this group (mDNS)
  - [ ] Subscribe to each broadcaster (unicast)

- [ ] Broadcaster registration
  - [ ] Broadcaster advertises group membership
  - [ ] Accepts JOIN requests
  - [ ] Sends RTP to all subscribers (unicast fan-out)

**Test:** 2 broadcasters, 3 listeners, same emergency group

### 4.2 SSM (Source-Specific Multicast) for Regular Channels

**Goal:** Subscribe to specific broadcaster

- [ ] SSM subscription
  - [ ] Listener specifies (Source, Group)
  - [ ] Example: (203:abcd::1, "talk")
  - [ ] Only receives from that specific source

- [ ] SSM JOIN handling
  - [ ] Send JOIN(source, group) → broadcaster
  - [ ] Broadcaster adds to subscriber list
  - [ ] Sends RTP unicast to subscriber

- [ ] SSM LEAVE handling
  - [ ] Explicit LEAVE message
  - [ ] Or timeout (30s no heartbeat)
  - [ ] Broadcaster removes from list

**Test:** Listener subscribes to specific broadcaster, not others

### 4.3 Subscription Protocol

**Goal:** Manage subscriptions over unicast

- [ ] Protocol packets (extend existing)
  - [ ] SUBSCRIBE (source, group, listener_ipv6, port)
  - [ ] UNSUBSCRIBE
  - [ ] HEARTBEAT (keepalive)
  - [ ] Already have these from v0.4!

- [ ] Subscriber tracking
  - [ ] Map: group → []subscribers
  - [ ] Update LastSeen on heartbeat
  - [ ] Prune stale (15s timeout)

- [ ] Fan-out logic
  - [ ] For each RTP packet
  - [ ] Send to all subscribers in group
  - [ ] Unicast to each subscriber's IPv6:port

**Test:** Subscribe, receive RTP, heartbeat, timeout works

---

## Layer 5: Application Layer (Emergency Features)

**Status:** 🚧 To implement

### 5.1 Emergency Channels

**Goal:** Pre-defined emergency channels

- [ ] Define channel registry
  ```yaml
  emergency:
    group: "emergency"
    port: 8790
    priority: critical
    auto_tune: true

  netcontrol:
    group: "netcontrol"
    port: 8791
    priority: emergency
    auto_tune: false

  medical:
    group: "medical"
    port: 8792
    priority: emergency
    auto_tune: prompt
  ```

- [ ] Load from config
- [ ] UI shortcuts (e.g., "Emergency" button)

**Test:** UI can quick-tune to emergency channels

### 5.2 Priority Signaling

**Goal:** Emergency broadcasts interrupt normal traffic

- [ ] RTCP APP extension for priority
  - [ ] Custom RTCP packet type
  - [ ] Priority field: 0=normal, 1=high, 2=emergency, 3=critical
  - [ ] Emergency message text

- [ ] Priority handling (receiver)
  - [ ] Detect high-priority RTCP
  - [ ] Log event
  - [ ] Notify UI
  - [ ] If critical: prompt user to switch

- [ ] Auto-tune behavior
  - [ ] User preference: always/prompt/never
  - [ ] If "always": auto-switch to emergency
  - [ ] If "prompt": show notification
  - [ ] If "never": just log

**Test:** Emergency broadcast triggers auto-tune

### 5.3 Net Control Protocol

**Goal:** Coordinate emergency response

- [ ] Net control messages (via RTCP APP)
  - [ ] "check-in" - request status
  - [ ] "stand-by" - wait for instructions
  - [ ] "report" - send status update
  - [ ] "all-clear" - emergency resolved

- [ ] Net control UI
  - [ ] Show participant list
  - [ ] Track check-ins
  - [ ] Display status reports

- [ ] Participant behavior
  - [ ] Respond to net control commands
  - [ ] Send status when requested

**Test:** Simulate emergency net with 5+ participants

### 5.4 Station Metadata

**Goal:** Track station info

- [ ] Station database (SQLite)
  - [ ] IPv6, callsign, group, port
  - [ ] Last seen timestamp
  - [ ] Metadata (description, location, etc.)
  - [ ] Reputation/trust score

- [ ] Persistence
  - [ ] Save discovered stations
  - [ ] Load on startup
  - [ ] Age out stale entries (7 days)

- [ ] Bookmarks
  - [ ] User can save favorite stations
  - [ ] Quick access in UI

**Test:** Discover station, persists across restart

---

## Layer 6: Service Layer (API)

**Status:** 🚧 To implement

### 6.1 Core Services

**Goal:** Clean service abstractions

- [ ] StationService
  - [ ] StartBroadcast(config)
  - [ ] StopBroadcast()
  - [ ] TuneTo(source, group, port)
  - [ ] Disconnect()
  - [ ] GetListeners() []Listener

- [ ] DiscoveryService
  - [ ] Advertise(metadata)
  - [ ] Browse() []Station
  - [ ] FindStation(callsign) Station
  - [ ] Subscribe(callback)

- [ ] CoordinationService (emergency)
  - [ ] DeclareEmergency(channel)
  - [ ] ClearEmergency(channel)
  - [ ] GetEmergencies() []Emergency

- [ ] StorageService
  - [ ] SaveStation(info)
  - [ ] GetStations() []Station
  - [ ] SaveBookmark(bookmark)
  - [ ] GetBookmarks() []Bookmark

**Test:** Unit tests for each service

### 6.2 Daemon Mode

**Goal:** Background service with API

- [ ] gRPC service definition
  - [ ] Define .proto file
  - [ ] Generate Go code
  - [ ] Implement server

- [ ] Daemon process
  - [ ] Run as background service
  - [ ] Expose gRPC on localhost:7998
  - [ ] Handle signals (graceful shutdown)

- [ ] CLI client
  - [ ] Connect to daemon gRPC
  - [ ] Control commands (start, stop, tune)
  - [ ] Status queries

**Test:** Start daemon, control via CLI

### 6.3 REST API (Alternative)

**Goal:** HTTP API for web/mobile

- [ ] REST endpoints
  - [ ] POST /broadcast/start
  - [ ] POST /broadcast/stop
  - [ ] POST /listen/tune
  - [ ] GET /stations
  - [ ] GET /emergency/status

- [ ] Server-Sent Events (SSE)
  - [ ] GET /events (stream)
  - [ ] Push station updates
  - [ ] Push emergency alerts

**Test:** curl commands work, SSE streams events

---

## Layer 7: UI Layer

**Status:** ⏳ Refactor existing

### 7.1 Terminal UI (TUI)

**Goal:** Update existing TUI for v2

- [ ] Update to use service layer
  - [ ] Remove direct broadcaster/listener calls
  - [ ] Call StationService methods
  - [ ] Use DiscoveryService for stations

- [ ] Add emergency features
  - [ ] Emergency button (quick-tune to 8790)
  - [ ] Priority alert banner
  - [ ] Net control view

- [ ] Show discovered stations
  - [ ] List from DiscoveryService
  - [ ] Filter by channel type
  - [ ] Real-time updates

**Test:** TUI works with new architecture

### 7.2 Web GUI

**Goal:** Update existing web GUI

- [ ] Connect to daemon API
  - [ ] Use gRPC-web or REST
  - [ ] Real-time updates via SSE

- [ ] Emergency UI
  - [ ] Big red "EMERGENCY" button
  - [ ] Alert notifications
  - [ ] Auto-tune prompt

- [ ] Station browser
  - [ ] Table view with filters
  - [ ] Click to tune
  - [ ] Bookmark stations

**Test:** Web GUI controls daemon

### 7.3 Mobile App (Phase 2)

**Goal:** Integrate with Yggdrasil mobile apps

- [ ] C library (libmeshradio)
  - [ ] Export C API
  - [ ] Build .so/.dylib
  - [ ] FFI bindings

- [ ] Android module
  - [ ] Integrate with Yggdrasil Android app
  - [ ] Use Android NSD (mDNS)
  - [ ] Background service
  - [ ] Notifications

- [ ] iOS module
  - [ ] Integrate with Yggdrasil iOS app
  - [ ] Use Bonjour
  - [ ] Background audio
  - [ ] Notifications

**Defer to Phase 2**

---

## Testing & Validation

### Unit Tests
- [ ] Protocol encoding/decoding
- [ ] Service layer methods
- [ ] Multicast group management
- [ ] Priority logic

### Integration Tests
- [ ] RTP end-to-end (local)
- [ ] mDNS discovery (local network)
- [ ] Multi-listener scenario
- [ ] Emergency channel switching

### System Tests
- [ ] Real Yggdrasil mesh (2+ machines)
- [ ] Latency measurement
- [ ] Bandwidth usage
- [ ] Network partition recovery

### Emergency Drills
- [ ] Simulated emergency scenario
- [ ] 10+ participants
- [ ] Net control coordination
- [ ] Measure response time

---

## Documentation

### Technical Docs
- [ ] Protocol specification
  - [ ] mDNS TXT record format
  - [ ] RTP/RTCP usage
  - [ ] Multicast overlay design
  - [ ] Emergency extensions

- [ ] API documentation
  - [ ] gRPC service definition
  - [ ] REST endpoints
  - [ ] Library API (for mobile)

- [ ] Architecture diagrams
  - [ ] Layer diagram
  - [ ] Message flow
  - [ ] Emergency scenario

### User Docs
- [ ] Quick start guide
- [ ] Emergency procedures
- [ ] Channel guide (8790-8799)
- [ ] Troubleshooting

### Operator Docs
- [ ] Daemon deployment
- [ ] Configuration reference
- [ ] Best practices
- [ ] Security considerations

---

## Milestones

### Milestone 1: RTP Streaming Works (2 weeks)
- [ ] RTP library integrated
- [ ] Opus encoding/decoding
- [ ] Can send/receive locally
- [ ] VLC can play stream

### Milestone 2: mDNS Discovery Works (2 weeks)
- [ ] mDNS advertisement
- [ ] Service browsing
- [ ] TXT record parsing
- [ ] Works on local network

### Milestone 3: Multicast Overlay Works (2 weeks)
- [ ] Regular multicast (emergency)
- [ ] SSM (regular channels)
- [ ] Multiple listeners
- [ ] Works on Yggdrasil mesh

### Milestone 4: Emergency Features Work (2 weeks)
- [ ] Emergency channels defined
- [ ] Priority signaling
- [ ] Auto-tune logic
- [ ] Net control basics

### Milestone 5: Production Ready (4 weeks)
- [ ] Daemon mode
- [ ] API documented
- [ ] TUI/GUI updated
- [ ] Real mesh testing
- [ ] Community feedback

### Milestone 6: Mobile Ready (4 weeks)
- [ ] libmeshradio library
- [ ] Android integration
- [ ] iOS integration
- [ ] Beta testing

---

## Open Questions

### For Community (@majestrate, @parnikkapore, @Revertron)

1. **mDNS over Yggdrasil mesh:**
   - Does mDNS work across Yggdrasil segments?
   - Need mDNS relay/proxy?

2. **Emergency multicast groups:**
   - Suggested naming: "emergency", "netcontrol", "medical"?
   - Or use different convention?

3. **Priority RTCP extension:**
   - Standard RTCP APP packet type OK?
   - Or define custom extension?

4. **Emergency port range:**
   - 8790-8799 good choice?
   - Different numbers better?

5. **Auto-tune behavior:**
   - Always/prompt/never for critical priority?
   - Default setting?

---

## Dependencies & Libraries

### Required
- [x] Yggdrasil (already have)
- [x] PortAudio (already have)
- [ ] RTP library (pion/rtp or pion/webrtc)
- [ ] mDNS library (zeroconf or hashicorp/mdns)
- [ ] Opus library (gopus - already have)

### Optional
- [ ] GStreamer (for advanced features)
- [ ] SQLite (for persistence - already have)
- [ ] gRPC (for daemon API)

### Platform-Specific
- [ ] Avahi (Linux mDNS)
- [ ] Bonjour (macOS mDNS)
- [ ] Android NSD (Android mDNS)
- [ ] iOS Bonjour (iOS mDNS)

---

## Notes

- Keep v0.4 subscription protocol as fallback during transition
- Prioritize emergency use case over features
- Test on real Yggdrasil mesh frequently
- Get community feedback early and often
- Document everything for future contributors

---

**Last Updated:** 2025-11-30
**Next Review:** Weekly during implementation
