# MeshRadio Foundation Design v2

**Version:** 2.0-standards
**Date:** 2025-11-30
**Status:** Architectural Blueprint (Standards-Based)
**Philosophy:** Use proven standards, add emergency-specific features

**Major Change from v1:** Adopting SSM/RTP/mDNS stack per community feedback

---

## Vision

**Build resilient emergency communication infrastructure using proven standards.**

When centralized services fail, communities need reliable ways to coordinate. MeshRadio provides HAM radio-style emergency communications over decentralized mesh networks using industry-standard protocols.

**Key Principle:** Build on proven standards (SSM, RTP, mDNS), add emergency-specific conventions.

---

## Architecture: Standards-Based Stack

### The Complete Stack

```
┌─────────────────────────────────────────────────────┐
│  APPLICATION LAYER (MeshRadio-Specific)             │
│  - Emergency channel management                     │
│  - Priority signaling                               │
│  - Net control automation                           │
│  - User interface (TUI/GUI/Mobile)                  │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  DISCOVERY LAYER (Standard: mDNS/Avahi)            │
│  - Service advertisement (Avahi/Bonjour)           │
│  - Source discovery via mDNS                        │
│  - SSM TXT record convention (new)                  │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  STREAMING LAYER (Standard: RTP over SSM)          │
│  - Real-time Transport Protocol (RFC 3550)         │
│  - Source-Specific Multicast (RFC 4607)            │
│  - Opus audio codec (RFC 6716)                     │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  TRANSPORT LAYER (Yggdrasil IPv6)                  │
│  - SSM tunneled over unicast                       │
│  - Encrypted, authenticated routing                │
└─────────────────────────────────────────────────────┘
```

### Why This Stack?

**From majestrate (Yggdrasil contributor):**
> "requires no new code to be written and existing tools can be adapted without needing special shims to speak yet another protocol"

**Benefits:**
- ✅ Use existing, battle-tested libraries
- ✅ Interoperable with standard RTP tools (VLC, GStreamer, etc.)
- ✅ Well-documented protocols (RFCs)
- ✅ Proven in production systems
- ✅ Security audited (RTP/SSM)

---

## Discovery: mDNS/Avahi (Standard + Emergency Extension)

### Standard mDNS Service Advertisement

```
Service Type: _meshradio._udp.local.

Example Advertisement:
  Name: EmergencyNet.Broadcast._meshradio._udp.local.
  Port: 8790
  TXT Records:
    ssm_source=203:0000:0000:0000::1
    ssm_group=ff3e::8000:emergency
    channel=emergency
    priority=critical
    callsign=W1EMERGENCY
    codec=opus
    bitrate=64
```

### mDNS TXT Record Convention (New)

**Required Fields:**
- `ssm_source` - IPv6 address of broadcaster (SSM source)
- `ssm_group` - SSM multicast group address
- `channel` - Channel designation (emergency, community, talk)
- `callsign` - Station callsign/identifier

**Optional Fields:**
- `priority` - normal, high, emergency, critical
- `codec` - opus, flac, etc.
- `bitrate` - kbps
- `location` - geographic info
- `description` - human-readable description

### Emergency Port Convention

**Dedicated Emergency "Frequencies":**

```yaml
# Emergency Port Range: 8790-8799
# Easy to remember, unlikely conflicts, works without mDNS

8790 - General Emergency Broadcast
       (like HAM 146.52 MHz simplex)

8791 - Net Control Coordination
       (organize emergency response)

8792 - Medical/Health
       (medical emergencies, coordination)

8793 - Infrastructure Status
       (power, water, comms status)

8794-8798 - Reserved Emergency Use

8799 - Regular Broadcast (non-emergency)

# Listener Port
9799 - Standard listener receive port
```

**Why Custom Ports?**
- Easy to communicate verbally in crisis
- Works when mDNS discovery fails
- No dependency on working DNS/discovery
- Self-documenting (87XX = emergency)

### Discovery Flow

**Normal Operation (mDNS working):**
```
1. Broadcaster advertises via mDNS
2. Listener queries: "_meshradio._udp.local."
3. Receives TXT records with SSM info
4. Subscribes to SSM group
5. Receives RTP stream
```

**Emergency Fallback (mDNS broken):**
```
1. Use known emergency port: 8790
2. Use known emergency channel: 203::1
3. Direct connection: meshradio tune 203::1:8790
4. No discovery needed
```

---

## Streaming: RTP over SSM

### Real-time Transport Protocol (RTP)

**Standard:** RFC 3550
**Audio Payload:** Opus (RFC 7587)

