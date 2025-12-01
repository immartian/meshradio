# MeshRadio Foundation Design

**Version:** 1.0-foundation
**Date:** 2025-11-30
**Status:** Architectural Blueprint
**Philosophy:** Infrastructure first, UI later

---

## Vision

**Build resilient emergency communication infrastructure for mesh networks.**

When centralized services fail (AWS outage, natural disaster, infrastructure attack), communities need reliable ways to coordinate. MeshRadio provides HAM radio-style emergency communications over decentralized mesh networks.

**Key Principle:** The foundation is a **robust, standards-based infrastructure** that any UI can consume - desktop TUI, web GUI, mobile apps, embedded devices, even voice assistants.

---

## Architecture Philosophy

### 1. **Separation of Concerns**

```
┌─────────────────────────────────────────────┐
│  UI LAYER (Multiple Implementations)        │
│  - Desktop TUI                              │
│  - Web GUI                                  │
│  - Mobile App (Yggdrasil Android/iOS)      │
│  - CLI tools                                │
│  - Future: Embedded, Voice, etc.           │
└─────────────────────────────────────────────┘
                    ↓ (API)
┌─────────────────────────────────────────────┐
│  SERVICE LAYER (Core Infrastructure)        │
│  - Station Service (broadcast/listen)      │
│  - Discovery Service (find stations)       │
│  - Coordination Service (emergency nets)   │
│  - Storage Service (local state)           │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│  PROTOCOL LAYER (Wire Format)               │
│  - Audio streaming                          │
│  - Subscription management                  │
│  - Discovery protocol                       │
│  - Emergency signaling                      │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│  TRANSPORT LAYER (Network)                  │
│  - Yggdrasil IPv6 (primary)                │
│  - Regular IPv6 (fallback)                 │
│  - Future: Direct RF                        │
└─────────────────────────────────────────────┘
```

### 2. **Design Constraints**

**MUST:**
- Work without any centralized infrastructure
- Survive network partitions gracefully
- Support emergency priority traffic
- Be debuggable in crisis situations
- Run on resource-constrained devices

**MUST NOT:**
- Require DNS, NTP, or cloud services
- Depend on specific UI framework
- Assume stable network topology
- Require manual configuration in emergency

### 3. **Deployment Models**

**Daemon Mode (Recommended):**
```
meshradio-daemon
  ├── Runs as background service
  ├── Exposes gRPC/REST API
  ├── Manages persistent state
  └── Multiple UIs can connect
```

**Embedded Mode:**
```
libmeshradio.so / meshradio.dylib / meshradio.dll
  ├── C library API
  ├── Embeddable in mobile apps
  ├── No daemon required
  └── Direct integration
```

**Standalone Mode:**
```
meshradio (current TUI/GUI)
  ├── All-in-one binary
  ├── No API exposed
  └── Quick testing/deployment
```

---

## Core Infrastructure Components

### Component 1: Station Service

**Responsibility:** Audio broadcast and reception

```go
// Station Service API
type StationService interface {
    // Broadcasting
    StartBroadcast(config BroadcastConfig) (Station, error)
    StopBroadcast(stationID string) error
    GetBroadcastStatus(stationID string) (Status, error)
    GetListeners(stationID string) ([]Listener, error)

    // Listening
    TuneTo(targetIPv6 net.IP, port int) (Stream, error)
    Disconnect() error
    GetStreamStatus() (StreamStatus, error)

    // Audio
    SetAudioInput(deviceID string) error
    SetAudioOutput(deviceID string) error
    GetAudioDevices() ([]AudioDevice, error)
}

type BroadcastConfig struct {
    Callsign    string
    Channel     Channel        // 203::/8 = emergency, etc.
    Priority    Priority       // Normal, High, Emergency
    Codec       CodecType      // Opus, FLAC, etc.
    Bitrate     int           // kbps
    Metadata    map[string]string
}

type Listener struct {
    IPv6        net.IP
    Callsign    string
    ConnectedAt time.Time
    LastSeen    time.Time
    PacketLoss  float64
    Latency     time.Duration
}
```

**Implementation Notes:**
- Must work as both library and daemon service
- Audio backend pluggable (PortAudio, PulseAudio, CoreAudio, Android Audio)
- Codec support modular (start Opus, add others later)
- Statistics tracking for monitoring

---

### Component 2: Discovery Service

**Responsibility:** Find and track active stations

