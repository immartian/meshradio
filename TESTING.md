# MeshRadio Testing Guide

Complete guide for testing all MeshRadio features.

## Quick Test (Single Machine)

For a quick sanity check on a single machine:

```bash
# Build everything
./build.sh

# Run quick tests
./rtp-test broadcast -callsign TEST &
sleep 2
./rtp-test listen -target ::1 -port 8799
```

## Full Integration Test (Two Nodes)

The integration test script validates all layers of the MeshRadio stack.

### Prerequisites

Both nodes must have:
- ✅ Go 1.21+
- ✅ Yggdrasil running
- ✅ PortAudio (for audio I/O)
- ✅ Avahi (for mDNS)
- ✅ MeshRadio built (`./build.sh`)

### Running the Integration Test

#### Option 1: Two Separate Machines

**On Node 1:**
```bash
./test-integration.sh node1
```

**On Node 2:**
```bash
./test-integration.sh node2
# When prompted, enter Node 1's IPv6 address
```

#### Option 2: Same Machine (Two Terminals)

**Terminal 1:**
```bash
./test-integration.sh node1
```

**Terminal 2:**
```bash
./test-integration.sh node2
# When prompted for IPv6, enter: ::1
```

### What Gets Tested

The integration test validates all 5 layers:

| Test | Layer | Feature | What to Verify |
|------|-------|---------|----------------|
| 1 | Layer 2 | RTP Streaming | Audio packets transmitted |
| 2 | Layer 3 | mDNS Discovery | Service advertised and discovered |
| 3 | Layer 4 | Regular Multicast | Any-source multicast works |
| 4 | Layer 4 | SSM | Source-specific multicast works |
| 5 | Layer 5 | Normal Priority | No special alerts |
| 6 | Layer 5 | High Priority | 📢 alerts displayed |
| 7 | Layer 5 | Emergency Priority | ⚠️ alerts displayed |
| 8 | Layer 5 | Critical Priority | 🚨 alerts displayed |

### Test Duration

- Full test suite: ~10-15 minutes
- Each test: 15-20 seconds
- Interactive pauses between tests for verification

## Manual Testing

### Test 1: RTP Streaming

**Broadcaster:**
```bash
./rtp-test broadcast -callsign STATION-TX -port 8799
```

**Listener:**
```bash
./rtp-test listen -target <broadcaster-ipv6> -port 8799
```

**Verify:**
- ✅ "Broadcasting" messages appear
- ✅ "Received packet" messages appear
- ✅ Sequence numbers increment
- ✅ No errors about failed sends

### Test 2: mDNS Discovery

**Advertiser (Terminal 1):**
```bash
./mdns-test advertise -callsign BEACON-1 -port 8790 -group emergency
```

**Browser (Terminal 2):**
```bash
./mdns-test browse
```

**Verify:**
- ✅ Service appears: `BEACON-1._meshradio._udp.local.`
- ✅ TXT records show: `group=emergency`, `port=8790`
- ✅ IPv6 address displayed
- ✅ Service disappears when advertiser stops

### Test 3: Multicast Overlay - Regular Multicast

**Broadcaster A (Terminal 1):**
```bash
./multicast-test broadcast-regular -callsign EMERGENCY-A -port 8790
```

**Broadcaster B (Terminal 2):**
```bash
./multicast-test broadcast-regular -callsign EMERGENCY-B -port 8790
```

**Listener (Terminal 3):**
```bash
# Get IPv6 addresses from Terminals 1 & 2 first
./multicast-test listen-regular -target <broadcaster-a-ipv6> -port 8790
```

**Verify:**
- ✅ Listener receives from BOTH broadcasters
- ✅ Different callsigns appear in logs
- ✅ Subscription manager tracks both sources
- ✅ Regular multicast mode confirmed

### Test 4: Multicast Overlay - SSM

**Broadcaster A (Terminal 1):**
```bash
./multicast-test broadcast-ssm -callsign COMMUNITY-A -port 8795
```

**Broadcaster B (Terminal 2):**
```bash
./multicast-test broadcast-ssm -callsign COMMUNITY-B -port 8795
```

**Listener (Terminal 3):**
```bash
# Get IPv6 address from Terminal 1 only
./multicast-test listen-ssm -target <broadcaster-a-ipv6> -port 8795
```

**Verify:**
- ✅ Listener receives ONLY from Broadcaster A
- ✅ Only COMMUNITY-A callsign appears in logs
- ✅ SSM mode confirmed in subscription log
- ✅ Broadcaster B packets ignored

### Test 5: Emergency Priority Levels

**Test Normal Priority:**

Broadcaster:
```bash
./emergency-test broadcast-normal
```

Listener:
```bash
./emergency-test listen-manual -target <ipv6> -port 8795 -group community
```

Verify: ✅ No priority alerts (just normal packet logs)

---

**Test High Priority:**

Broadcaster:
```bash
./emergency-test broadcast-high
```

Listener:
```bash
./emergency-test listen-manual -target <ipv6> -port 8793 -group weather
```

Verify: ✅ See `📢 High priority broadcast` alerts

---

**Test Emergency Priority:**

Broadcaster:
```bash
./emergency-test broadcast-emergency
```

Listener:
```bash
./emergency-test listen-manual -target <ipv6> -port 8791 -group netcontrol
```

