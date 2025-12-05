# MeshRadio Design Document

**Version:** 1.0
**Date:** 2025-11-25
**Status:** Foundation Design

## 1. Executive Summary

MeshRadio is a decentralized radio broadcasting system built on top of the Yggdrasil network. It provides HAM radio-like functionality (broadcasting, scanning, calling) using Yggdrasil IPv6 addresses as "frequencies" instead of traditional radio spectrum.

### Key Innovation
Instead of simulating radio frequencies, MeshRadio leverages the Yggdrasil IPv6 address space as a vast "frequency spectrum" where:
- **IPv6 addresses = Tunable frequencies**
- **IPv6 subnets = Radio bands**
- **Scanning = IPv6 range exploration**
- **Signal strength = Yggdrasil routing metrics**

## 2. Core Concepts

### 2.1 IPv6 as Frequency Spectrum

Traditional radio uses electromagnetic frequencies (e.g., 146.520 MHz). MeshRadio uses IPv6 addresses as virtual frequencies:

```
Traditional Radio          →    MeshRadio
─────────────────                ──────────
146.520 MHz (frequency)    →    202:1234:5678:abcd::1 (IPv6)
144-148 MHz (2m band)      →    202::/8 (talk radio band)
Frequency scanning         →    IPv6 subnet scanning
Signal strength (RSSI)     →    Hop count + RTT + bandwidth
```

### 2.2 Station Types

| Type | Description | IPv6 Range | Use Case |
|------|-------------|------------|----------|
| Broadcaster | One-to-many transmitter | 200::/7 | Radio stations, podcasts |
| Repeater | Store-and-forward relay | 300::/8 | Extend network reach |
| Mobile | Portable station | Any | Portable operators |
| Net Control | Organized net coordinator | Designated IPv6 | Scheduled nets, events |
| Beacon | Announcement-only | Any | Status, position updates |

### 2.3 Communication Modes

#### Simplex (Direct)
```
Station A (202:aaaa::1) ←──────→ Station B (202:bbbb::1)
```
Direct point-to-point communication.

#### Broadcast (One-to-Many)
```
                    ┌──→ Listener 1 (202:1111::1)
Broadcaster ────────┼──→ Listener 2 (202:2222::1)
(202:station::1)    └──→ Listener 3 (202:3333::1)
```
Station broadcasts to multiple listeners via multicast or multiple streams.

#### Repeater (Store-and-Forward)
```
Station A → Repeater (300:rep::1) → Station B
            (stores, re-broadcasts)
```
Extended range through intermediate nodes.

#### Multicast (Group)
```
Multiple Stations → ff05::200:talk ← Multiple Listeners
```
Group communication via IPv6 multicast addresses.

## 3. Architecture

### 3.1 System Layers

```
┌───────────────────────────────────────────────────────────┐
│              Application Layer                            │
│  - Broadcasting   - Scanning   - Calling   - Recording    │
├───────────────────────────────────────────────────────────┤
│              Protocol Layer                               │
│  - Station Beacons    - Discovery Protocol                │
│  - Audio Streaming    - Metadata Exchange                 │
│  - DHT Registry       - Signal Quality Monitoring         │
├───────────────────────────────────────────────────────────┤
│              Audio Layer                                  │
│  - Opus Codec (voice)     - Input/Output (PortAudio)      │
│  - FLAC Codec (music)     - Buffering & Jitter Control    │
├───────────────────────────────────────────────────────────┤
│              Transport Layer                              │
│  - UDP (audio streaming)  - TCP (control/metadata)        │
│  - Multicast Support      - Flow Control                  │
├───────────────────────────────────────────────────────────┤
│              Yggdrasil Network Layer                      │
│  - IPv6 Addressing        - Mesh Routing                  │
│  - Encrypted Transport    - NAT Traversal                 │
└───────────────────────────────────────────────────────────┘
```

### 3.2 Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        MeshRadio Core                       │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Broadcaster │  │   Scanner    │  │    Caller    │     │
│  │   Module     │  │    Module    │  │    Module    │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                  │                  │             │
│         └──────────────────┼──────────────────┘             │
│                            │                                │
│                    ┌───────┴────────┐                       │
│                    │  Protocol Core │                       │
│                    │   - Beacons    │                       │
│                    │   - Discovery  │                       │
│                    │   - Registry   │                       │
│                    └───────┬────────┘                       │
│                            │                                │
│         ┌──────────────────┼──────────────────┐             │
│         │                  │                  │             │
│  ┌──────┴───────┐  ┌───────┴──────┐  ┌───────┴──────┐     │
│  │    Audio     │  │   Network    │  │   Storage    │     │
│  │   Processor  │  │   Manager    │  │   Manager    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 4. IPv6 Band Plan

### 4.1 Frequency Allocation

```yaml
bands:
  music:
    range: "200::/7"
    description: "Music broadcasting stations"
    subbands:
      - range: "200::/8"
        type: "General music"
      - range: "201::/8"
        type: "Live performances"

  talk:
    range: "202::/8"
    description: "Talk radio, podcasts, news"
    subbands:
      - range: "202:0::/16"
        type: "News broadcasts"
      - range: "202:1::/16"
        type: "Talk shows"
      - range: "202:f::/16"
        type: "Educational content"

  emergency:
    range: "203::/8"
    description: "Emergency and priority communications"
    priority: high
    monitoring: mandatory

  amateur:
    range: "204::/8"
    description: "Amateur/HAM equivalent - experimental"
    subbands:
      - range: "204:0::/16"
        type: "Voice (Simplex)"
      - range: "204:1::/16"
        type: "Digital modes"
      - range: "204:2::/16"
        type: "Repeater outputs"

  repeaters:
    range: "300::/8"
    description: "Repeater and relay stations"

  calling:
    multicast: "ff05::/16"
    description: "Calling and group channels"
    channels:
      - address: "ff05::200:cq"
        description: "General CQ calling"
      - address: "ff05::200:emerg"
        description: "Emergency calling"
      - address: "ff05::200:net"
        description: "Net control"
```

