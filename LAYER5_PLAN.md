# Layer 5: Emergency Features Implementation Plan

**Date:** 2025-12-03
**Goal:** Implement emergency communications features
**Duration:** ~2-3 hours

---

## Overview

Layer 5 adds emergency-specific features on top of the multicast overlay:

1. **Emergency Channel Registry**: Pre-defined emergency channels with metadata
2. **Priority Signaling**: Indicate broadcast priority (normal/high/emergency/critical)
3. **Auto-tune**: Automatically switch to critical emergency broadcasts
4. **Emergency Notifications**: Alert listeners to high-priority broadcasts

---

## Architecture

### Priority Levels

```
Priority: Critical (3)
├── Auto-tune: Always (unless user disabled)
├── Visual Alert: Red banner
└── Audio Alert: Optional beep

Priority: Emergency (2)
├── Auto-tune: Prompt user
├── Visual Alert: Orange banner
└── Audio Alert: None

Priority: High (1)
├── Auto-tune: Never
├── Visual Alert: Yellow indicator
└── Audio Alert: None

Priority: Normal (0)
├── Auto-tune: Never
├── Visual Alert: None
└── Audio Alert: None
```

### Emergency Channels

```
Port 8790: General Emergency
- Group: "emergency"
- Priority: Critical
- Auto-tune: Always
- Description: "General emergency broadcast"

Port 8791: Net Control
- Group: "netcontrol"
- Priority: Emergency
- Auto-tune: Prompt
- Description: "Emergency net control coordination"

Port 8792: Medical
- Group: "medical"
- Priority: Emergency
- Auto-tune: Prompt
- Description: "Medical emergency coordination"

Port 8793: Weather
- Group: "weather"
- Priority: High
- Auto-tune: Never
- Description: "Weather alerts and warnings"

Port 8794: Search & Rescue
- Group: "sar"
- Priority: Emergency
- Auto-tune: Prompt
- Description: "Search and rescue coordination"
```

---

## Implementation Steps

### Part 1: Emergency Channel Registry (30 min)

**Task 1.1: Create emergency package**

```
pkg/emergency/
  channels.go  - Channel registry and definitions
  priority.go  - Priority levels and handling
  types.go     - Common types
```

**Task 1.2: Define ChannelRegistry**

```go
type Channel struct {
    Name        string
    Group       string
    Port        int
    Priority    Priority
    AutoTune    AutoTuneMode
    Description string
}

type ChannelRegistry struct {
    channels map[string]*Channel  // Key: channel name
}

var StandardChannels = map[string]Channel{
    "emergency": {
        Name:        "emergency",
        Group:       "emergency",
        Port:        8790,
        Priority:    PriorityCritical,
        AutoTune:    AutoTuneAlways,
        Description: "General emergency broadcast",
    },
    "netcontrol": { ... },
    "medical": { ... },
    "weather": { ... },
    "sar": { ... },
}
```

**Task 1.3: Priority types**

```go
type Priority int

const (
    PriorityNormal   Priority = 0
    PriorityHigh     Priority = 1
    PriorityEmergency Priority = 2
    PriorityCritical Priority = 3
)

type AutoTuneMode int

const (
    AutoTuneNever  AutoTuneMode = 0
    AutoTunePrompt AutoTuneMode = 1
    AutoTuneAlways AutoTuneMode = 2
)
```

### Part 2: Priority Signaling (45 min)

**Task 2.1: Extend protocol for priority**

Add priority to existing packets:

```go
// In pkg/protocol/packet.go
type Packet struct {
    // ... existing fields ...
    Priority uint8  // Use existing Reserved field or add new field
}
```

Or use Flags field:
```go
const (
    FlagPriority0 uint8 = 0x10  // Bit 4-5 for priority
    FlagPriority1 uint8 = 0x20
)

func (p *Packet) GetPriority() Priority {
    return Priority((p.Flags >> 4) & 0x03)
}

func (p *Packet) SetPriority(pri Priority) {
    p.Flags = (p.Flags & 0xCF) | (uint8(pri) << 4)
}
```

**Task 2.2: Broadcaster sends priority**