```go
type DiscoveryService interface {
    // Discovery
    FindStations(filter Filter) ([]StationInfo, error)
    SubscribeToUpdates(callback func(StationInfo)) (Subscription, error)

    // Registration
    Register(station StationInfo) error
    Unregister(stationID string) error
    UpdateMetadata(stationID string, metadata Metadata) error

    // Queries
    GetStation(ipv6 net.IP) (StationInfo, error)
    GetStationsByBand(band string) ([]StationInfo, error)
    GetEmergencyStations() ([]StationInfo, error)
    GetNearbyStations(hops int) ([]StationInfo, error)
}

type StationInfo struct {
    IPv6        net.IP
    Port        int
    Callsign    string
    Channel     Channel
    Priority    Priority
    Metadata    Metadata
    LastSeen    time.Time
    Quality     SignalQuality
    Hops        int           // Network distance
}

type Filter struct {
    Bands      []string      // "emergency", "community", etc.
    Priority   *Priority     // Filter by priority level
    MaxHops    *int         // Network proximity
    MinQuality *int         // Signal quality threshold
}
```

**Discovery Mechanisms (Pluggable):**

```go
// Discovery backend interface
type DiscoveryBackend interface {
    Announce(station StationInfo) error
    Query(filter Filter) ([]StationInfo, error)
    Subscribe(callback func(StationInfo)) error
}

// Implementations:
type GossipDiscovery struct { ... }    // Epidemic broadcast
type DHTDiscovery struct { ... }       // Distributed hash table
type SSMDiscovery struct { ... }       // SSM overlay (if feasible)
type ManualDiscovery struct { ... }    // Static config file
```

**Priority:** Start with **Gossip + Manual**, add DHT/SSM later

---

### Component 3: Coordination Service

**Responsibility:** Emergency nets, net control, priority management

```go
type CoordinationService interface {
    // Emergency
    DeclareEmergency(channel Channel, reason string) error
    ClearEmergency(channel Channel) error
    GetEmergencyStatus() ([]Emergency, error)

    // Net Control
    StartNetControl(channel Channel) (NetControl, error)
    JoinNet(channel Channel) error
    LeaveNet(channel Channel) error

    // Priority
    SetPriority(level Priority) error
    GetPriorityQueue() ([]Message, error)
}

type Emergency struct {
    Channel     Channel
    DeclaredBy  string        // Callsign
    DeclaredAt  time.Time
    Reason      string
    Status      EmergencyStatus
}

type NetControl struct {
    Channel     Channel
    Controller  string        // Callsign
    Participants []string
    StartedAt   time.Time
}

type Priority int
const (
    PriorityNormal Priority = iota
    PriorityHigh
    PriorityEmergency
    PriorityCritical
)
```

**Emergency Behavior:**
- Emergency broadcasts interrupt normal traffic
- Auto-tune to emergency channel (user override available)
- Emergency logs persisted for after-action review

---

### Component 4: Storage Service

**Responsibility:** Persistent state, logs, configuration

```go
type StorageService interface {
    // Configuration
    GetConfig() (Config, error)
    SetConfig(config Config) error

    // Station Database
    SaveStation(info StationInfo) error
    GetStations() ([]StationInfo, error)
    DeleteStation(ipv6 net.IP) error

    // Logs
    LogEvent(event Event) error
    GetLogs(filter LogFilter) ([]Event, error)

    // Bookmarks
    SaveBookmark(bookmark Bookmark) error
    GetBookmarks() ([]Bookmark, error)
}

type Event struct {
    Timestamp   time.Time
    Type        EventType      // Broadcast, Listen, Emergency, etc.
    Severity    Severity       // Info, Warning, Error, Critical
    Message     string
    Metadata    map[string]interface{}
}
```

**Storage Backends:**
- SQLite for desktop/daemon (current)
- LevelDB for embedded (lighter)
- File-based for emergency fallback

---

## Protocol Layer Design

### Wire Protocol Versioning

```go
// Protocol version negotiation
const (
    ProtocolVersion1 = 0x01  // v0.4 subscription-based
    ProtocolVersion2 = 0x02  // Future: SSM overlay
)

type PacketHeader struct {
    Version    uint8    // Protocol version
    Type       uint8    // Packet type
    Flags      uint8    // Feature flags
    ...
}
```

**Backward Compatibility:**
- Nodes negotiate highest common version
- v1 (subscription) always supported
- v2 (SSM) optional enhancement

### Packet Types (Extended)