### 4.2 Special Addresses

| Address | Purpose | Usage |
|---------|---------|-------|
| `ff02::1` | All nodes on local link | Local discovery |
| `ff05::200:cq` | General calling (CQ) | "Calling any station" |
| `ff05::200:emerg` | Emergency calling | Priority communications |
| `ff05::200:beacon` | Beacon channel | Station announcements |
| `ff08::200:*` | Global multicast groups | Topic-based groups |

## 5. Protocol Specification

### 5.1 Packet Format

All MeshRadio packets follow this base format:

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|Version| Type  |     Flags     |         Payload Length        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
+                         Timestamp (64-bit)                    +
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
+                                                               +
|                                                               |
+                    Source IPv6 (128-bit)                      +
|                                                               |
+                                                               +
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       Callsign (16 bytes)                     |
+                                                               +
|                                                               |
+                                                               +
|                                                               |
+                                                               +
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| Sequence Num  | Signal Quality|   Reserved    |               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+               +
|                                                               |
+                         Payload Data                          +
|                         (variable length)                     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

#### Field Descriptions

- **Version** (4 bits): Protocol version (currently 0x1)
- **Type** (4 bits): Packet type (see below)
- **Flags** (8 bits): Various flags (encryption, priority, etc.)
- **Payload Length** (16 bits): Length of payload in bytes
- **Timestamp** (64 bits): Unix timestamp in milliseconds
- **Source IPv6** (128 bits): Sender's Yggdrasil IPv6
- **Callsign** (128 bits): UTF-8 encoded station callsign
- **Sequence Number** (8 bits): Packet sequence for ordering
- **Signal Quality** (8 bits): 0-255 quality metric
- **Reserved** (8 bits): Reserved for future use
- **Payload Data**: Type-specific payload

### 5.2 Packet Types

| Type | Value | Name | Description |
|------|-------|------|-------------|
| 0x00 | 0 | BEACON | Station announcement/heartbeat |
| 0x01 | 1 | AUDIO | Audio stream packet |
| 0x02 | 2 | METADATA | Station metadata update |
| 0x03 | 3 | CALL_CQ | General calling (CQ) |
| 0x04 | 4 | CALL_SELECTIVE | Selective calling |
| 0x05 | 5 | DISCOVERY_REQ | Discovery request |
| 0x06 | 6 | DISCOVERY_RESP | Discovery response |
| 0x07 | 7 | REGISTRY_QUERY | DHT registry query |
| 0x08 | 8 | REGISTRY_RESPONSE | DHT registry response |
| 0x09 | 9 | SIGNAL_REPORT | Signal quality report |
| 0x0A | 10 | EMERGENCY | Emergency broadcast |

### 5.3 Beacon Packet Payload

```json
{
  "station_type": "broadcaster|repeater|mobile",
  "station_name": "Tech Talk Radio",
  "operating_bands": [
    "202:1234::/32"
  ],
  "modes": ["audio_stream", "voice"],
  "power_level": 128,
  "location": {
    "latitude": 40.7128,
    "longitude": -74.0060,
    "grid_square": "FN30as"
  },
  "metadata": {
    "description": "Technology news and discussions",
    "language": "en",
    "codec": "opus",
    "bitrate": 64000
  },
  "uptime": 86400,
  "beacon_interval": 300
}
```

### 5.4 Audio Stream Packet Payload

```
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| Codec Type    | Sample Rate   |   Channels    |   Bitrate     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       Frame Timestamp                         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
+                    Encoded Audio Data                         +
|                       (variable length)                       |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

Codec Types:
- 0x01: Opus (recommended for voice)
- 0x02: FLAC (for music/high quality)
- 0x03: AAC
- 0x04: MP3

## 6. Bootstrap Strategy

### 6.1 The Bootstrap Problem

MeshRadio faces a two-layered bootstrap challenge:

1. **Yggdrasil Layer**: Connecting to the Yggdrasil mesh network
2. **MeshRadio Layer**: Discovering the first radio station to listen to

Without solving both layers, a new user cannot participate in the network.

### 6.2 Yggdrasil Bootstrap Solutions

Yggdrasil itself has established bootstrap mechanisms that MeshRadio inherits:

#### A. Public Peers
```yaml
# Yggdrasil config includes public peer list
peers:
  - tls://[2001:db8::1]:12345
  - tcp://ygg.example.com:54321
  - quic://bootstrap.yggdrasil.network:9001
