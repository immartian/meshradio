# Building MeshRadio

Quick guide to build MeshRadio on a fresh node.

## Prerequisites

### 1. Install Go (1.21+)

```bash
# Download and install Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
```

### 2. Install Yggdrasil

```bash
# Debian/Ubuntu
wget https://github.com/yggdrasil-network/yggdrasil-go/releases/download/v0.5.5/yggdrasil-0.5.5-amd64.deb
sudo dpkg -i yggdrasil-0.5.5-amd64.deb

# Start Yggdrasil
sudo systemctl enable yggdrasil
sudo systemctl start yggdrasil

# Get your IPv6 address
sudo yggdrasilctl getSelf
# Look for "IPv6 address" in output
```

### 3. Install PortAudio (for audio I/O)

```bash
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install -y portaudio19-dev

# Fedora/RHEL
sudo dnf install portaudio-devel

# Arch Linux
sudo pacman -S portaudio
```

### 4. Install Avahi (for mDNS discovery)

```bash
# Debian/Ubuntu
sudo apt-get install -y avahi-daemon avahi-utils

# Fedora/RHEL
sudo dnf install avahi avahi-tools

# Start Avahi
sudo systemctl enable avahi-daemon
sudo systemctl start avahi-daemon
```

## Clone Repository

```bash
git clone https://github.com/immartian/meshradio.git
cd meshradio
```

## Build All Binaries

```bash
# Build everything
go build ./...

# Build specific programs
go build ./cmd/meshradio           # Main TUI/GUI program
go build ./cmd/rtp-test            # RTP streaming test
go build ./cmd/mdns-test           # mDNS discovery test
go build ./cmd/multicast-test      # Multicast overlay test
go build ./cmd/emergency-test      # Emergency priority test
```

## Built Binaries

After building, you'll have these executables in the current directory:

```
meshradio          - Main program (TUI and Web GUI)
rtp-test          - Test RTP streaming
mdns-test         - Test mDNS service discovery
multicast-test    - Test multicast overlay (groups, SSM)
emergency-test    - Test emergency priority features
```

## Quick Test

### 1. Verify Yggdrasil is running

```bash
# Get your IPv6 address
yggdrasilctl getSelf | grep "IPv6 address"
# Should show something like: 200:1234:5678::abcd
```

### 2. Run basic audio test (loopback)

```bash
# Terminal 1 - Broadcaster
./meshradio

# In TUI:
# - Press 'b' to start broadcasting
# - Your callsign: TEST-TX
# - Port: 8799 (default)
# - Group: test

# Terminal 2 - Listener
./meshradio

# In TUI:
# - Press 'l' to listen
# - Target IPv6: <from Terminal 1>
# - Port: 8799
# - Group: test
```

### 3. Test emergency features

```bash
# Terminal 1 - Emergency broadcast
./emergency-test broadcast-critical

# Terminal 2 - Listen and watch for alerts
./emergency-test listen-manual \
  -target <ipv6-from-terminal1> \
  -port 8790 \
  -group emergency

# You should see: 🚨 CRITICAL EMERGENCY BROADCAST
```

## Common Issues

### Issue: "failed to create transport: bind: address already in use"
**Solution:** Port is already in use. Try a different port or kill the process using it:
```bash
sudo lsof -i :8790
sudo kill <PID>
```

### Issue: "Could not detect Yggdrasil IPv6"
**Solution:** Make sure Yggdrasil is running:
```bash
sudo systemctl status yggdrasil
sudo systemctl start yggdrasil
```

### Issue: "Error initializing PortAudio"
**Solution:** Install PortAudio development libraries (see Prerequisites)

### Issue: mDNS services not discovered
**Solution:**
1. Check Avahi is running: `sudo systemctl status avahi-daemon`
2. Test mDNS: `avahi-browse -a`
3. Check firewall allows mDNS (UDP port 5353)

## Running on Remote Node (Headless)

### Option 1: Web GUI Mode

```bash
# Start with web GUI
./meshradio --gui --port 8080

# Access from another machine over Yggdrasil
firefox http://[<yggdrasil-ipv6>]:8080
```

### Option 2: Daemon Mode (Future)

```bash
# Not yet implemented - see TODO.md Layer 6
# Coming soon: meshradio-daemon with gRPC/REST API
```

### Option 3: Screen/Tmux for TUI

```bash
# Install screen or tmux
sudo apt-get install screen

# Start in screen
screen -S meshradio
./meshradio
# Detach: Ctrl+A, D
# Reattach: screen -r meshradio
```

## Cross-Compilation

Build for different architectures:

```bash
# Linux ARM64 (Raspberry Pi, etc.)
GOOS=linux GOARCH=arm64 go build -o meshradio-arm64 ./cmd/meshradio

# Linux ARM (Raspberry Pi Zero, etc.)
GOOS=linux GOARCH=arm GOARM=7 go build -o meshradio-armv7 ./cmd/meshradio

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o meshradio-darwin-amd64 ./cmd/meshradio

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o meshradio-darwin-arm64 ./cmd/meshradio

# Windows
GOOS=windows GOARCH=amd64 go build -o meshradio.exe ./cmd/meshradio
```

Note: Cross-compiled binaries may need platform-specific libraries (PortAudio, Avahi) installed on target system.

## Network Setup

### Firewall Rules

MeshRadio uses these ports:

```bash
# UDP ports for RTP streams
8790-8799  # Emergency and standard channels

# TCP port for Web GUI (optional)
8080       # Or any port specified with --port

# mDNS
5353/udp   # Service discovery (usually allowed by default)
```

Example firewall rules (ufw):

```bash
sudo ufw allow 8790:8799/udp comment "MeshRadio channels"
sudo ufw allow 8080/tcp comment "MeshRadio Web GUI"
sudo ufw allow 5353/udp comment "mDNS"
```

### Yggdrasil Peering

For testing across nodes, make sure your nodes are peered:

```bash
# Node 1: Add Node 2 as peer
sudo yggdrasilctl addPeer tcp://<node2-ip>:9001

# Check connection
sudo yggdrasilctl getPeers
```

## Development Build

For development with live reload:

```bash
# Install air (live reload tool)
go install github.com/cosmtrek/air@latest

# Run with auto-rebuild
air

# Or use go run
go run ./cmd/meshradio
```

## Testing

```bash
# Run tests (when available)
go test ./...

# Run specific test
go test ./pkg/protocol

# With coverage
go test -cover ./...
```

## Next Steps

See the main README.md for:
- Usage examples
- Emergency channel guide (8790-8799)
- Protocol documentation
- Contributing guidelines

For detailed implementation status, see TODO.md