```go
// In internal/broadcaster/broadcaster.go
func (b *Broadcaster) Start() error {
    // ... existing code ...

    // Set priority based on channel
    b.priority = b.getPriorityForGroup(b.group)
}

func (b *Broadcaster) broadcastLoop() {
    // ... create packet ...

    packet.SetPriority(b.priority)

    // ... send packet ...
}
```

**Task 2.3: Listener detects priority**

```go
// In internal/listener/listener.go
func (l *Listener) handleAudioPacket(packet *protocol.Packet) {
    priority := packet.GetPriority()

    if priority >= emergency.PriorityCritical {
        l.handleCriticalBroadcast(packet)
    } else if priority >= emergency.PriorityEmergency {
        l.handleEmergencyBroadcast(packet)
    }

    // ... normal audio handling ...
}
```

### Part 3: Auto-tune Logic (45 min)

**Task 3.1: Emergency notification**

```go
type EmergencyNotification struct {
    Channel    string
    Priority   Priority
    Callsign   string
    IPv6       net.IP
    Port       int
    Timestamp  time.Time
}

type EmergencyHandler interface {
    OnEmergencyBroadcast(notif EmergencyNotification)
}
```

**Task 3.2: Listener auto-tune**

```go
func (l *Listener) handleCriticalBroadcast(packet *protocol.Packet) {
    // Check auto-tune settings
    if l.autoTuneMode == emergency.AutoTuneNever {
        l.notifyEmergency(packet)
        return
    }

    if l.autoTuneMode == emergency.AutoTuneAlways {
        l.switchToEmergency(packet)
        return
    }

    // AutoTunePrompt: Ask user
    l.promptEmergencySwitch(packet)
}

func (l *Listener) switchToEmergency(packet *protocol.Packet) {
    sourceIPv6 := protocol.BytesToIPv6(packet.SourceIPv6)

    // Save current channel
    l.savedChannel = l.currentChannel

    // Switch to emergency broadcast
    l.tuneToSource(sourceIPv6, emergency.PortEmergency)

    l.notifyEmergencyActive()
}
```

**Task 3.3: UI integration**

```go
// In pkg/ui/model.go
type emergencyNotifMsg struct {
    notification emergency.EmergencyNotification
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case emergencyNotifMsg:
        return m.handleEmergency(msg.notification)
    // ...
    }
}

func (m Model) handleEmergency(notif emergency.EmergencyNotification) (Model, tea.Cmd) {
    if notif.Priority >= emergency.PriorityCritical {
        // Show critical alert banner
        m.showCriticalAlert(notif)

        // Auto-tune if enabled
        if m.autoTune {
            return m.tuneToEmergency(notif)
        }
    }

    return m, nil
}
```

### Part 4: Emergency Test Program (30 min)

**Task 4.1: Create test program**

```
cmd/emergency-test/
  main.go - Test emergency features
```

**Modes:**
- `broadcast-critical` - Broadcast with critical priority
- `broadcast-emergency` - Broadcast with emergency priority
- `listen-autotune` - Listen with auto-tune enabled
- `listen-manual` - Listen without auto-tune

**Task 4.2: Test scenarios**

**Scenario 1: Critical broadcast with auto-tune**
```bash
# Terminal 1: Normal broadcast
./emergency-test -mode broadcast -priority normal -port 8795

# Terminal 2: Listener (auto-tune enabled)
./emergency-test -mode listen -autotune always

# Terminal 3: Critical emergency broadcast
./emergency-test -mode broadcast -priority critical -port 8790

# Expected: Listener auto-switches to emergency broadcast
```

**Scenario 2: Emergency with prompt**
```bash
# Listener with prompt mode
./emergency-test -mode listen -autotune prompt

# Expected: Listener shows prompt, waits for user input
```

---

## Code Structure

```
pkg/emergency/
  types.go       - Priority, AutoTuneMode, Channel
  channels.go    - ChannelRegistry, StandardChannels
  priority.go    - Priority handling logic
  notification.go - EmergencyNotification

pkg/protocol/
  packet.go      - Add GetPriority() / SetPriority()

internal/broadcaster/
  broadcaster.go - Set priority based on channel

internal/listener/
  listener.go    - Handle priority, auto-tune logic

cmd/emergency-test/
  main.go        - Test program

pkg/ui/
  model.go       - Emergency UI alerts
```