```

MeshRadio doesn't need to solve this - users follow standard Yggdrasil setup.

#### B. Local Network Discovery
Yggdrasil supports local peer discovery via multicast:
- Automatically finds peers on same LAN
- No configuration needed
- Ideal for local/community deployments

### 6.3 MeshRadio Station Bootstrap Solutions

Once connected to Yggdrasil, finding the first MeshRadio station:

#### Strategy 1: Well-Known Station Registry (Recommended)

**Concept**: Maintain a curated list of stable, always-on stations.

```yaml
# Built into MeshRadio binary or config
bootstrap_stations:
  - ipv6: "202:aaaa:bbbb:cccc::1"
    callsign: "W1BOOT"
    name: "MeshRadio Bootstrap Station"
    type: "directory"
    band: "talk"
    reliability: "high"

  - ipv6: "202:1111:2222:3333::1"
    callsign: "W2NEWS"
    name: "News Network"
    type: "broadcaster"
    band: "talk"

  - ipv6: "200:5555:6666:7777::1"
    callsign: "W3MUSIC"
    name: "Music Bootstrap"
    type: "broadcaster"
    band: "music"
```

**Implementation:**
```go
// Embedded in binary
const DefaultBootstrapStations = `
# Auto-updated on build
# Generated: 2025-11-25
bootstrap_stations:
  - ipv6: "202:aaaa:bbbb:cccc::1"
    callsign: "W1BOOT"
    # ...
`

// User can override
func LoadBootstrapStations() []Station {
    // Try user config first
    if stations := loadUserBootstrap(); len(stations) > 0 {
        return stations
    }
    // Fall back to embedded defaults
    return parseDefaultBootstrap()
}
```

#### Strategy 2: DNS-Style Directory Service

**Concept**: Use a well-known IPv6 address as a "station directory service"

```
Directory Service at: 202:0000:0000:0000::1 (reserved)

Query Format:
  GET /stations?band=talk&limit=10

Response:
  {
    "stations": [
      {
        "ipv6": "202:1234:5678::1",
        "callsign": "W1AW",
        "name": "Tech Talk",
        "last_seen": "2025-11-25T14:00:00Z",
        "quality": 245
      }
    ],
    "total": 150
  }
```

**Advantages:**
- Dynamic station list
- Real-time availability
- Can filter by band/type

**Challenges:**
- Directory service becomes central point
- Needs to be highly available
- Could be distributed (multiple directories)

#### Strategy 3: DHT Seeding

**Concept**: Bootstrap DHT with known nodes, then discover via DHT queries.

```go
// Bootstrap DHT nodes (stable, long-running)
dhtBootstrapNodes := []string{
    "202:aaaa:bbbb:cccc::1",  // DHT bootstrap 1
    "202:dddd:eeee:ffff::1",  // DHT bootstrap 2
    "202:1111:2222:3333::1",  // DHT bootstrap 3
}

// Join DHT network
dht := NewDHT()
for _, node := range dhtBootstrapNodes {
    dht.Bootstrap(node)
}

// Query for stations
stations := dht.Query("meshradio:stations:talk")
```

**DHT contains:**
- `meshradio:stations:<band>` → List of active stations
- `meshradio:callsign:<callsign>` → IPv6 lookup
- `meshradio:directory` → Full station registry

#### Strategy 4: Multicast Discovery (LAN-first)

**Concept**: For local deployments, use IPv6 multicast to discover nearby stations.

```go
// Send discovery beacon on multicast
multicastAddr := "ff02::1:meshradio"  // Link-local multicast

// All MeshRadio stations listen on this address
beacon := DiscoveryBeacon{
    Type: "discovery_request",
    Seeking: "any_station",
}

sendMulticast(multicastAddr, beacon)

// Nearby stations respond
responses := listenForResponses(5 * time.Second)
```

**Best for:**
- Community mesh networks
- Local radio clubs
- LAN parties / events

#### Strategy 5: Social Bootstrap (Out-of-band)

**Concept**: Share station addresses through external channels.

```
Methods:
- QR codes at meetups
- Website listings (meshradio.network/stations)
- Social media posts
- IRC/Discord channels
- Email lists
- Printed directories
```

Example:
```
Join MeshRadio Tech Talk:
meshradio dial 202:1234:5678:abcd::1