```
RTP Packet Format:
┌───────────────────────────────────────┐
│  RTP Header (12 bytes)                │
│  - Version, Padding, Extension        │
│  - Payload Type (Opus = 111)          │
│  - Sequence Number                    │
│  - Timestamp                          │
│  - SSRC (Synchronization Source)     │
├───────────────────────────────────────┤
│  Opus Audio Payload                   │
│  - 20ms frame (typical)               │
│  - 48kHz sample rate                  │
│  - 64kbps bitrate (configurable)      │
└───────────────────────────────────────┘
```

**Benefits:**
- Standard tools understand it (VLC, GStreamer, ffmpeg)
- Built-in sequence numbers (detect packet loss)
- Timestamps for synchronization
- Well-defined payload formats

### Source-Specific Multicast (SSM)

**Standard:** RFC 4607
**Group Range:** ff3e::/16 (organization-local SSM)

```
SSM Model:
  (Source, Group) = (Broadcaster IPv6, Multicast Group)

Example:
  Source: 203:0000:0000:0000::1
  Group:  ff3e::8000:emergency

Listeners subscribe to specific (S,G) pair
Only receive traffic from that source
```

**SSM over Yggdrasil:**
- SSM semantics at application layer
- Tunneled as unicast over Yggdrasil IPv6
- Broadcaster maintains subscriber list
- Fan-out to each subscriber

### Emergency Channel Groups

```yaml
# SSM Group Allocation (ff3e::/16)

ff3e::8000:emergency    # General emergency
ff3e::8001:netcontrol   # Net control
ff3e::8002:medical      # Medical coordination
ff3e::8003:infra        # Infrastructure status

ff3e::9000:community    # Community nets (by region)
ff3e::9001:neighborhood # Neighborhood-specific

ff3e::a000:talk         # Regular talk radio
```

---

## Implementation Architecture

### Service Layer (MeshRadio-Specific)

```go
// StationService - Emergency-aware RTP broadcaster
type StationService interface {
    // Broadcasting
    StartBroadcast(config BroadcastConfig) error
    StopBroadcast() error
    GetListeners() []Listener

    // Listening
    TuneTo(source net.IP, group net.IP, port int) error
    Disconnect() error

    // Emergency
    DeclareEmergency(channel Channel) error
    ClearEmergency(channel Channel) error
}

type BroadcastConfig struct {
    Callsign    string
    Channel     Channel        // emergency, community, talk
    Priority    Priority       // normal, high, emergency, critical
    Port        int           // 8790-8799 (emergency), 8799 (regular)

    // SSM Parameters
    SSMGroup    net.IP        // ff3e::8XXX:...

    // RTP Parameters
    PayloadType uint8         // 111 = Opus
    SampleRate  int          // 48000
    Bitrate     int          // 64000
}
```

### Discovery Service (Avahi Integration)

```go
// DiscoveryService - mDNS/Avahi wrapper
type DiscoveryService interface {
    // Service Advertisement
    Advertise(service ServiceInfo) error
    Withdraw() error

    // Service Discovery
    Browse(serviceType string) ([]ServiceInfo, error)
    Subscribe(callback func(ServiceInfo)) error

    // Emergency Queries
    FindEmergencyStations() ([]ServiceInfo, error)
}

type ServiceInfo struct {
    Name        string
    Type        string        // "_meshradio._udp"
    Domain      string        // "local."
    Port        int

    // TXT Record Data
    SSMSource   net.IP
    SSMGroup    net.IP
    Channel     string
    Priority    string
    Callsign    string
    Codec       string
    Bitrate     int
}
```

### RTP Service (Standard Library)

```go
// RTPService - Wrapper around RTP library (e.g., pion/rtp)
type RTPService interface {
    // Sender
    NewSender(config SenderConfig) (RTPSender, error)

    // Receiver
    NewReceiver(config ReceiverConfig) (RTPReceiver, error)
}

type RTPSender interface {
    WriteRTP(packet *rtp.Packet) error
    WriteOpus(pcm []byte) error  // Convenience wrapper
    Close() error
}

type RTPReceiver interface {
    ReadRTP() (*rtp.Packet, error)
    ReadOpus() ([]byte, error)   // Convenience wrapper
    Close() error
}
```

---

## Libraries & Dependencies

### Required Standard Libraries

**Discovery (mDNS/Avahi):**
```go
// Option 1: Avahi C library (Linux)
import "github.com/holoplot/go-avahi"

// Option 2: Pure Go mDNS
import "github.com/hashicorp/mdns"

// Option 3: Zeroconf (cross-platform)
import "github.com/grandcat/zeroconf"
```

**RTP Streaming:**
```go
// Pion WebRTC RTP implementation
import "github.com/pion/rtp"
import "github.com/pion/webrtc/v3"

// Or GStreamer Go bindings
import "github.com/tinyzimmer/go-gst/gst"
```

**Opus Audio Codec:**
```go
// Go Opus bindings
import "github.com/hrfee/gopus"

// Or via GStreamer
```