```go
const (
    // v1 - Current
    PacketTypeAudio          = 0x01
    PacketTypeSubscribe      = 0x10
    PacketTypeHeartbeat      = 0x11

    // v1.1 - Emergency
    PacketTypeEmergency      = 0x20
    PacketTypePriority       = 0x21
    PacketTypeNetControl     = 0x22

    // v1.2 - Discovery
    PacketTypeGossip         = 0x30
    PacketTypeStationQuery   = 0x31
    PacketTypeStationReply   = 0x32

    // v2 - Future (SSM)
    PacketTypeSSMJoin        = 0x40
    PacketTypeSSMLeave       = 0x41
    PacketTypeSSMReport      = 0x42
)
```

### Emergency Packet Format

```go
type EmergencyPacket struct {
    Channel      [16]byte   // Emergency channel IPv6
    Severity     uint8      // Emergency severity
    Reason       [128]byte  // UTF-8 reason string
    OriginIPv6   [16]byte   // Declaring station
    Timestamp    uint64     // Unix timestamp
}
```

**Emergency Behavior:**
1. Node receives EMERGENCY packet
2. Logs emergency event
3. Notifies UI layer (if connected)
4. UI decides: auto-tune, notify user, or ignore

---

## Discovery Protocol: Gossip (Phase 1)

### Why Gossip?

- ✅ Simple, proven for mesh networks
- ✅ Self-healing on partition/merge
- ✅ No single point of failure
- ✅ Works offline (LAN-only)
- ✅ Low complexity, debuggable in crisis

### Gossip Protocol Spec

```go
type GossipMessage struct {
    Stations   []StationInfo  // Known stations
    SeqNum     uint64        // Increasing sequence number
    TTL        uint8         // Hops remaining
}

// Gossip algorithm
func (g *GossipService) Run() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        // 1. Pick random peers
        peers := g.selectRandomPeers(3)

        // 2. Send our station list
        msg := g.createGossipMessage()
        for _, peer := range peers {
            g.send(peer, msg)
        }

        // 3. Merge received lists
        g.mergeStationLists()

        // 4. Age out stale entries
        g.pruneStaleStations(5 * time.Minute)
    }
}
```

**Station Freshness:**
- Stations have TTL (time-to-live)
- Refreshed by periodic heartbeat
- Expired stations pruned from gossip
- UI shows "last seen" timestamp

---

## API Design (Daemon Mode)

### gRPC Service Definition

```protobuf
service MeshRadio {
    // Station control
    rpc StartBroadcast(BroadcastRequest) returns (BroadcastResponse);
    rpc StopBroadcast(StopRequest) returns (EmptyResponse);
    rpc TuneTo(TuneRequest) returns (TuneResponse);
    rpc Disconnect(EmptyRequest) returns (EmptyResponse);

    // Discovery
    rpc FindStations(FindRequest) returns (stream StationInfo);
    rpc GetStation(GetStationRequest) returns (StationInfo);

    // Emergency
    rpc DeclareEmergency(EmergencyRequest) returns (EmptyResponse);
    rpc GetEmergencies(EmptyRequest) returns (stream Emergency);

    // Status
    rpc GetStatus(EmptyRequest) returns (StatusResponse);
    rpc SubscribeEvents(EmptyRequest) returns (stream Event);
}
```

**Alternative: REST API**

```
POST   /api/v1/broadcast/start
POST   /api/v1/broadcast/stop
GET    /api/v1/broadcast/status
GET    /api/v1/broadcast/listeners

POST   /api/v1/listen/tune
POST   /api/v1/listen/disconnect
GET    /api/v1/listen/status

GET    /api/v1/stations
GET    /api/v1/stations/:ipv6
POST   /api/v1/stations/search

POST   /api/v1/emergency/declare
POST   /api/v1/emergency/clear
GET    /api/v1/emergency/active

GET    /api/v1/events (SSE stream)
```

**Both APIs expose the same Service Layer**

---

## Channel Plan (Emergency-First)

```yaml
# Emergency Channels (203::/8)
emergency:
  base: "203::"
  channels:
    general:      "203:0000::1"  # Any emergency
    netcontrol:   "203:1000::1"  # Coordinate response
    medical:      "203:2000::1"  # Medical emergencies
    infrastructure: "203:3000::1"  # Power/water/comms status
    evacuation:   "203:4000::1"  # Evacuation coordination
    allclear:     "203:f000::1"  # Emergency resolved

# Community Nets (204::/8)
community:
  base: "204::"
  organization: geographic
  examples:
    north_side:   "204:1000::/20"
    downtown:     "204:2000::/20"
    university:   "204:3000::/20"

# Regular Talk (202::/8)
talk:
  base: "202::"
  purpose: non-emergency communications
```