Or scan QR code:
[QR CODE containing meshradio:// URI]
```

#### Strategy 6: Hybrid Approach (RECOMMENDED)

Combine multiple strategies for resilience:

```go
func BootstrapMeshRadio() []Station {
    stations := []Station{}

    // 1. Try built-in bootstrap list first (fastest)
    if bootstrapStations := loadEmbeddedBootstrap(); len(bootstrapStations) > 0 {
        log.Info("Trying embedded bootstrap stations...")
        stations = append(stations, verifyStations(bootstrapStations)...)
    }

    // 2. Try multicast discovery (LAN)
    if localStations := discoverMulticast(5 * time.Second); len(localStations) > 0 {
        log.Info("Found local stations via multicast")
        stations = append(stations, localStations...)
    }

    // 3. Try DHT bootstrap
    if dhtStations := bootstrapDHT(); len(dhtStations) > 0 {
        log.Info("Found stations via DHT")
        stations = append(stations, dhtStations...)
    }

    // 4. Try directory service (if configured)
    if dirStations := queryDirectory(); len(dirStations) > 0 {
        log.Info("Found stations via directory")
        stations = append(stations, dirStations...)
    }

    // 5. Check user bookmarks
    if bookmarks := loadBookmarks(); len(bookmarks) > 0 {
        log.Info("Using saved bookmarks")
        stations = append(stations, bookmarks...)
    }

    // 6. Fallback: prompt user for manual entry
    if len(stations) == 0 {
        log.Warn("No stations found via bootstrap. Manual configuration required.")
        return promptManualStation()
    }

    return deduplicateAndRank(stations)
}
```

### 6.4 Bootstrap Station Requirements

For a station to serve as a reliable bootstrap node:

```yaml
bootstrap_requirements:
  availability:
    uptime: ">99%"
    monitoring: "automated health checks"

  stability:
    yggdrasil_version: "stable release"
    ipv6_stability: "static address preferred"
    bandwidth: ">1 Mbps sustained"

  content:
    type: "directory_service OR broadcaster"
    beacon_interval: "60 seconds"

  governance:
    operator: "known community member"
    contact: "published contact info"
    backup: "redundant instances"
```

### 6.5 Decentralization Strategy

To avoid single points of failure:

#### Geographic Distribution
```
Bootstrap stations distributed globally:
- North America: 202:1000::/20 (3-5 stations)
- Europe: 202:2000::/20 (3-5 stations)
- Asia: 202:3000::/20 (3-5 stations)
- Oceania: 202:4000::/20 (2-3 stations)
- South America: 202:5000::/20 (2-3 stations)
```

#### Rotating Bootstrap List
```go
// Update bootstrap list from multiple sources
func UpdateBootstrapList() {
    sources := []string{
        "https://bootstrap.meshradio.network/stations.yaml",
        "https://mirror1.meshradio.org/stations.yaml",
        "https://mirror2.meshradio.org/stations.yaml",
    }

    // Fetch from multiple sources, merge results
    merged := fetchAndMerge(sources)

    // Verify cryptographic signatures
    verified := verifySignatures(merged)

    // Update local bootstrap cache
    saveBootstrapCache(verified)
}
```

#### Community-Run Bootstrap Nodes

Encourage community to run bootstrap stations:

```bash
# Easy bootstrap station setup
meshradio bootstrap-station \
  --callsign W1COMMUNITY \
  --type directory \
  --register  # Registers with community registry

# Station automatically:
# - Announces itself via DHT
# - Provides directory service
# - Responds to discovery requests
# - Submits health reports
```

### 6.6 First-Run Experience

```
$ meshradio init

Welcome to MeshRadio!

Checking Yggdrasil connection... ✓ Connected
Your IPv6: 200:1234:5678:9abc::1

Bootstrapping MeshRadio network...
[1/5] Loading embedded bootstrap stations... ✓ (3 stations)
[2/5] Scanning local network... ✓ (1 station found)
[3/5] Connecting to DHT... ✓ (Bootstrap successful)
[4/5] Querying station directory... ✓ (47 stations available)
[5/5] Building station database... ✓

Found 51 active stations!

Top stations by signal quality:
  1. [202:aaaa:1111::1] W1NEWS - Tech News Network (S9+)
  2. [200:bbbb:2222::1] W2JAZZ - Jazz Radio (S9)
  3. [202:cccc:3333::1] W3TALK - Community Talk (S8)

Would you like to:
  [L] Listen to top station
  [S] Scan all bands
  [C] Configure station
  [Q] Quit

Choice:
```

### 6.7 Offline Bootstrap

For air-gapped or isolated networks:

```bash
# Export bootstrap data from connected network
meshradio export-bootstrap --output bootstrap.yaml

# Transfer file to isolated network (USB, etc.)

# Import on isolated network
meshradio import-bootstrap --input bootstrap.yaml

# This provides initial station list without internet
```

### 6.8 Bootstrap Protocol Extension

Add bootstrap-specific packet type:

```go
type BootstrapRequest struct {
    Version      uint8
    RequestType  uint8  // 0=any, 1=directory, 2=broadcaster
    BandFilter   string // Optional band preference
    MaxResults   uint8
}

type BootstrapResponse struct {
    Stations     []StationInfo
    TTL          uint32  // How long this info is valid
    NextUpdate   int64   // When to refresh
    MoreAvailable bool
}
```

## 7. Discovery Protocol

### 7.1 Station Discovery Flow

```
Listener                          Network                    Broadcaster
   |                                 |                            |
   |------- Discovery Request ------>|                            |
   |     (ff05::200:beacon)          |                            |
   |                                 |                            |
   |                                 |<----- Beacon Packet -------|
   |                                 |  (periodic announcement)   |
   |                                 |                            |
   |<------ Beacon Packet -----------|                            |
   |                                 |                            |
   |------- Direct Connect --------------------------------->|    |
   |     (to broadcaster IPv6)       |                            |
   |                                 |                            |
   |<====== Audio Stream ===================================|    |
   |                                 |                            |
```

### 6.2 Discovery Methods

#### Passive Discovery
Listen for beacon packets on multicast channels:
- Stations broadcast beacons every 5 minutes (configurable)
- Listeners collect beacon data to build station database
- No active probing required

#### Active Discovery
Send discovery requests to find stations:
```
1. Send DISCOVERY_REQ to ff05::200:beacon
2. Stations respond with DISCOVERY_RESP containing metadata
3. Build station list from responses
4. Query DHT for additional stations
```

#### DHT Registry
Distributed hash table for callsign ↔ IPv6 resolution:
```
1. Hash callsign using SHA-256
2. Find closest nodes in DHT
3. Store/retrieve IPv6 address
4. TTL-based expiration (default: 24 hours)
```

### 6.3 Scanner Operation

```python
# Pseudocode for scanning algorithm

def scan_band(ipv6_range, scan_mode):
    """
    Scan IPv6 range for active stations
    """
    results = []

    if scan_mode == "passive":
        # Listen for beacons
        listen_multicast("ff05::200:beacon", timeout=30)

    elif scan_mode == "active":
        # Probe range
        for ipv6 in generate_scan_targets(ipv6_range):
            if probe_station(ipv6):
                results.append(get_station_info(ipv6))

    elif scan_mode == "smart":
        # Use routing table + DHT
        nearby = get_yggdrasil_neighbors()
        for node in nearby:
            if is_meshradio_station(node):
                results.append(node)

        # Query DHT
        dht_stations = query_dht_for_band(ipv6_range)
        results.extend(dht_stations)

    return results
```

### 6.4 Scan Modes

| Mode | Description | Speed | Completeness | Network Impact |
|------|-------------|-------|--------------|----------------|
| Passive | Listen only | Slow | Low | None |
| Active | Probe targets | Medium | Medium | Low |
| Smart | Use DHT + routing table | Fast | High | Low |
| Sequential | Step through range | Very Slow | Complete | High |
| Random | Random walk | Medium | Medium | Low |

## 7. Signal Quality Metrics

### 7.1 Quality Calculation

Signal quality is derived from Yggdrasil routing metrics:

```go
type SignalQuality struct {
    HopCount      int       // Number of hops in route
    RTT           int64     // Round-trip time (ms)
    Bandwidth     int64     // Available bandwidth (bps)
    PacketLoss    float64   // Packet loss percentage
    Jitter        int64     // Jitter (ms)
    QualityScore  uint8     // Overall quality (0-255)
}

func CalculateQuality(metrics SignalQuality) uint8 {
    // Weighted scoring algorithm
    hopScore := max(0, 255 - (metrics.HopCount * 20))
    rttScore := max(0, 255 - (metrics.RTT / 10))
    lossScore := max(0, 255 - (metrics.PacketLoss * 255))

    quality := (hopScore * 0.3) + (rttScore * 0.3) + (lossScore * 0.4)
    return uint8(quality)
}
```

### 7.2 Quality Reporting

Stations periodically exchange SIGNAL_REPORT packets:
- Reported every 30 seconds during active connection
- Used to optimize routing and stream quality
- Triggers quality degradation warnings

Quality Levels:
- **250-255**: Excellent (S9+)
- **200-249**: Good (S7-S9)
- **150-199**: Fair (S5-S7)
- **100-149**: Poor (S3-S5)
- **0-99**: Very Poor (S1-S3)

## 8. Calling System

### 8.1 CQ Calling (General)

```
Operator                    Calling Channel                  Network
   |                       (ff05::200:cq)                       |
   |                                                            |
   |------- CALL_CQ Packet ------>|                            |
   |   {                          |                            |
   |     "callsign": "W1AW",      |                            |
   |     "mode": "voice",         |                            |
   |     "seeking": "any"         |                            |
   |   }                          |                            |
   |                              |--- Broadcast to all ------>|
   |                              |                            |
   |<------------------------- Response (if interested) -------|
   |                              |                            |
   |------- Direct QSO --------------------------------->|      |
```

### 8.2 Selective Calling

Direct calling to specific station or callsign:

```json
{
  "type": "CALL_SELECTIVE",
  "from_callsign": "W1AW",
  "to_callsign": "W2XYZ",
  "to_ipv6": "202:5678:abcd::1",
  "ctcss_token": "optional_access_token",
  "message": "Request QSO on 204:1234::1"
}
```

### 8.3 Net Control

Organized nets with designated controller:

```yaml
net:
  name: "Tech Talk Tuesday Net"
  control_station: "202:aaaa:bbbb::1"
  control_callsign: "W1AW"
  frequency: "204:net:tuesday::1"
  schedule:
    - day: "tuesday"
      time: "20:00 UTC"
      duration: 3600
  check_in_procedure:
    - "Wait for net control to open"
    - "State callsign when prompted"
    - "Wait for acknowledgment"
```

## 9. Security Considerations

### 9.1 Yggdrasil Native Security

MeshRadio inherits Yggdrasil's security features:
- **End-to-end encryption**: All traffic encrypted by Yggdrasil
- **Authenticated routing**: Prevents route hijacking
- **IPv6 ownership**: Cryptographically bound addresses

### 9.2 Application-Level Security

Additional security measures:

#### Access Control
```yaml
security:
  callsign_verification:
    enabled: true
    method: "pgp_signature"  # Sign beacons with PGP key

  station_whitelist:
    enabled: false
    allowed_callsigns: []

  station_blacklist:
    enabled: true
    blocked_ipv6: []
    blocked_callsigns: []
```

#### Content Authentication
- Stations sign beacons with cryptographic keys
- Callsign → Public Key mapping in DHT
- Verify beacon signatures before trusting

#### Rate Limiting
```yaml
rate_limiting:
  max_beacons_per_minute: 10
  max_discovery_requests: 100
  max_audio_streams: 5
```

### 9.3 Abuse Prevention

- **Spam filtering**: Rate limit beacon broadcasts
- **Quality-based filtering**: Ignore low-quality/malicious stations
- **Reputation system**: Track station behavior
- **Emergency override**: Priority for emergency broadcasts

## 10. Data Formats

### 10.1 Configuration Files

#### Main Configuration (meshradio.yaml)
```yaml
version: "1.0"

station:
  callsign: "W1AW"
  operator_name: "John Doe"
  location:
    latitude: 40.7128
    longitude: -74.0060
    grid_square: "FN30as"

yggdrasil:
  daemon_address: "localhost:9001"
  admin_socket: "/var/run/yggdrasil.sock"
  interface: "tun0"

broadcasting:
  enabled: true
  ipv6: "202:1234:5678:abcd::1"
  codec: "opus"
  bitrate: 64000
  sample_rate: 48000
  channels: 2
  beacon_interval: 300  # seconds

scanning:
  enabled: true
  bands:
    - "202::/8"   # Talk radio
    - "204::/8"   # Amateur
  mode: "smart"
  scan_interval: 10  # seconds

discovery:
  dht_enabled: true
  passive_listen: true
  beacon_multicast: "ff05::200:beacon"

audio:
  input_device: "default"
  output_device: "default"
  buffer_size: 4096
  latency: "low"

logging:
  level: "info"
  file: "/var/log/meshradio/meshradio.log"
```

#### Bookmarks (bookmarks.yaml)
```yaml
bookmarks:
  - name: "Tech News Network"
    ipv6: "202:aaaa:1111::1"
    callsign: "W1NEWS"
    band: "talk"
    tags: ["news", "technology"]
    added: "2025-11-25T14:00:00Z"

  - name: "Jazz Radio"
    ipv6: "200:bbbb:2222::1"
    callsign: "W2JAZZ"
    band: "music"
    tags: ["music", "jazz"]
    added: "2025-11-20T10:30:00Z"
```

#### Band Plan (bands.yaml)
```yaml
# See section 4.1 for full band plan specification
```

### 10.2 Station Database Schema

```sql
-- SQLite schema for local station database

CREATE TABLE stations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ipv6 TEXT UNIQUE NOT NULL,
    callsign TEXT NOT NULL,
    station_name TEXT,
    station_type TEXT,  -- broadcaster, repeater, mobile
    last_seen INTEGER,  -- Unix timestamp
    signal_quality INTEGER,
    hop_count INTEGER,
    rtt INTEGER,
    metadata JSON,
    created_at INTEGER,
    updated_at INTEGER
);

