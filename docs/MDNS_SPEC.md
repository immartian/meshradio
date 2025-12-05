# MeshRadio mDNS/Avahi Service Specification

**Version:** 1.0
**Status:** Layer 3 (Discovery) - mDNS Implementation ✅
**Date:** 2025-12-03

## Overview

MeshRadio uses mDNS (Multicast DNS) / Avahi for automatic service discovery on local networks. This allows stations to be discovered without manual configuration.

## Service Type

```
_meshradio._udp.local.
```

- **Service Type:** `_meshradio._udp`
- **Protocol:** UDP
- **Domain:** `local.`

## TXT Record Format

### Required Fields

| Field      | Type   | Description                                    | Example         |
|------------|--------|------------------------------------------------|-----------------|
| `callsign` | string | Station identifier (HAM-style callsign)        | `W1EMERGENCY`   |
| `channel`  | string | Channel type                                   | `emergency`     |
| `group`    | string | Multicast group label (same as channel)        | `emergency`     |
| `port`     | int    | RTP streaming port                             | `8790`          |

### Optional Fields

| Field      | Type   | Description                                    | Example         |
|------------|--------|------------------------------------------------|-----------------|
| `priority` | string | Priority level                                 | `critical`      |
| `codec`    | string | Audio codec                                    | `opus`          |
| `bitrate`  | int    | Bitrate in kbps                                | `64`            |

## Field Values

### Channel/Group Values

- `emergency` - Emergency communications (ports 8790-8794)
- `community` - Community/public service (ports 8795-8797)
- `talk` - General conversation (ports 8798-8799)

### Priority Values

- `normal` - Standard priority
- `high` - High priority
- `emergency` - Emergency priority
- `critical` - Critical emergency priority

### Codec Values

- `opus` - Opus audio codec (default)

## Example TXT Records

### Emergency Channel

```
_meshradio._udp.local.
TXT:
  callsign=W1EMERGENCY
  group=emergency
  channel=emergency
  port=8790
  priority=critical
  codec=opus
  bitrate=64
```

### Community Channel

```
_meshradio._udp.local.
TXT:
  callsign=MESH-COMMUNITY
  group=community
  channel=community
  port=8795
  priority=normal
  codec=opus
  bitrate=64
```

## Port Allocation

Emergency channels use standardized port numbers for easy verbal sharing:

| Port Range  | Usage                          | Priority      |
|-------------|--------------------------------|---------------|
| 8790-8794   | Emergency channels             | Critical/High |
| 8795-8797   | Community/public service       | Normal/High   |
| 8798-8799   | General conversation/talk      | Normal        |

**Rationale:** Port numbers containing "799" reference Yggdrasil's Nordic mythology theme while being easy to remember and share verbally during emergencies.

## Discovery Flow

### Broadcasting Station

1. Start RTP broadcaster on chosen port (e.g., 8790)
2. Create `ServiceInfo` with station metadata
3. Call `mdns.NewAdvertiser(info)` to advertise service
4. Service becomes discoverable on local network
5. On shutdown, call `Shutdown()` to withdraw advertisement

### Listening Station

1. Create `mdns.NewBrowser()`
2. Call `Browse(opts)` with optional filters
3. Receive list of discovered services
4. Select service and extract IPv6 address and port
5. Connect to broadcaster using RTP

## Testing Tools

### Advertise a Service

```bash
./mdns-test -mode advertise -callsign STATION1 -port 8790 -channel emergency -priority critical
```

### Browse for Services

```bash
./mdns-test -mode browse -timeout 3
```

### Browse with Filter

```bash
./mdns-test -mode browse -filter emergency
```

### Query Specific Callsign

```bash
./mdns-test -mode query -callsign STATION1
```

### Using Avahi Tools (Linux)

```bash
# Browse for MeshRadio services
avahi-browse -r _meshradio._udp

# Resolve specific service
avahi-resolve -n STATION1._meshradio._udp.local
```

## Implementation Example

### Advertiser

```go
import "github.com/meshradio/meshradio/pkg/mdns"

info := mdns.ServiceInfo{
    Name:     "EMERGENCY-1",
    Port:     8790,
    Callsign: "W1EMERGENCY",
    Group:    "emergency",
    Channel:  "emergency",
    Priority: "critical",
    Codec:    "opus",
    Bitrate:  64,
}

advertiser, err := mdns.NewAdvertiser(info)
if err != nil {
    log.Fatal(err)
}
defer advertiser.Shutdown()

// Service is now advertised
// Keep running...
```

### Browser

```go
import "github.com/meshradio/meshradio/pkg/mdns"

browser, err := mdns.NewBrowser()
if err != nil {
    log.Fatal(err)
}

opts := mdns.BrowseOptions{
    Timeout:       3 * time.Second,
    FilterChannel: "emergency", // Optional filter
}

services, err := browser.Browse(opts)
if err != nil {
    log.Fatal(err)
}

for _, service := range services {
    fmt.Printf("Found: %s at %s\n",
        service.Callsign,
        service.GetIPv6Addr())
}
```

## Network Scope

### Local Network (LAN)

mDNS works on:
- Local Ethernet segments
- WiFi networks
- Link-local IPv6 (fe80::)

### Yggdrasil Mesh

**Important:** mDNS is **link-local only** and does NOT work across Yggdrasil mesh segments.

**Workarounds:**
1. **Manual entry:** Use emergency port numbers (8790-8799) for direct connection
2. **mDNS relay:** Future enhancement to relay mDNS across mesh segments
3. **Out-of-band sharing:** Share IPv6:port via other channels (SMS, QR code, etc.)

## Security Considerations

1. **No Authentication:** mDNS does not provide authentication
2. **Local Network Trust:** Assumes local network is trusted
3. **Firewall:** Ensure port 5353/UDP (mDNS) is allowed
4. **Emergency Use:** Design assumes emergency scenario where trust is implicit

## Future Enhancements

- [ ] mDNS relay/proxy for mesh-wide discovery
- [ ] Service updates (requires re-advertising in current implementation)
- [ ] DNS-SD service browsing
- [ ] Integration with Web UI
- [ ] QR code generation for out-of-band sharing

## References

- [RFC 6762 - Multicast DNS](https://datatracker.ietf.org/doc/html/rfc6762)
- [RFC 6763 - DNS-Based Service Discovery](https://datatracker.ietf.org/doc/html/rfc6763)
- [Avahi Documentation](https://www.avahi.org/)
- [zeroconf Library](https://github.com/grandcat/zeroconf)

---

**Layer 3 Complete!** ✅ mDNS discovery works on local networks.

**Next:** Layer 4 (Multicast Overlay)