**SSM Implementation:**
```go
// Custom SSM overlay (tunnel over unicast)
// Build on existing RTP/multicast libraries
// Add subscriber tracking + fan-out
```

### Platform Support

```yaml
Desktop:
  - Avahi (Linux)
  - Bonjour (macOS)
  - Bonjour for Windows

Mobile:
  - NSD (Network Service Discovery) on Android
  - Bonjour on iOS

Embedded:
  - Lightweight mDNS responder
  - Minimal RTP sender/receiver
```

---

## Emergency Features (Application Layer)

### Priority Signaling

**RTCP Extension (Emergency Priority):**
```go
// Add custom RTCP packet for emergency priority
type EmergencyRTCP struct {
    Type       uint8   // RTCP type (custom)
    Priority   uint8   // 0=normal, 1=high, 2=emergency, 3=critical
    Channel    string  // emergency, medical, etc.
    Message    string  // Optional emergency message
}
```

**Auto-Tune Behavior:**
```
On receiving emergency RTCP:
1. Log emergency event
2. Notify user (sound alert, notification)
3. If priority >= emergency:
   - Pause current stream
   - Auto-tune to emergency channel
   - Show emergency message
4. User can override/dismiss
```

### Net Control Protocol

**Built on RTP/RTCP:**
```go
type NetControlMessage struct {
    Command    string  // "check-in", "stand-by", "report", "all-clear"
    Target     string  // Callsign or "all"
    Message    string
}

// Sent as RTCP APP packet (RFC 3550 section 6.7)
```

**Net Control Flow:**
```
1. Net Control announces on channel (RTCP APP)
2. Stations acknowledge (RTCP sender report)
3. Net Control coordinates check-ins
4. Status reports via RTCP
5. All-clear signal when done
```

### Emergency Channels

```yaml
# Pre-defined Emergency Channels

emergency:
  channel: "emergency"
  port: 8790
  ssm_group: ff3e::8000:emergency
  priority: critical
  auto_tune: true
  description: "General emergency broadcast"

netcontrol:
  channel: "netcontrol"
  port: 8791
  ssm_group: ff3e::8001:netcontrol
  priority: emergency
  auto_tune: false
  description: "Emergency coordination"

medical:
  channel: "medical"
  port: 8792
  ssm_group: ff3e::8002:medical
  priority: emergency
  auto_tune: prompt
  description: "Medical emergencies"
```

---

## SSM Overlay Implementation

### Tunneling SSM over Yggdrasil Unicast

**Concept:**
- Application-layer SSM semantics
- Actual transport: Yggdrasil IPv6 unicast
- Broadcaster tracks subscribers
- Fan-out RTP packets to each

**Implementation:**

```go
type SSMOverlay struct {
    // Subscriber management
    subscribers  map[string]*Subscriber  // key: "ipv6:port"
    mu           sync.RWMutex

    // RTP sender
    rtpSender    *RTPSender

    // SSM group info
    sourceIPv6   net.IP
    groupIPv6    net.IP
}

func (s *SSMOverlay) HandleJoin(subscriber net.IP, port int) {
    // Add subscriber to registry
    s.mu.Lock()
    s.subscribers[fmt.Sprintf("%s:%d", subscriber, port)] = &Subscriber{
        IPv6:     subscriber,
        Port:     port,
        JoinedAt: time.Now(),
    }
    s.mu.Unlock()
}

func (s *SSMOverlay) SendRTP(packet *rtp.Packet) {
    // Fan out to all subscribers (unicast)
    s.mu.RLock()
    defer s.mu.RUnlock()

    for _, sub := range s.subscribers {
        go s.sendToSubscriber(sub, packet)
    }
}
```

**IGMP/MLD Equivalent:**
```
Standard SSM:    IGMPv3/MLDv2 JOIN message
MeshRadio SSM:   SUBSCRIBE packet with (Source, Group)

Standard SSM:    IGMPv3/MLDv2 LEAVE message
MeshRadio SSM:   UNSUBSCRIBE or timeout

Standard SSM:    Periodic membership reports
MeshRadio SSM:   Heartbeat packets
```

---

## Deployment Modes

### Mode 1: Daemon + API (Recommended)

```
┌────────────────────────────────┐
│  meshradio-daemon              │
│  - Avahi advertisement         │
│  - RTP sender/receiver         │
│  - SSM overlay                 │
│  - gRPC/REST API               │
└────────────────────────────────┘
         ↑ (API)
    ┌────┴────┬────────┬────────┐
    │         │        │        │
  [TUI]    [Web GUI] [CLI]  [Mobile]
```

### Mode 2: Library (Mobile)

```
┌────────────────────────────────┐
│  Yggdrasil Mobile App          │
│  └── libmeshradio.so           │
│      - mDNS responder          │
│      - RTP stack               │
│      - SSM overlay             │
└────────────────────────────────┘
```