CREATE TABLE beacons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    station_id INTEGER,
    timestamp INTEGER,
    beacon_data JSON,
    FOREIGN KEY (station_id) REFERENCES stations(id)
);

CREATE TABLE scan_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    band TEXT,
    scan_mode TEXT,
    started_at INTEGER,
    completed_at INTEGER,
    stations_found INTEGER,
    results JSON
);

CREATE TABLE contacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    station_id INTEGER,
    contact_time INTEGER,
    duration INTEGER,
    mode TEXT,  -- voice, digital, etc.
    notes TEXT,
    FOREIGN KEY (station_id) REFERENCES stations(id)
);

CREATE INDEX idx_stations_ipv6 ON stations(ipv6);
CREATE INDEX idx_stations_callsign ON stations(callsign);
CREATE INDEX idx_stations_last_seen ON stations(last_seen);
CREATE INDEX idx_beacons_timestamp ON beacons(timestamp);
```

## 11. User Interface

### 11.1 CLI Commands

```bash
# Broadcasting
meshradio broadcast --callsign W1AW --ipv6 202:1234::1 --input mic
meshradio broadcast --file music.flac --loop

# Listening
meshradio listen --ipv6 202:1234::1
meshradio dial 202:1234::1  # Alias for listen

# Scanning
meshradio scan --band talk                    # Scan talk radio band
meshradio scan --range 202::/16               # Scan specific range
meshradio scan --mode smart --duration 60     # Smart scan for 60s
meshradio scan --auto                         # Auto-scan all bands

