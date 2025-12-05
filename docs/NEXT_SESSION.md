# Next Session Plan

**Date:** TBD
**Goal:** Implement Layer 3 (mDNS Discovery)
**Duration:** ~2-3 hours

---

## Session Objectives

### Primary Goal: mDNS Service Advertisement & Discovery

**What we'll build:**
- Advertise MeshRadio stations via mDNS
- Browse/discover stations on network
- Define TXT record format (emergency/regular channels)
- Test discovery on local network

**Success Criteria:**
- ✅ Broadcaster advertises via mDNS
- ✅ Listener discovers broadcaster automatically
- ✅ TXT records contain all metadata (group, channel, callsign, port)
- ✅ Works on local LAN

---

## Step-by-Step Plan

### Part 1: Library Integration (30 min)

**Task 1.1: Choose mDNS library**
```bash
# Add dependency
go get github.com/grandcat/zeroconf

# Why zeroconf?
# - Pure Go, cross-platform
# - Good documentation
# - Active maintenance
# - Works on Linux/macOS/Windows
```

**Task 1.2: Create mDNS package structure**
```
pkg/mdns/
  advertiser.go  - Advertise services
  browser.go     - Discover services
  txtrecord.go   - TXT record format helpers
```

### Part 2: Service Advertisement (45 min)

**Task 2.1: Define TXT record format**

Document standard format:
```yaml
_meshradio._udp.local.
TXT:
  group=emergency          # Multicast group label
  channel=emergency        # Channel type
  callsign=W1EMERGENCY     # Station identifier
  port=8790               # RTP port
  priority=critical       # normal/high/emergency/critical
  codec=opus              # Audio codec
  bitrate=64              # kbps
```

**Task 2.2: Implement Advertiser**
```go
pkg/mdns/advertiser.go

type Advertiser interface {
    Advertise(info ServiceInfo) error
    UpdateTXT(txt map[string]string) error
    Shutdown() error
}

// Start advertising when broadcast starts
// Withdraw when broadcast stops
```

**Task 2.3: Test advertisement**
```bash
# Terminal 1: Advertise
./mdns-test -mode advertise -callsign TEST1 -port 8790

# Terminal 2: Check with avahi-browse
avahi-browse -r _meshradio._udp

# Should see: TEST1 with TXT records
```

### Part 3: Service Discovery (45 min)

**Task 3.1: Implement Browser**
```go
pkg/mdns/browser.go

type Browser interface {
    Browse() ([]ServiceInfo, error)
    Subscribe(callback func(ServiceInfo)) error
    Stop() error
}

// Returns list of discovered stations
// Callback for real-time updates
```

**Task 3.2: Parse TXT records**
```go
pkg/mdns/txtrecord.go

// Helper functions
func ParseTXTRecord(txt []string) (ServiceInfo, error)
func CreateTXTRecord(info ServiceInfo) []string
```

**Task 3.3: Test discovery**
```bash
# Terminal 1: Advertise station
./mdns-test -mode advertise -callsign STATION1 -port 8790

# Terminal 2: Browse for stations
./mdns-test -mode browse

# Should print:
# Found: STATION1 at [fe80::1]:8790
# Group: emergency, Channel: emergency, Priority: critical
```

### Part 4: Integration Test (30 min)

**Task 4.1: Create mdns-test program**
```
cmd/mdns-test/
  main.go
    - advertise mode
    - browse mode
    - query mode (find specific callsign)
```

**Task 4.2: End-to-end test**
```
Test flow:
1. Start broadcaster (advertises via mDNS)
2. Start listener (discovers via mDNS)
3. Listener auto-tunes to discovered station
4. Verify RTP stream works
```

**Task 4.3: Test emergency channels**
```bash
# Advertise on emergency channel
./mdns-test -mode advertise -channel emergency -port 8790

# Browse for emergency stations only
./mdns-test -mode browse -filter emergency

# Should only show emergency stations
```

---

## Code to Write

### File 1: `pkg/mdns/advertiser.go`
```go
package mdns

import (
    "github.com/grandcat/zeroconf"
)

type Advertiser struct {
    server *zeroconf.Server
}

func NewAdvertiser(info ServiceInfo) (*Advertiser, error) {
    // Create mDNS server
    // Register _meshradio._udp service
    // Set TXT records
    return &Advertiser{...}, nil
}

func (a *Advertiser) Shutdown() error {
    // Clean shutdown
}
```

