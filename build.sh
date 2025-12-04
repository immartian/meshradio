#!/bin/bash
#
# MeshRadio Build Script
# Builds all binaries for the current platform
#

set -e  # Exit on error

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║ MeshRadio Build Script                                       ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed"
    echo "   Please install Go 1.21+ from https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "✅ Go found: $GO_VERSION"

# Check for dependencies
echo
echo "Checking dependencies..."

# Check PortAudio
if pkg-config --exists portaudio-2.0; then
    echo "✅ PortAudio found"
else
    echo "⚠️  Warning: PortAudio not found"
    echo "   Audio I/O will not work. Install with:"
    echo "   - Debian/Ubuntu: sudo apt-get install portaudio19-dev"
    echo "   - Fedora/RHEL: sudo dnf install portaudio-devel"
    echo "   - Arch: sudo pacman -S portaudio"
fi

# Check Yggdrasil
if command -v yggdrasil &> /dev/null || command -v yggdrasilctl &> /dev/null; then
    echo "✅ Yggdrasil found"
else
    echo "⚠️  Warning: Yggdrasil not found"
    echo "   Install from: https://yggdrasil-network.github.io/installation.html"
fi

# Check Avahi (for mDNS)
if command -v avahi-daemon &> /dev/null || systemctl is-active --quiet avahi-daemon; then
    echo "✅ Avahi (mDNS) found"
else
    echo "⚠️  Warning: Avahi not found"
    echo "   Service discovery may not work. Install with:"
    echo "   - Debian/Ubuntu: sudo apt-get install avahi-daemon"
    echo "   - Fedora/RHEL: sudo dnf install avahi"
fi

echo
echo "Building binaries..."
echo

# Build all programs
PROGRAMS=(
    "cmd/meshradio:meshradio:Main TUI/GUI program"
    "cmd/rtp-test:rtp-test:RTP streaming test"
    "cmd/mdns-test:mdns-test:mDNS discovery test"
    "cmd/multicast-test:multicast-test:Multicast overlay test"
    "cmd/emergency-test:emergency-test:Emergency priority test"
)

SUCCESS=0
FAILED=0

for prog in "${PROGRAMS[@]}"; do
    IFS=':' read -r path binary desc <<< "$prog"

    echo -n "Building $binary... "
    if go build -o "$binary" "./$path" 2>/dev/null; then
        echo "✅ $desc"
        SUCCESS=$((SUCCESS + 1))
    else
        echo "❌ Failed"
        FAILED=$((FAILED + 1))
    fi
done

echo
echo "═══════════════════════════════════════════════════════════════"
echo "Build Summary: $SUCCESS succeeded, $FAILED failed"
echo "═══════════════════════════════════════════════════════════════"

if [ $FAILED -eq 0 ]; then
    echo
    echo "✅ All binaries built successfully!"
    echo
    echo "Available programs:"
    echo "  ./meshradio        - Main program (TUI and Web GUI)"
    echo "  ./rtp-test         - Test RTP streaming"
    echo "  ./mdns-test        - Test mDNS discovery"
    echo "  ./multicast-test   - Test multicast overlay"
    echo "  ./emergency-test   - Test emergency features"
    echo
    echo "Quick start:"
    echo "  ./meshradio --help              # Show help"
    echo "  ./meshradio                     # Start TUI"
    echo "  ./meshradio --gui --port 8080   # Start Web GUI"
    echo
    echo "For detailed instructions, see BUILD.md"
    exit 0
else
    echo
    echo "⚠️  Some builds failed. Check error messages above."
    echo "   See BUILD.md for dependency installation instructions."
    exit 1
fi