# Discovery
meshradio discover --local                    # Discover local stations
meshradio discover --dht --callsign W2XYZ    # Lookup in DHT

# Calling
meshradio call cq --callsign W1AW             # CQ call
meshradio call --to W2XYZ --message "QSO?"    # Selective call

# Station Management
meshradio bookmark add --ipv6 202:1234::1 --name "News"
meshradio bookmark list
meshradio bookmark remove --name "News"

# Information
meshradio info --ipv6 202:1234::1             # Station info
meshradio status                              # Local status
meshradio bands                               # Show band plan

# Interactive Mode
meshradio tui                                 # Launch TUI
meshradio daemon                              # Run as daemon
```

### 11.2 TUI (Terminal User Interface)

```
┌─ MeshRadio v1.0 ──────────────────────────────────────────────────────────┐
│ Station: W1AW | Mode: Listening | Yggdrasil: Connected                    │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│ ┌─ Current Station ─────────────────────────────────────────────────────┐ │
│ │ Frequency: 202:1234:5678:abcd::1                                      │ │
│ │ Callsign:  W2NEWS                                                     │ │
│ │ Name:      Tech News Network                                          │ │
│ │ Signal:    ▓▓▓▓▓▓▓▓░░ (8/10) | Hops: 3 | RTT: 45ms                  │ │
│ │ Status:    ● BROADCASTING                                             │ │
│ │                                                                        │ │
│ │ Volume:    ▓▓▓▓▓▓▓░░░░ 70%    [▲/▼ to adjust]                        │ │
│ └────────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│ ┌─ Memory Banks ─────────────┐  ┌─ Recent Activity ──────────────────┐   │
│ │ 1. Tech News - 202:aaaa::1 │  │ 14:23 W3XYZ calling on CQ        │   │
│ │ 2. Jazz Radio - 200:bbbb::1│  │ 14:15 202:9876::1 beacon rcvd    │   │
│ │ 3. Emergency - 203:eeee::1 │  │ 14:10 Scan complete: 15 stations │   │
│ │ 4. Net Tues - 204:net::1   │  │ 14:05 Connected to W2NEWS        │   │
│ │                            │  │ 14:00 Station started            │   │
│ └────────────────────────────┘  └──────────────────────────────────────┘   │
│                                                                            │
│ ┌─ Spectrum Activity (202::/8 - Talk Radio Band) ──────────────────────┐  │
│ │                                                                        │  │
│ │  202:0000::/16  ░░░▓▓░░░░░░░░░░░  3 stations                         │  │
│ │  202:1000::/16  ▓▓▓▓▓▓▓▓▓▓░░░░░░  8 stations  ← You are here         │  │
│ │  202:2000::/16  ░░░░░░░░░░░░░░░░  0 stations                         │  │
│ │  202:f000::/16  ░░▓▓░░░░░░░░░░░░  2 stations                         │  │
│ │                                                                        │  │
│ └────────────────────────────────────────────────────────────────────────┘  │
│                                                                            │
│ [S]can [D]ial [C]all [B]ookmarks [M]onitor [R]ecord [I]nfo [Q]uit        │
└────────────────────────────────────────────────────────────────────────────┘
```

### 11.3 Web Interface (Future)

Optional web dashboard for station management:
- Real-time spectrum visualization
- Station database browser
- Configuration management
- Audio streaming player
- Contact log management

## 12. Implementation Roadmap

### Phase 1: Foundation (v0.1)
- [ ] Core protocol implementation
- [ ] Basic packet encoding/decoding
- [ ] Yggdrasil integration
- [ ] Simple broadcaster
- [ ] Simple listener
- [ ] CLI interface

### Phase 2: Discovery (v0.2)
- [ ] Beacon system
- [ ] Passive discovery
- [ ] Active scanning
- [ ] Station database
- [ ] Band plan configuration

### Phase 3: Audio (v0.3)
- [ ] Opus codec integration
- [ ] Audio input/output (PortAudio)
- [ ] Streaming protocol
- [ ] Buffer management
- [ ] Quality adaptation

### Phase 4: Advanced Features (v0.4)
- [ ] DHT registry
- [ ] Signal quality metrics
- [ ] Calling system (CQ, selective)
- [ ] Bookmark management
- [ ] Recording functionality

### Phase 5: User Experience (v0.5)
- [ ] TUI implementation
- [ ] Configuration wizard
- [ ] Better error handling
- [ ] Logging system
- [ ] Documentation

### Phase 6: Network Features (v0.6)
- [ ] Repeater mode
- [ ] Multicast support
- [ ] Net control features
- [ ] Emergency priority
- [ ] QSL card system

### Phase 7: Polish (v1.0)
- [ ] Performance optimization
- [ ] Security hardening
- [ ] Comprehensive testing
- [ ] User documentation
- [ ] Example configurations

## 13. Technical Stack

### 13.1 Language & Core Libraries

**Primary Language:** Go (Golang)
- Excellent Yggdrasil integration (yggdrasil-go)
- High performance networking
- Built-in concurrency
- Cross-platform support
- Static binary deployment

**Core Dependencies:**
```go
require (
    github.com/yggdrasil-network/yggdrasil-go v0.5.0
    github.com/hrfee/go-opus v0.0.0  // Opus audio codec
    github.com/gordonklaus/portaudio v0.0.0  // Audio I/O
    github.com/spf13/cobra v1.8.0   // CLI framework
    github.com/charmbracelet/bubbletea v0.25.0  // TUI framework
    github.com/syndtr/goleveldb v1.0.0  // Embedded database
    gopkg.in/yaml.v3 v3.0.1  // Configuration
    github.com/google/uuid v1.5.0  // UUID generation
)
```

### 13.2 Audio Processing

- **Codec:** libopus (via go-opus bindings)
  - Low latency for voice
  - Adaptive bitrate
  - Excellent quality/bandwidth ratio

- **I/O:** PortAudio
  - Cross-platform audio capture/playback
  - Low latency support
  - Multiple backend support

### 13.3 Storage

- **LevelDB:** Local station database, cache
- **SQLite:** Contact logs, history (alternative)
- **YAML:** Configuration files
- **JSON:** Runtime data, IPC

### 13.4 Networking

- **Native Go net package:** IPv6 support
- **Yggdrasil daemon integration:** Via admin API
- **Protocol Buffers:** Optional for efficient serialization

## 14. Performance Considerations

### 14.1 Target Specifications

```yaml
performance_targets:
  audio_latency: "<100ms end-to-end"
  scan_speed: ">1000 IPv6/second (active mode)"
  concurrent_streams: ">10 simultaneous"
  memory_usage: "<50MB idle, <200MB under load"
  cpu_usage: "<5% idle, <25% broadcasting"