### File 2: `pkg/mdns/browser.go`
```go
package mdns

type Browser struct {
    resolver *zeroconf.Resolver
    services []ServiceInfo
}

func NewBrowser() (*Browser, error) {
    // Create mDNS resolver
    return &Browser{...}, nil
}

func (b *Browser) Browse() ([]ServiceInfo, error) {
    // Browse for _meshradio._udp services
    // Parse TXT records
    // Return discovered stations
}
```

### File 3: `pkg/mdns/types.go`
```go
package mdns

type ServiceInfo struct {
    Name     string
    Host     string
    Port     int
    IPv6     net.IP

    // TXT record fields
    Group    string
    Channel  string
    Callsign string
    Priority string
    Codec    string
    Bitrate  int
}
```

### File 4: `cmd/mdns-test/main.go`
```go
package main

// Modes:
// - advertise: Advertise a test station
// - browse: Discover stations
// - query: Find specific callsign
```

---

## Testing Plan

### Test 1: Local LAN Discovery
```
Machine A & B on same WiFi/Ethernet:
1. Machine A advertises
2. Machine B browses
3. Should discover in <5 seconds
```

### Test 2: Multiple Stations
```
Start 3 advertisers:
- STATION1 (emergency, port 8790)
- STATION2 (community, port 8799)
- STATION3 (emergency, port 8790)

Browser should find all 3
Filter by emergency → should show STATION1, STATION3
```

### Test 3: Emergency Priority
```
Advertise with priority=critical
Browser filters for priority >= emergency
Should auto-tune (in future integration)
```

### Test 4: Service Updates
```
Advertise → update TXT record → verify browser sees update
Advertise → shutdown → verify browser removes entry
```

---

## Potential Issues & Solutions

### Issue 1: mDNS doesn't work across Yggdrasil mesh
**Problem:** mDNS is link-local (broadcast-based)
**Solution:**
- Phase 1: Works on LAN (good enough for MVP)
- Phase 2: Add mDNS relay/proxy for mesh segments
- Phase 3: Fallback to manual entry (emergency ports)

### Issue 2: TXT record size limits
**Problem:** Too much metadata in TXT records
**Solution:** Keep essential fields only, fetch details via API

### Issue 3: Firewall blocks mDNS
**Problem:** Port 5353 blocked
**Solution:** Document firewall rules, fallback to manual

### Issue 4: Multiple network interfaces
**Problem:** mDNS announces on wrong interface
**Solution:** Allow interface selection in config

---

## Documentation to Update

### Update TODO.md
- [x] Layer 2: RTP Streaming (completed)
- [ ] Layer 3.1: mDNS library integration
- [ ] Layer 3.2: Service advertisement
- [ ] Layer 3.3: Service discovery
- [ ] Layer 3.4: TXT record format

### Create MDNS_SPEC.md
Document TXT record convention for community feedback:
```
# MeshRadio mDNS/Avahi Service Specification

Service Type: _meshradio._udp.local.

Required TXT fields: ...
Optional TXT fields: ...
Examples: ...
```

### Update RTP_TEST.md
Add section: "Discovery via mDNS (coming soon)"

---

## After This Session

**What we'll have:**
- ✅ Automatic station discovery
- ✅ No manual IPv6 entry needed
- ✅ Emergency channel filtering
- ✅ Real-time station updates

**What's next:**
- Layer 4: Multicast Overlay (SSM/Regular)
- Integration: mDNS → RTP streaming
- Emergency features: Priority, auto-tune

---

## Quick Start (Next Session)

```bash
# 1. Add dependency
go get github.com/grandcat/zeroconf

# 2. Create package structure
mkdir -p pkg/mdns
touch pkg/mdns/{advertiser,browser,types,txtrecord}.go

# 3. Create test program
mkdir -p cmd/mdns-test
touch cmd/mdns-test/main.go

# 4. Start coding!
```

---

## Time Estimate

- Library integration: 30 min
- Advertiser implementation: 45 min
- Browser implementation: 45 min
- Test program: 30 min
- Testing & debugging: 30 min

**Total: ~3 hours**

---

## Success Metrics

By end of next session:
- [ ] `./mdns-test -mode advertise` works
- [ ] `./mdns-test -mode browse` finds stations
- [ ] TXT records parsed correctly
- [ ] Emergency channels filterable
- [ ] Works on local LAN

---

## Open Questions for Community (Post After)

After Layer 3 complete, post to GitHub:

**"mDNS/Avahi Discovery Implementation - Feedback Wanted"**

1. TXT record format - is this standard enough?
2. mDNS over Yggdrasil mesh - relay approach viable?
3. Emergency channel filtering - should be client-side or advertiser-side?

---

**Ready to go!** 🚀

See you next session for Layer 3: mDNS Discovery!