Verify: ✅ See `⚠️  EMERGENCY BROADCAST` alerts

---

**Test Critical Priority:**

Broadcaster:
```bash
./emergency-test broadcast-critical
```

Listener:
```bash
./emergency-test listen-manual -target <ipv6> -port 8790 -group emergency
```

Verify: ✅ See `🚨 CRITICAL EMERGENCY BROADCAST` alerts

## Performance Testing

### Latency Test

```bash
# Node 1: Broadcaster with timing
./rtp-test broadcast -callsign LATENCY-TEST -port 8799

# Node 2: Listener with timing
./rtp-test listen -target <ipv6> -port 8799

# Calculate end-to-end latency from logs
# Look for packet transmission time vs. reception time
```

### Multi-Listener Test

Test broadcaster fan-out to multiple listeners:

```bash
# Node 1: Broadcaster
./multicast-test broadcast-regular -callsign BROADCAST-1 -port 8790

# Nodes 2-5: Listeners
./multicast-test listen-regular -target <node1-ipv6> -port 8790
```

**Verify:**
- ✅ All listeners receive packets
- ✅ Broadcaster shows correct listener count
- ✅ No packet loss or errors

### Network Partition Test

Test recovery after network interruption:

1. Start broadcaster and listener
2. Disable network (or stop Yggdrasil)
3. Wait 20 seconds
4. Re-enable network (or restart Yggdrasil)

**Verify:**
- ✅ Listener timeout detected (15s)
- ✅ Heartbeat resumes after reconnection
- ✅ Audio stream recovers automatically

## Stress Testing

### High Packet Rate

```bash
# Modify audio config for smaller frames (higher rate)
# Edit code: FrameSize: 480 (10ms frames = 100 packets/sec)
# Rebuild and test
```

### Long Duration

```bash
# Run broadcaster and listener for 1+ hour
# Monitor memory usage, CPU usage, packet drops
```

### Multiple Channels

Run multiple channels simultaneously:

```bash
# Terminal 1: Emergency channel
./emergency-test broadcast-critical

# Terminal 2: Community channel
./emergency-test broadcast-normal

# Terminal 3: Weather channel
./emergency-test broadcast-high

# Terminal 4: Listen to all (open 3 listener processes)
```

## Troubleshooting Tests

### Issue: No packets received

**Debug steps:**
```bash
# Check Yggdrasil connectivity
yggdrasilctl getSelf
yggdrasilctl getPeers

# Check port is listening
sudo netstat -tulpn | grep 8790

# Check firewall
sudo ufw status
sudo ufw allow 8790:8799/udp

# Test with localhost first
./rtp-test broadcast -callsign TEST &
./rtp-test listen -target ::1 -port 8799
```

### Issue: mDNS not discovering services

**Debug steps:**
```bash
# Check Avahi is running
sudo systemctl status avahi-daemon

# Test Avahi directly
avahi-browse -a

# Check firewall allows mDNS
sudo ufw allow 5353/udp

# Test on localhost
./mdns-test advertise -callsign TEST &
./mdns-test browse
```

### Issue: Priority alerts not showing

**Debug steps:**
```bash
# Verify priority is set in broadcaster logs
# Should see: "Registered broadcaster in group 'emergency' with priority 'critical'"

# Check listener detects priority
# Should see priority field in packet logs

# Verify packet flags encoding
# Debug with: grep "Priority:" output
```

## Automated Testing (Future)

### Unit Tests

```bash
# Run Go unit tests (when available)
go test ./pkg/protocol -v
go test ./pkg/emergency -v
go test ./pkg/multicast -v
```

### Integration Tests

```bash
# Automated integration test suite (planned)
go test ./test/integration -v
```

### CI/CD Pipeline

```yaml
# .github/workflows/test.yml (planned)
# - Build on Ubuntu, macOS, Windows
# - Run unit tests
# - Run integration tests with mocked network
# - Check for regressions
```

## Test Results Log

Keep a log of test results:

```bash
# Create test log
./test-integration.sh node1 2>&1 | tee test-results-$(date +%Y%m%d-%H%M%S).log
```

## Reporting Issues

When reporting test failures, include:

1. **Test command:** Exact command that failed
2. **Error output:** Full error messages
3. **Environment:** OS, Go version, Yggdrasil version
4. **Network:** Local? Remote? Yggdrasil mesh?
5. **Logs:** Relevant log snippets from both nodes

Submit issues at: https://github.com/immartian/meshradio/issues

## Test Checklist

Before a release, verify:

- [ ] All 8 integration tests pass
- [ ] RTP streaming works both directions
- [ ] mDNS discovery works on local network
- [ ] Regular multicast works (emergency channels)
- [ ] SSM works (regular channels)
- [ ] All 4 priority levels display correctly
- [ ] Multi-listener fan-out works (5+ listeners)
- [ ] Network partition recovery works
- [ ] Long duration test (1+ hour) stable
- [ ] Cross-platform (Linux, macOS) tested

## Performance Benchmarks

Target metrics:

- **Latency:** < 100ms end-to-end
- **Jitter:** < 20ms variation
- **Packet Loss:** < 1% under normal conditions
- **Max Listeners:** 50+ per broadcaster
- **CPU Usage:** < 5% when idle, < 20% when broadcasting
- **Memory:** < 50MB per process

Record actual results in test logs for comparison.