scalability:
  max_bookmarks: 10000
  max_database_stations: 100000
  beacon_handling: ">100/second"
```

### 14.2 Optimization Strategies

- **Connection pooling:** Reuse Yggdrasil connections
- **Lazy loading:** Load station data on-demand
- **Caching:** Cache DHT lookups, routing metrics
- **Buffering:** Adaptive jitter buffer for audio
- **Goroutine pools:** Limit concurrent operations
- **Database indexing:** Optimize common queries

## 15. Testing Strategy

### 15.1 Unit Tests
- Protocol encoding/decoding
- Packet validation
- Audio codec wrappers
- Configuration parsing
- DHT operations

### 15.2 Integration Tests
- Yggdrasil connectivity
- Audio streaming end-to-end
- Discovery protocol
- Scanner accuracy
- Database operations

### 15.3 Performance Tests
- Audio latency measurements
- Scan speed benchmarks
- Concurrent connection limits
- Memory leak detection
- CPU profiling

### 15.4 Network Tests
- Multi-node test network
- Various hop counts
- Packet loss simulation
- Bandwidth constraints
- Latency injection

## 16. Documentation Plan

### 16.1 User Documentation
- Installation guide
- Quick start tutorial
- Configuration reference
- CLI command reference
- TUI user guide
- Troubleshooting guide
- FAQ

### 16.2 Developer Documentation
- Architecture overview
- Protocol specification (this document)
- API reference
- Contributing guide
- Code style guide
- Testing guide

### 16.3 Operator Documentation
- Band plan and etiquette
- Station setup guide
- Repeater operation
- Net control procedures
- Emergency communications

## 17. Future Enhancements

### 17.1 Planned Features
- **Digital modes:** PSK, RTTY, FT8-style protocols
- **SSTV:** Slow-scan television over mesh
- **Packet radio:** AX.25-style packet networking
- **APRS integration:** Position reporting and messaging
- **Voice codecs:** Additional codec support (Codec2, etc.)
- **Encryption:** Optional end-to-end encryption layer
- **Federation:** Bridge to other mesh networks

### 17.2 Research Areas
- **AI noise reduction:** ML-based audio enhancement
- **Automatic routing:** Smart relay selection
- **Mesh optimization:** Better route finding
- **Compression:** Advanced audio compression
- **QoS:** Quality of service for priority traffic

## 18. Licensing & Legal

### 18.1 Software License
**Recommended:** GNU General Public License v3.0 (GPL-3.0)
- Ensures open-source ecosystem
- Compatible with Yggdrasil license
- Protects user freedom

### 18.2 Callsign Considerations
- MeshRadio does not require FCC amateur radio licenses
- Operates on internet protocol, not RF spectrum
- Users may choose HAM-style callsigns for identification
- No regulatory restrictions on "virtual frequencies"

### 18.3 Content Policy
- Users responsible for broadcast content
- Emergency channels must be kept clear
- Respect network etiquette
- No spam or abusive beaconing

## 19. Community & Governance

### 19.1 Development Model
- Open-source development on GitHub
- Community-driven feature requests
- Regular release schedule
- Semantic versioning

### 19.2 Band Coordination
- Community-maintained band plan
- Voluntary frequency coordination
- Emergency channel reservation
- Special event frequencies

### 19.3 Support Channels
- GitHub issues for bugs
- Discussion forum for features
- IRC/Matrix chat for real-time help
- Documentation wiki

## 20. References

### 20.1 Related Projects
- **Yggdrasil Network:** https://yggdrasil-network.github.io/
- **Amateur Radio Digital Communications:** ARDC
- **Codec2:** Open-source speech codec
- **GNU Radio:** Software-defined radio

### 20.2 Standards & Protocols
- **IPv6:** RFC 8200
- **Multicast:** RFC 4291
- **Opus Codec:** RFC 6716
- **APRS:** APRS Protocol Reference

### 20.3 Inspiration
- **Traditional HAM Radio:** ARRL Handbook
- **Mesh Networking:** B.A.T.M.A.N., OLSR
- **Internet Radio:** Icecast, Shoutcast
- **Digital Voice:** FreeDV, Codec2

---

## Appendix A: Glossary

- **CQ:** General call to all stations ("Calling any station")
- **DHT:** Distributed Hash Table
- **DX:** Long distance communication
- **Net:** Organized on-air meeting
- **QSL:** Confirmation of contact
- **QSO:** Two-way radio contact/conversation
- **Repeater:** Store-and-forward relay station
- **Simplex:** Direct station-to-station communication
- **Squelch:** Noise suppression when no signal present
- **SSB:** Single sideband (communication mode)
- **SSTV:** Slow-scan television
- **TUI:** Terminal User Interface
- **Yggdrasil:** Encrypted IPv6 mesh network

## Appendix B: Example Workflows

### B.1 First Time Setup
```bash
# Install MeshRadio
go install github.com/meshradio/meshradio@latest

