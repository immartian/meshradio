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

**Status:** ✅ Complete

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

**Status:** ✅ Complete

### 3.1 mDNS Library Integration

**Goal:** Advertise and discover stations

- [x] Choose mDNS library
  - Option A: `github.com/grandcat/zeroconf` (cross-platform, pure Go)
  - Option B: `github.com/hashicorp/mdns` (pure Go, simpler)
  - Option C: `github.com/holoplot/go-avahi` (Linux only, native Avahi)
  - **Decision:** zeroconf (cross-platform, v1.0.0)

- [x] Service advertisement (broadcaster)
  - [x] Define service type: `_meshradio._udp.local.`
  - [x] Advertise when broadcast starts
  - [x] Withdraw when broadcast stops
  - [x] Set TXT records (see below)

- [x] Service discovery (listener)
  - [x] Browse for `_meshradio._udp` services
  - [x] Parse TXT records
  - [x] Track discovered stations
  - [x] Handle service removal

**Test:** ✅ mdns-test program works for both advertise and browse modes

### 3.2 TXT Record Convention

**Goal:** Define standard metadata format

- [x] Define required TXT fields
  - [x] `group` - Multicast group label (e.g., "emergency")
  - [x] `channel` - Channel type (emergency/community/talk)
  - [x] `callsign` - Station identifier
  - [x] `port` - RTP port (8790-8799)

- [x] Define optional TXT fields
  - [x] `priority` - normal/high/emergency/critical
  - [x] `codec` - opus/flac
  - [x] `bitrate` - kbps
  - [ ] `source` - Broadcaster IPv6 (for SSM) - Defer to Layer 4
  - [ ] `description` - Human-readable - Future enhancement

- [x] Example TXT record:
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

- [x] Document in specification (see MDNS_SPEC.md)

**Test:** ✅ TXT records created and parsed correctly

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

**Status:** ✅ Complete (Core + Integration)

### 4.1 Regular Multicast (Emergency Channels)

**Goal:** Any-source multicast for emergency

- [x] Group management
  - [x] Track subscribers per group (e.g., "emergency")
  - [x] Multiple broadcasters can send to same group
  - [x] Listeners receive from ALL sources in group

- [x] JOIN handling
  - [x] Listener sends JOIN("emergency")
  - [x] Discover all broadcasters for this group (mDNS)
  - [x] Subscribe to each broadcaster (unicast)

- [x] Broadcaster registration
  - [x] Broadcaster advertises group membership
  - [x] Accepts JOIN requests
  - [x] Sends RTP to all subscribers (unicast fan-out)

**Test:** ✅ multicast-test demonstrates 2 broadcasters, 4 listeners, same emergency group

### 4.2 SSM (Source-Specific Multicast) for Regular Channels

**Goal:** Subscribe to specific broadcaster

- [x] SSM subscription
  - [x] Listener specifies (Source, Group)
  - [x] Example: (203:abcd::1, "talk")
  - [x] Only receives from that specific source

- [x] SSM JOIN handling
  - [x] Send JOIN(source, group) → broadcaster
  - [x] Broadcaster adds to subscriber list
  - [x] Sends RTP unicast to subscriber

- [x] SSM LEAVE handling
  - [x] Explicit LEAVE message (PacketTypeUnsubscribe)
  - [x] Timeout (15s no heartbeat)
  - [x] Broadcaster removes from list

**Test:** ✅ multicast-test demonstrates SSM (listener only receives from COMMUNITY-A, not B)

### 4.3 Subscription Protocol

**Goal:** Manage subscriptions over unicast

- [x] Protocol packets (extend existing)
  - [x] SUBSCRIBE (source, group, listener_ipv6, port)
  - [x] UNSUBSCRIBE (PacketTypeUnsubscribe 0x12)
  - [x] HEARTBEAT (keepalive)
  - [x] Extended SubscribePayload with Group and SSMSource fields

- [x] Subscriber tracking
  - [x] Map: group → []subscribers
  - [x] Update LastSeen on heartbeat
  - [x] Prune stale (15s timeout)

- [x] Fan-out logic
  - [x] For each RTP packet
  - [x] Send to all subscribers in group
  - [x] Unicast to each subscriber's IPv6:port

**Test:** ✅ Subscribe, receive RTP, heartbeat, timeout works

### 4.4 Integration

- [x] Integrated with Broadcaster (internal/broadcaster/)
- [x] Integrated with Listener (internal/listener/)
- [x] Updated TUI (pkg/ui/)
- [x] Updated Web GUI (pkg/gui/)
- [x] Backward compatible with v0.4

**Documentation:**
- See MULTICAST_SPEC.md for full specification
- See LAYER4_PLAN.md for implementation details

---

## Layer 5: Application Layer (Emergency Features)

**Status:** ✅ Priority Signaling Complete | 🚧 Auto-tune To implement

### 5.1 Emergency Channels

**Goal:** Pre-defined emergency channels

- [x] Define channel registry
  - [x] emergency (8790) - critical priority
  - [x] netcontrol (8791) - emergency priority
  - [x] medical (8792) - emergency priority
  - [x] weather (8793) - high priority
  - [x] sar (8794) - emergency priority
  - [x] community (8795) - normal priority
  - [x] talk (8798) - normal priority
  - [x] test (8799) - normal priority

- [x] Load channel definitions (pkg/emergency/channels.go)
- [x] Emergency settings with auto-tune preferences
- [ ] UI shortcuts (e.g., "Emergency" button) - Layer 7

**Test:** ✅ Channel registry works, emergency-test program demonstrates

### 5.2 Priority Signaling

**Goal:** Emergency broadcasts interrupt normal traffic

- [x] Priority encoding in packet flags
  - [x] Use bits 4-5 of Flags field (not RTCP)
  - [x] Priority field: 0=normal, 1=high, 2=emergency, 3=critical
  - [x] SetPriority() and GetPriority() methods

- [x] Priority handling (receiver)
  - [x] Detect priority in audio packets
  - [x] Log priority changes
  - [x] Visual alerts with emojis (🚨 critical, ⚠️ emergency, 📢 high)
  - [ ] If critical: prompt user to switch - Future enhancement

- [x] Broadcaster priority assignment
  - [x] Automatic priority based on channel/group
  - [x] Priority set from channel registry on startup

- [ ] Auto-tune behavior - Future enhancement
  - [x] User preference types defined: always/prompt/never
  - [x] EmergencySettings struct with preferences
  - [ ] If "always": auto-switch to emergency
  - [ ] If "prompt": show notification
  - [ ] If "never": just log (currently implemented)

**Test:** ✅ emergency-test demonstrates priority detection and visual alerts

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