---

## Emergency Channel Specification

### Standard Channels

| Channel     | Port | Group       | Priority  | Auto-tune | Description                    |
|-------------|------|-------------|-----------|-----------|--------------------------------|
| emergency   | 8790 | emergency   | Critical  | Always    | General emergency broadcast    |
| netcontrol  | 8791 | netcontrol  | Emergency | Prompt    | Emergency net control          |
| medical     | 8792 | medical     | Emergency | Prompt    | Medical emergency coordination |
| weather     | 8793 | weather     | High      | Never     | Weather alerts and warnings    |
| sar         | 8794 | sar         | Emergency | Prompt    | Search and rescue              |
| community   | 8795 | community   | Normal    | Never     | Community/public service       |
| talk        | 8798 | talk        | Normal    | Never     | General conversation           |
| test        | 8799 | test        | Normal    | Never     | Testing                        |

### Priority Behavior

**Critical (3):**
- Red banner in UI
- Auto-tune: Always (unless explicitly disabled)
- Overrides all other broadcasts
- Use case: Active emergency in progress

**Emergency (2):**
- Orange banner in UI
- Auto-tune: Prompt user
- High importance but allows user choice
- Use case: Emergency coordination, net control

**High (1):**
- Yellow indicator in UI
- Auto-tune: Never
- Important information, not critical
- Use case: Weather alerts, warnings

**Normal (0):**
- No special indication
- Auto-tune: Never
- Regular communications
- Use case: Community chat, testing

---

## User Preferences

```go
type EmergencySettings struct {
    AutoTuneMode      AutoTuneMode  // Always/Prompt/Never
    CriticalChannels  []string      // Which channels trigger auto-tune
    VisualAlerts      bool          // Show banners
    AudioAlerts       bool          // Play alert sound
    SavedChannel      bool          // Remember channel before emergency
    AutoReturn        bool          // Return to saved channel when emergency ends
}
```

---

## API Extensions

### Start Emergency Broadcast

```
POST /broadcast/emergency
{
  "channel": "emergency",
  "priority": "critical"
}
```

### Emergency Status

```
GET /emergency/status
{
  "active": true,
  "channel": "emergency",
  "priority": "critical",
  "broadcaster": "W1EMERGENCY",
  "started_at": "2025-12-03T10:30:00Z"
}
```

### Emergency History

```
GET /emergency/history
[
  {
    "channel": "emergency",
    "priority": "critical",
    "broadcaster": "W1EMERGENCY",
    "started_at": "2025-12-03T10:30:00Z",
    "ended_at": "2025-12-03T11:15:00Z"
  }
]
```

---

## Testing Plan

### Unit Tests
- [ ] ChannelRegistry lookup
- [ ] Priority encoding/decoding
- [ ] Auto-tune decision logic

### Integration Tests
- [ ] Critical broadcast triggers auto-tune
- [ ] Emergency broadcast shows prompt
- [ ] High priority shows indicator only
- [ ] Normal priority has no special handling

### System Tests
- [ ] End-to-end: Critical broadcast → auto-tune → listener switches
- [ ] Multiple listeners with different auto-tune settings
- [ ] Return to saved channel after emergency ends

---

## Success Criteria

By end of this session:

- [ ] Emergency channel registry defined
- [ ] Priority encoding in protocol packets
- [ ] Broadcaster sets priority based on channel
- [ ] Listener detects and handles priority
- [ ] Auto-tune logic implemented
- [ ] Emergency test program works
- [ ] UI shows emergency alerts
- [ ] Documentation complete

---

## Future Enhancements

- [ ] Emergency broadcast recording
- [ ] Emergency message queue (replay missed alerts)
- [ ] Multi-language emergency alerts
- [ ] Emergency contact directory
- [ ] Integration with external alert systems (NOAA, FEMA, etc.)
- [ ] Emergency drill mode (testing without real alerts)

---

**Ready to build emergency features!** 🚨