### Mode 3: Standalone (Testing)

```
┌────────────────────────────────┐
│  meshradio (all-in-one)        │
│  - Embedded mDNS               │
│  - RTP sender/receiver         │
│  - Built-in TUI                │
└────────────────────────────────┘
```

---

## Implementation Phases

### Phase 1: Standards Foundation (Jan 2026)

**Week 1-2: RTP Streaming**
- [ ] Integrate RTP library (pion/rtp)
- [ ] Opus codec integration
- [ ] Basic sender/receiver
- [ ] Test: Send/receive RTP locally

**Week 3-4: mDNS Discovery**
- [ ] Integrate Avahi/zeroconf
- [ ] Define mDNS TXT record format
- [ ] Service advertisement
- [ ] Service browsing
- [ ] Test: Discovery on local network

**Week 5-6: SSM Overlay**
- [ ] Subscriber tracking
- [ ] JOIN/LEAVE handling
- [ ] RTP fan-out to subscribers
- [ ] Test: Multi-listener scenario

**Week 7-8: Emergency Features**
- [ ] Emergency port convention (8790-8799)
- [ ] Priority RTCP extension
- [ ] Emergency channel definitions
- [ ] Auto-tune logic
- [ ] Test: Emergency broadcast

### Phase 2: Integration & Testing (Feb 2026)

**Week 1-2: Daemon Mode**
- [ ] gRPC API definition
- [ ] Daemon service
- [ ] CLI client
- [ ] API documentation

**Week 3-4: Real Mesh Testing**
- [ ] Deploy on Yggdrasil mesh (3+ nodes)
- [ ] mDNS over mesh validation
- [ ] RTP latency measurement
- [ ] Emergency scenario testing

### Phase 3: Mobile & Production (Mar-Apr 2026)

**Week 1-2: Mobile Library**
- [ ] C library API (libmeshradio)
- [ ] Android integration
- [ ] iOS integration
- [ ] Background service

**Week 3-4: Production Hardening**
- [ ] Error handling
- [ ] Metrics & monitoring
- [ ] Documentation
- [ ] Community testing

---

## Open Questions for Community

### Technical

**@majestrate:**
1. SSM tunneling over Yggdrasil - any gotchas we should watch for?
2. mDNS across mesh segments - relay needed or works natively?
3. Emergency priority - RTCP extension vs. separate signaling?

**For the mDNS TXT record:**
```
_meshradio._udp.local.
  TXT "ssm_source=201:abcd::1"
      "ssm_group=ff3e::8000:emergency"
      "channel=emergency"
      "priority=critical"
```
Does this format work, or should we follow existing convention?

### Emergency Use Case

**@Revertron:**
1. Emergency port range (8790-8799) - good choice or different numbers?
2. Auto-tune to emergency - acceptable behavior or too intrusive?

---

## Success Criteria

### Phase 1 (Standards Foundation)
- [ ] RTP streaming works with standard tools (VLC can play)
- [ ] mDNS discovery finds stations within 10s
- [ ] Emergency broadcast overrides normal traffic
- [ ] Tested on 3+ Yggdrasil nodes

### Phase 2 (Production Ready)
- [ ] Daemon API fully documented
- [ ] Mobile library integrated in Yggdrasil apps
- [ ] Emergency drill successful (community test)
- [ ] 100+ concurrent listeners per broadcaster

### Phase 3 (Widely Deployed)
- [ ] Used in real community emergency
- [ ] Interoperates with standard RTP tools
- [ ] Sub-second emergency alert propagation
- [ ] 1000+ users

---

## Migration from v0.4

### What to Keep
- Service layer architecture concept
- Emergency channel philosophy
- Daemon/library pattern

### What to Replace
- Custom subscription protocol → SSM
- Custom audio packets → RTP
- Gossip discovery → mDNS/Avahi

### Migration Path
```
1. Keep v0.4 working (backward compat mode)
2. Add RTP/mDNS alongside
3. Deprecate custom protocol
4. Remove after 6 months
```

---

## References

### Standards
- RFC 3550 - RTP (Real-time Transport Protocol)
- RFC 4607 - SSM (Source-Specific Multicast)
- RFC 6762 - mDNS (Multicast DNS)
- RFC 6716 - Opus Audio Codec
- RFC 7587 - RTP Payload Format for Opus

### Libraries
- Pion RTP: https://github.com/pion/rtp
- go-avahi: https://github.com/holoplot/go-avahi
- zeroconf: https://github.com/grandcat/zeroconf
- gopus: https://github.com/hrfee/gopus

### Prior Art
- GStreamer RTP examples
- PulseAudio RTP module
- VLC network streaming

---

**Foundation: Standards-based. Emergency-ready. Built to last.**