# Ensure Yggdrasil is running
systemctl status yggdrasil

# Initialize configuration
meshradio init --callsign W1AW

# Edit configuration
nano ~/.config/meshradio/meshradio.yaml

# Test audio devices
meshradio test audio

# Start station
meshradio daemon
```

### B.2 Casual Listening
```bash
# Quick scan for active stations
meshradio scan --auto --duration 30

# Tune to found station
meshradio dial 202:1234:5678::1

# Save to bookmarks if you like it
meshradio bookmark add --current --name "Morning News"
```

### B.3 Broadcasting a Show
```bash
# Start broadcaster
meshradio broadcast \
  --callsign W1AW \
  --ipv6 202:1234:5678:abcd::1 \
  --input mic \
  --name "Tech Talk Hour"

# Monitor listener count
meshradio status --listeners

# Stop broadcast
# (Ctrl+C or send SIGTERM)
```

### B.4 Making a Contact
```bash
# Call CQ on calling channel
meshradio call cq --callsign W1AW

# Wait for response...
# Received call from W2XYZ

# Establish QSO
meshradio qso --with W2XYZ --freq 204:1234::1

# Exchange signal reports
# Log contact automatically

# Send QSL confirmation
meshradio qsl --to W2XYZ
```

---

**Document Version:** 1.0
**Last Updated:** 2025-11-25
**Status:** Foundation Design - Ready for Implementation
**Next Steps:** Begin Phase 1 implementation
