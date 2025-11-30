# Migration to v0.4-alpha

Quick checklist for upgrading from v0.3 to v0.4.

## What Changed?

**Bottom line:** Multicast → Subscription-based unicast

## Code Changes Required

### If You Were Using the Listener API

```go
// OLD (v0.3)
l, err := listener.New(listener.Config{
    TargetIPv6:  targetIP,
    TargetPort:  8799,
    LocalPort:   9799,
    AudioConfig: audioConfig,
})

// NEW (v0.4) - Added callsign and local IPv6
l, err := listener.New(listener.Config{
    Callsign:    "W2LISTENER",        // NEW: Your callsign
    LocalIPv6:   yourYggdrasilIPv6,   // NEW: Your Yggdrasil IPv6
    LocalPort:   9799,
    TargetIPv6:  broadcasterIP,
    TargetPort:  8799,
    AudioConfig: audioConfig,
})
```

### If You Were Using the Broadcaster API

```go
// Broadcaster API unchanged, but behavior different:
// - No longer broadcasts to multicast address
// - Now tracks individual listeners
// - New methods available:

count := broadcaster.GetListenerCount()
listeners := broadcaster.GetListeners()  // Get snapshot of connected listeners
```

## User-Facing Changes

### Broadcasting

```bash
# OLD (v0.3)
$ meshradio broadcast --callsign W1ALICE
> Broadcasting on ff02::1  # Multicast (broken on mesh)

# NEW (v0.4)
$ meshradio broadcast --callsign W1ALICE
> Broadcasting on 201:abcd:1234::1:8799
> Share this address with listeners!  # Share your IPv6:port
> 0 listeners connected
> [later] 1 listener connected: W2BOB
```

### Listening

```bash
# OLD (v0.3)
$ meshradio listen
> Listening on port 9799...
> Waiting for broadcasts on ff02::1  # Passive listening

# NEW (v0.4)
$ meshradio listen 201:abcd:1234::1:8799  # Must specify broadcaster's address
> Subscribing to 201:abcd:1234::1:8799...
> Sent SUBSCRIBE
> Receiving audio from W1ALICE
```

## Discovery Changes

### Before (v0.3)
- Listeners passively received multicast beacons
- Automatic discovery within LAN

### After (v0.4)
- **Manual discovery (MVP):** Share IPv6:port out-of-band
- Listeners explicitly subscribe to broadcaster

### How to Share Your Station

Choose any method:
1. **Chat/IRC/Matrix:** "Tune to 201:abcd:1234::1:8799"
2. **Email:** Send your IPv6:port
3. **QR code:** Generate QR with `meshradio://201:abcd:1234::1:8799`
4. **Config file:** Share `stations.yaml` with your station info

## Testing Your Migration

### Step 1: Get Your Yggdrasil IPv6
```bash
yggdrasilctl getSelf
# Note the IPv6 address
```

### Step 2: Start Broadcasting
```bash
./meshradio broadcast --callsign YOURSTATION
# Note the IPv6:port shown (e.g., 201:abcd::1:8799)
```

### Step 3: Test Listener (Another Machine)
```bash
# On a different machine with Yggdrasil running:
./meshradio listen 201:abcd::1:8799 --callsign LISTENER1
```

### Expected Output

**Broadcaster terminal:**
```
Broadcasting on 201:abcd:1234::1:8799
Share this address with listeners!
0 listeners connected
[5s later]
New listener: LISTENER1 (202:5678::2)
1 listener connected
Broadcasting: seq=50, size=120 bytes to 1 listeners
```

**Listener terminal:**
```
Subscribing to 201:abcd:1234::1:8799...
Sent SUBSCRIBE to 201:abcd:1234::1:8799
Subscribed to 201:abcd:1234::1:8799
Receiving audio from YOURSTATION
Received: packets=50, seq=49, from=YOURSTATION
```

## Troubleshooting

### "No audio received"

Check:
1. Both nodes connected to Yggdrasil: `yggdrasilctl getSelf`
2. Can ping broadcaster: `ping6 201:abcd:1234::1`
3. Listener using correct port (usually 8799)
4. Broadcaster running before listener subscribes

### "Listener timeout" on broadcaster

- Listener not sending heartbeats (check listener is running)
- Network issue preventing heartbeat packets
- Listener IPv6 changed (restart listener)

### "Failed to subscribe"

- Broadcaster not running
- Incorrect IPv6:port
- Yggdrasil not connected between nodes
- Firewall blocking UDP

## Rollback (If Needed)

```bash
git checkout v0.3-alpha
go build -o meshradio ./cmd/meshradio
```

**Warning:** v0.3 multicast only works on local LAN, not across Yggdrasil mesh.

## What's Next?

See [ARCHITECTURE_V04.md](ARCHITECTURE_V04.md) for:
- Complete architecture details
- Future discovery plans
- Performance characteristics

---

**Questions?** Open an issue on GitHub or check the docs.