---

## Mobile App Integration

### Architecture

```
┌─────────────────────────────────┐
│  Yggdrasil Mobile App           │
│  (Existing Android/iOS)         │
├─────────────────────────────────┤
│  + MeshRadio Plugin/Module      │
│    - libmeshradio.so/.dylib     │
│    - Native UI (React Native?)  │
│    - Background service         │
└─────────────────────────────────┘
         ↓ (uses)
┌─────────────────────────────────┐
│  libmeshradio (C Library)       │
│  - Station Service              │
│  - Discovery Service            │
│  - Coordination Service         │
│  - FFI bindings (Rust/Go)       │
└─────────────────────────────────┘
         ↓
┌─────────────────────────────────┐
│  Yggdrasil Core                 │
│  (IPv6 mesh transport)          │
└─────────────────────────────────┘
```

### Mobile-Specific Considerations

**Battery:**
- Adaptive gossip interval (30s active, 5min background)
- Suspend audio processing when screen off
- Low-power listening mode

**Permissions:**
- Microphone (for broadcast)
- Notifications (for emergency alerts)
- Background execution (for net monitoring)

**UX:**
- Emergency notification banner
- Quick-tune to emergency channel
- Offline mode (station list cached)

---

## Implementation Phases

### Phase 1: Solid Foundation (Dec 2025 - Jan 2026)

**Goal:** Production-ready infrastructure

```
Week 1-2: Service Layer
- [ ] Refactor v0.4 into Service architecture
- [ ] Station Service with proper API
- [ ] Storage Service (SQLite)
- [ ] Unit tests for each service

Week 3-4: Discovery
- [ ] Gossip protocol implementation
- [ ] Discovery Service API
- [ ] Manual discovery (config file)
- [ ] Integration tests

Week 5-6: Emergency Support
- [ ] Coordination Service
- [ ] Emergency packet types
- [ ] Channel plan implementation
- [ ] Priority queue

Week 7-8: API Layer
- [ ] gRPC service definition
- [ ] Daemon mode
- [ ] CLI client (test API)
- [ ] Documentation
```

### Phase 2: Mobile-Ready (Feb-Mar 2026)

```
Week 1-2: Library Packaging
- [ ] C library API (libmeshradio)
- [ ] FFI bindings (if Go, use cgo)
- [ ] Mobile build system
- [ ] API stability guarantees

Week 3-4: Mobile Integration
- [ ] Android module (Yggdrasil app)
- [ ] iOS module
- [ ] Background service
- [ ] Battery optimization

Week 5-6: Mobile UI
- [ ] React Native UI (or native)
- [ ] Emergency notifications
- [ ] Simple broadcast/listen
- [ ] Beta testing
```

### Phase 3: Advanced Features (Apr-Jun 2026)

```
- [ ] SSM overlay (if recommended by community)
- [ ] DHT discovery
- [ ] Repeater/relay support
- [ ] Multi-source coordination
- [ ] QoS/priority enforcement
```

---

## Technical Decisions

### Language Choice

**Current: Go**
- ✅ Good for daemon services
- ✅ Cross-platform
- ✅ Good mobile support (gomobile)
- ✅ Existing Yggdrasil integration

**Alternative: Rust**
- ✅ Better for library (smaller binaries)
- ✅ Excellent FFI
- ✅ Memory safety guarantees
- ❌ More complex mobile integration

**Recommendation:** Stay with Go, optimize later if needed

### Audio Backend

**Desktop:**
- PortAudio (current, works well)

**Mobile:**
- Android: AAudio / OpenSL ES
- iOS: CoreAudio / AVAudioEngine

**Abstraction:**
```go
type AudioBackend interface {
    OpenInputStream(config StreamConfig) (InputStream, error)
    OpenOutputStream(config StreamConfig) (OutputStream, error)
    ListDevices() ([]Device, error)
}
```

### State Management

**Daemon Mode:**
- SQLite for persistent state
- In-memory cache for active sessions
- Write-ahead logging for crash recovery

**Embedded Mode:**
- Caller provides storage interface
- Can be in-memory only

---

## Testing Strategy

### Unit Tests
```
- [ ] Each service layer component
- [ ] Protocol encoding/decoding
- [ ] Gossip algorithm
- [ ] Emergency priority logic
```

### Integration Tests
```
- [ ] Multi-node simulation (3+ nodes)
- [ ] Network partition scenarios
- [ ] Emergency cascade behavior
- [ ] Discovery convergence
```

