# MeshRadio v0.4-alpha Architecture

**Date:** 2025-11-29
**Status:** MVP Implementation Complete

## TL;DR

Redesigned from multicast to **subscription-based unicast streaming** after discovering Yggdrasil doesn't support IPv6 multicast routing.

## Key Changes from v0.3

### Before (v0.3 - Broken)
```
Broadcaster → ff02::1 (multicast) → ❌ Doesn't work across Yggdrasil mesh
                                     ✅ Only works on local LAN
```

### After (v0.4 - Works)
```
Listener → SUBSCRIBE → Broadcaster
Broadcaster → Unicast UDP → Listener 1
                          → Listener 2
                          → Listener N

Listener → HEARTBEAT (every 5s) → Broadcaster
```

## Why the Change?

**Discovery:** Yggdrasil does NOT route IPv6 multicast across the mesh overlay.

From [yggdrasil-network/yggdrasil-go#991](https://github.com/yggdrasil-network/yggdrasil-go/discussions/991):
> "Yggdrasil does not do any routing of multicast traffic at the moment, largely due to the fact that we don't yet have a good way of sharing group memberships in a way that actually scales."

**What Yggdrasil supports:**
- ✅ Link-local multicast for peer discovery (finding Yggdrasil nodes on LAN)
- ✅ Unicast IPv6 routing across the mesh
- ❌ IPv6 multicast routing across the mesh

## New Protocol

### Packet Types Added

```go
PacketTypeSubscribe  = 0x10  // Listener → Broadcaster: "I want to listen"
PacketTypeHeartbeat  = 0x11  // Listener → Broadcaster: "I'm still here"
```

### Subscribe Flow

**1. Listener initiates subscription:**
```
Listener sends to Broadcaster's IPv6:port:
{
  Type: SUBSCRIBE,
  Payload: {
    ListenerIPv6: [16]byte,  // Where to send audio
    ListenerPort: uint16,     // Listener's port
    Callsign:     [16]byte,   // Who's listening
  }
}
```

**2. Broadcaster starts streaming:**
- Adds listener to internal registry
- Immediately starts sending AUDIO packets to ListenerIPv6:ListenerPort
- No ACK required (fire-and-forget for MVP)

**3. Listener sends heartbeats:**
```
Every 5 seconds:
{
  Type: HEARTBEAT,
  Payload: {
    ListenerIPv6: [16]byte,
    Timestamp:    uint64,
  }
}
```

**4. Broadcaster monitors heartbeats:**
- Updates `LastSeen` timestamp when heartbeat received
- Removes listeners who haven't sent heartbeat in 15 seconds

## Implementation Details

### Broadcaster Changes

**New state:**
```go
type ListenerConn struct {
    IPv6        net.IP
    Port        uint16
    Callsign    string
    ConnectedAt time.Time
    LastSeen    time.Time
}

type Broadcaster struct {
    // ... existing fields ...
    listeners    map[string]*ListenerConn  // key: "ipv6:port"
    listenersMux sync.RWMutex
}
```

**New goroutines:**
1. `subscriptionLoop()` - Handles incoming SUBSCRIBE and HEARTBEAT packets
2. `broadcastLoop()` - Modified to send to each listener (unicast)
3. `heartbeatMonitor()` - Removes stale listeners (15s timeout)

**Removed:**
- ~~`beaconLoop()`~~ - No longer needed
- ~~`multicastAddr`~~ - No multicast anymore

### Listener Changes

**New state:**
```go
type Listener struct {
    // ... existing fields ...
    callsign      string
    localIPv6     net.IP
    localPort     int
    subscribed    bool
    lastHeartbeat time.Time
}
```

**New methods:**
1. `subscribe()` - Sends SUBSCRIBE packet on Start()
2. `heartbeatLoop()` - Sends HEARTBEAT every 5 seconds

### Protocol Helpers

New file: `pkg/protocol/subscription.go`

```go
func MarshalSubscribe(*SubscribePayload) []byte
func UnmarshalSubscribe([]byte) (*SubscribePayload, error)
func MarshalHeartbeat(*HeartbeatPayload) []byte
func UnmarshalHeartbeat([]byte) (*HeartbeatPayload, error)
```

## Discovery Strategy (MVP)

**No built-in discovery** for MVP. Use out-of-band methods:

### Option 1: Manual Share (Recommended for MVP)
```bash
# Broadcaster shares via chat/email/etc:
"Tune to: 201:abcd:1234::1:8799"

# Listener uses that address
meshradio listen 201:abcd:1234::1:8799
```

### Option 2: Simple Text File
```yaml
# ~/.meshradio/stations.yaml
stations:
  - name: "Alice's Station"
    ipv6: "201:abcd:1234::1"
    port: 8799
    callsign: "W1ALICE"

  - name: "Bob's Music"
    ipv6: "202:beef:cafe::2"
    port: 8799
    callsign: "W2BOB"
```

### Future: DHT/Directory Service
Phase 2 will add automatic discovery via DHT or directory nodes.

## User Experience

### Broadcaster
```bash
$ meshradio broadcast --callsign W1ALICE
> Broadcasting on 201:abcd:1234::1:8799
> Share this address with listeners!
>
> 0 listeners connected
> [5s later] New listener: W2BOB (202:5678::2)
> 1 listener connected
> Broadcasting: seq=50, size=120 bytes to 1 listeners
```

### Listener
```bash
$ meshradio listen 201:abcd:1234::1:8799 --callsign W2BOB
> Subscribing to 201:abcd:1234::1:8799...
> Sent SUBSCRIBE to 201:abcd:1234::1:8799
> Subscribed to 201:abcd:1234::1:8799
> Receiving audio from W1ALICE
> Received: packets=50, seq=49, from=W1ALICE
```

## Performance Characteristics

### Bandwidth Usage (per listener)

**Opus @ 64kbps, 20ms frames:**
- Audio data: ~64kbps
- Protocol overhead: ~3.2kbps (64-byte header per 20ms)
- Total: ~67kbps per listener

**For 10 listeners:** ~670kbps upload from broadcaster
**For 100 listeners:** ~6.7Mbps upload from broadcaster

### Latency

**End-to-end latency components:**
1. Audio capture: ~20ms (frame size)
2. Encoding: <5ms (Opus)
3. Network (Yggdrasil): Variable (depends on hops)
4. Decoding: <5ms
5. Audio output: ~20ms (buffer)

**Target:** <100ms end-to-end on local mesh

### Resource Usage

**Broadcaster (10 listeners):**
- Memory: <50MB
- CPU: <10% (1 core)
- Network: ~670kbps upload

**Listener:**
- Memory: <30MB
- CPU: <5%
- Network: ~67kbps download

## Testing Plan

### Manual Testing (2 nodes)

**Setup:**
```bash
# Node 1 (Broadcaster)
yggdrasil-go running
yggdrasilctl getSelf
# Note your IPv6: e.g., 201:abcd:1234::1

meshradio broadcast --callsign STATION1

# Node 2 (Listener)
yggdrasil-go running
meshradio listen 201:abcd:1234::1:8799 --callsign LISTENER1
```

**Expected behavior:**
1. Listener sends SUBSCRIBE
2. Broadcaster logs "New listener: LISTENER1"
3. Listener starts receiving audio
4. Heartbeats sent every 5s
5. If listener stops, broadcaster times out after 15s

### Multi-Listener Test (3+ nodes)

Same setup, multiple listeners tune to same broadcaster:
- Broadcaster shows N listeners connected
- Each listener receives audio independently
- Disconnecting one doesn't affect others

## Migration Notes

### Breaking Changes from v0.3

1. **Listener API changed:**
   ```go
   // Old (v0.3)
   listener.New(Config{TargetIPv6, TargetPort, LocalPort, ...})

   // New (v0.4)
   listener.New(Config{
       Callsign, LocalIPv6, LocalPort,  // Added
       TargetIPv6, TargetPort, ...
   })
   ```

2. **Broadcaster no longer uses multicast:**
   - No `multicastAddr` field
   - No `beaconLoop()`
   - New `listeners` map

3. **Discovery is manual:**
   - No automatic scanning (yet)
   - Share IPv6:port out-of-band

### Code Removed

- `Broadcaster.beaconLoop()`
- `Broadcaster.multicastAddr`
- `Broadcaster.multicastPort`
- Multicast-related logic

### Code Added

- `pkg/protocol/subscription.go` (new file)
- `PacketTypeSubscribe`, `PacketTypeHeartbeat` constants
- `Broadcaster.subscriptionLoop()`
- `Broadcaster.handleSubscribe()`
- `Broadcaster.handleHeartbeat()`
- `Broadcaster.heartbeatMonitor()`
- `Broadcaster.GetListenerCount()`
- `Broadcaster.GetListeners()`
- `Listener.subscribe()`
- `Listener.heartbeatLoop()`

## Next Steps (Post-MVP)

### Phase 1.5: Improve UX
- [ ] Show listener list in Web GUI
- [ ] Add listener statistics (packet loss, RTT)
- [ ] Better error messages

### Phase 2: Discovery
- [ ] Simple directory service (single node)
- [ ] Hardcoded bootstrap station list
- [ ] Station registration protocol

### Phase 3: Resilience
- [ ] Multiple directory nodes
- [ ] Gossip-based station discovery
- [ ] DHT implementation

### Phase 4: Optimization
- [ ] Adaptive bitrate based on network conditions
- [ ] Jitter buffer improvements
- [ ] Packet loss concealment

## References

- [Yggdrasil Discussion #991](https://github.com/yggdrasil-network/yggdrasil-go/discussions/991) - Multicast limitation
- [DESIGN.md](DESIGN.md) - Original design (needs update)
- [README.md](README.md) - Updated for v0.4

---

**Built:** 2025-11-29
**Compiles:** ✅
**Tested:** ⏳ (needs 2-node mesh test)
