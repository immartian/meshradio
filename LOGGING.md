# MeshRadio Logging System

## Overview

MeshRadio uses a structured logging system with automatic log rotation to manage debug and operational logs from both broadcaster and listener components.

## Log Locations

By default, logs are stored in:
```
~/.meshradio/logs/
├── broadcaster.log      # Broadcaster component logs
├── listener.log         # Listener component logs
├── playback.log         # Audio playback debugging
├── *.log.1             # Rotated log files (oldest backup)
├── *.log.2             # Rotated log files
└── *.log.3             # Rotated log files (newest backup)
```

## Log Levels

- **DEBUG**: Detailed diagnostic information (heartbeats, packet counts, buffer status)
- **INFO**: General operational messages (startup, shutdown, connections)
- **WARN**: Warning conditions (buffer underruns, timeouts, recoverable errors)
- **ERROR**: Error conditions (failed operations, unrecoverable errors)

## Configuration

### Environment Variables

Control logging behavior with these environment variables:

```bash
# Set log level (DEBUG, INFO, WARN, ERROR)
export MESHRADIO_LOG_LEVEL=DEBUG

# Set custom log directory
export MESHRADIO_LOG_DIR=/var/log/meshradio
```

### Examples

**Production mode (INFO level, clean output):**
```bash
export MESHRADIO_LOG_LEVEL=INFO
./meshradio --gui --callsign MY-STATION
```

**Debug mode (verbose logging):**
```bash
export MESHRADIO_LOG_LEVEL=DEBUG
./music-broadcast --dir ~/Music --callsign DJ-DEBUG
```

**Custom log directory:**
```bash
export MESHRADIO_LOG_DIR=/tmp/meshradio-logs
./meshradio
```

## Log Rotation

Logs automatically rotate when they reach 10 MB in size. The system keeps the last 3 backup files:

- `component.log` - Current log file
- `component.log.1` - Most recent backup
- `component.log.2` - Second backup
- `component.log.3` - Oldest backup (deleted when new rotation occurs)

## Log Format

All log entries follow this format:
```
YYYY-MM-DD HH:MM:SS [LEVEL] [component] message
```

Example:
```
2024-12-05 10:30:45 [INFO] [playback] Starting audio playback: 48000 Hz, 2 channels, frameSize=960
2024-12-05 10:30:50 [DEBUG] [playback] Playback: callback=250, buffer=145/150
2024-12-05 10:31:00 [WARN] [playback] Playback underrun: callback=300, buffer=0/150, timeout waiting for frame
```

## Viewing Logs

### Real-time monitoring:
```bash
# Watch all playback logs
tail -f ~/.meshradio/logs/playback.log

# Watch broadcaster logs
tail -f ~/.meshradio/logs/broadcaster.log

# Filter for specific level
grep "\[WARN\]" ~/.meshradio/logs/*.log

# Filter for heartbeat messages
grep "Heartbeat" ~/.meshradio/logs/broadcaster.log
```

### Search historical logs:
```bash
# Find all error messages
grep -r "\[ERROR\]" ~/.meshradio/logs/

# Find buffer underruns
grep "underrun" ~/.meshradio/logs/playback.log*

# Count log entries by level
grep -oh "\[.*\]" ~/.meshradio/logs/playback.log | sort | uniq -c
```

## Component-Specific Logging

### Playback Logger

The playback logger tracks audio output performance:

- **Startup**: Logs sample rate, channel count, frame size
- **Periodic status** (every ~5 seconds): Callback count, buffer fill level
- **Underruns**: Warnings when buffer runs dry
- **Shutdown**: Final statistics

Example:
```
2024-12-05 10:30:45 [INFO] [playback] Starting audio playback: 48000 Hz, 2 channels, frameSize=960
2024-12-05 10:30:50 [DEBUG] [playback] Playback: callback=250, buffer=145/150
2024-12-05 10:30:55 [DEBUG] [playback] Playback: callback=500, buffer=148/150
2024-12-05 10:31:00 [INFO] [playback] Audio playback stopped
```

### Broadcaster Logger

The broadcaster logger tracks transmission and subscriber management:

- Subscriber registrations and heartbeats
- Packet transmission counts
- Group management (creation, deletion)
- Subscriber pruning

### Listener Logger

The listener logger tracks reception:

- Connection status
- Packet reception counts
- Audio decode statistics
- Buffer management

## Troubleshooting

### Issue: Logs growing too large

**Solution**: Reduce log level to INFO or WARN:
```bash
export MESHRADIO_LOG_LEVEL=INFO
```

### Issue: Can't find logs

**Solution**: Check default location:
```bash
ls -lh ~/.meshradio/logs/
```

Or set explicit path:
```bash
export MESHRADIO_LOG_DIR=/tmp/meshradio-logs
```

### Issue: Permission denied writing logs

**Solution**: Ensure log directory exists and is writable:
```bash
mkdir -p ~/.meshradio/logs
chmod 755 ~/.meshradio/logs
```

### Issue: Too many debug messages

**Solution**: Use INFO level for normal operation, DEBUG only when troubleshooting:
```bash
# Normal operation
export MESHRADIO_LOG_LEVEL=INFO

# Only when debugging specific issues
export MESHRADIO_LOG_LEVEL=DEBUG
```

## Performance Impact

- **DEBUG level**: Logs every ~5 seconds (250 callbacks), minimal performance impact
- **INFO level**: Logs startup/shutdown and major events only
- **WARN/ERROR level**: Only logs problems, negligible performance impact

Log rotation happens asynchronously and does not block audio processing.

## Migration from Old System

Previous versions wrote logs to `/tmp/meshradio-playback.log` without rotation. The new system:

- Uses `~/.meshradio/logs/` by default (persists across reboots)
- Automatically rotates logs to prevent unbounded growth
- Separates logs by component for easier debugging
- Supports configurable log levels
- Writes to both file and stdout (dual output)

Old log files in `/tmp/` can be safely deleted.