### System Tests
```
- [ ] Actual Yggdrasil mesh (2+ machines)
- [ ] Mobile app on real devices
- [ ] Battery usage profiling
- [ ] Network bandwidth measurement
```

### Emergency Drills
```
- [ ] Community scenario testing
- [ ] Failure mode validation
- [ ] Response time measurement
- [ ] User experience feedback
```

---

## Security Considerations

### Threat Model

**In Scope:**
- DoS attacks (message flooding)
- Spam broadcasts
- Malicious emergency declarations
- Station impersonation

**Out of Scope (Yggdrasil handles):**
- Traffic interception (encrypted by Ygg)
- IP spoofing (authenticated routing)
- Man-in-the-middle (Ygg prevents)

### Mitigations

**Rate Limiting:**
```go
type RateLimiter struct {
    maxBroadcasts   int  // Per hour
    maxEmergencies  int  // Per day
    maxGossip       int  // Per minute
}
```

**Reputation:**
```go
type StationReputation struct {
    IPv6              net.IP
    EmergencyFalseAlarms int
    SpamReports       int
    UptimeSeconds     uint64
    Score             float64  // 0-100
}
```

**User Controls:**
- Block/mute stations
- Require minimum reputation for auto-tune
- Emergency confirmation dialog

---

## Metrics & Monitoring

### Key Metrics

**Station Service:**
- Active broadcasts
- Active listeners
- Audio packet loss %
- Latency (p50, p95, p99)

**Discovery Service:**
- Stations in database
- Discovery query rate
- Gossip message rate
- Cache hit rate

**Emergency:**
- Active emergencies
- Emergency response time
- False alarm rate

**System:**
- CPU usage
- Memory usage
- Network bandwidth
- Battery drain (mobile)

### Observability

```go
// Prometheus-style metrics
type Metrics interface {
    IncrementCounter(name string, value int)
    SetGauge(name string, value float64)
    RecordHistogram(name string, value float64)
}

// Example usage
metrics.IncrementCounter("meshradio_broadcasts_total", 1)
metrics.SetGauge("meshradio_active_listeners", listenerCount)
metrics.RecordHistogram("meshradio_audio_latency_ms", latency)
```

---

## Documentation Requirements

### For Developers

- [ ] Service Layer API docs
- [ ] Protocol specification
- [ ] Gossip algorithm details
- [ ] Build/test instructions
- [ ] Mobile integration guide

### For Operators

- [ ] Daemon deployment guide
- [ ] Configuration reference
- [ ] Emergency channel usage
- [ ] Troubleshooting guide
- [ ] Best practices

### For End Users

- [ ] Quick start guide
- [ ] Emergency procedures
- [ ] Privacy & security
- [ ] FAQ

---

## Questions for Community

### Technical Architecture

**@majestrate:**
1. SSM overlay: Should we build this into foundation, or add later?
2. For emergency use case, any protocol considerations we're missing?
3. Performance targets: How many simultaneous broadcasters can Ygg handle per node?

**@parnikkapore:**
1. Gossip vs. DHT: Which is more resilient for partitioned networks?
2. Mobile battery: Any Yggdrasil mobile optimizations we should follow?

### Emergency Use Case

**@Revertron:**
1. Net control features: What's essential vs. nice-to-have?
2. Real HAM radio operators: What would they expect from digital equivalent?

---

## Success Criteria

### Phase 1 (Foundation)
- [ ] 3+ nodes mesh tested successfully
- [ ] Emergency broadcast works end-to-end
- [ ] Discovery finds stations within 30s
- [ ] Daemon API fully documented
- [ ] Community feedback positive

### Phase 2 (Mobile)
- [ ] Android/iOS apps in beta
- [ ] Battery drain <5% per hour listening
- [ ] Emergency notifications reliable
- [ ] 100+ users testing

### Phase 3 (Production)
- [ ] Used in real community drill
- [ ] Handles 50+ simultaneous listeners per broadcaster
- [ ] 99.9% emergency message delivery
- [ ] Sub-second emergency alert propagation

---

## Next Steps

**This Week:**
1. Post this design to GitHub for community review
2. Get feedback on SSM vs. Gossip
3. Validate emergency channel plan
4. Test v0.4 on real Yggdrasil mesh

**Next 2 Weeks:**
1. Refactor v0.4 into Service Layer architecture
2. Implement basic Gossip discovery
3. Add emergency packet types
4. Create gRPC API definition

**By End of December:**
1. Foundation complete
2. API documented
3. CLI client working
4. Ready for mobile integration

---

**This is the foundation. Everything else builds on top.**
