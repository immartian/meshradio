#!/bin/bash
# Dependency checker for MeshRadio

set -e

echo "🔍 Checking MeshRadio dependencies..."
echo

# Check Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo "✅ Go: $GO_VERSION"
else
    echo "❌ Go: Not found"
    echo "   Install from: https://go.dev/dl/"
    exit 1
fi

# Check Yggdrasil
if command -v yggdrasilctl &> /dev/null; then
    echo "✅ Yggdrasil: Installed"
    if yggdrasilctl getSelf &> /dev/null; then
        echo "   ✅ Yggdrasil daemon: Running"
    else
        echo "   ⚠️  Yggdrasil daemon: Not running"
        echo "      Start with: sudo systemctl start yggdrasil"
    fi
else
    echo "⚠️  Yggdrasil: Not found (optional for testing)"
    echo "   Install from: https://yggdrasil-network.github.io/"
fi

# Check PortAudio
if pkg-config --exists portaudio-2.0 2>/dev/null; then
    PA_VERSION=$(pkg-config --modversion portaudio-2.0)
    echo "✅ PortAudio: v$PA_VERSION"
else
    echo "❌ PortAudio: Not found (required for real audio)"
    echo "   Install:"
    if [ -f /etc/debian_version ]; then
        echo "   sudo apt-get install portaudio19-dev"
    elif [ -f /etc/redhat-release ]; then
        echo "   sudo dnf install portaudio-devel"
    elif [ -f /etc/arch-release ]; then
        echo "   sudo pacman -S portaudio"
    else
        echo "   Check your package manager for portaudio development files"
    fi
fi

# Check Opus
if pkg-config --exists opus 2>/dev/null; then
    OPUS_VERSION=$(pkg-config --modversion opus)
    echo "✅ Opus codec: v$OPUS_VERSION"
else
    echo "❌ Opus codec: Not found (required for audio compression)"
    echo "   Install:"
    if [ -f /etc/debian_version ]; then
        echo "   sudo apt-get install libopus-dev"
    elif [ -f /etc/redhat-release ]; then
        echo "   sudo dnf install opus-devel"
    elif [ -f /etc/arch-release ]; then
        echo "   sudo pacman -S opus"
    else
        echo "   Check your package manager for opus development files"
    fi
fi

echo
echo "📦 Summary:"
echo "   Core dependencies (Go): Ready"
echo "   Network (Yggdrasil): $(command -v yggdrasilctl &> /dev/null && echo 'Ready' || echo 'Not installed')"
echo "   Audio I/O (PortAudio): $(pkg-config --exists portaudio-2.0 && echo 'Ready' || echo 'Missing')"
echo "   Audio Codec (Opus): $(pkg-config --exists opus && echo 'Ready' || echo 'Missing')"
echo
echo "💡 MeshRadio will work without audio dependencies but with simulated audio only."
