#!/bin/bash
# Install audio dependencies for MeshRadio

set -e

echo "🎵 Installing MeshRadio Audio Dependencies"
echo "=========================================="
echo

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo "❌ Cannot detect OS"
    exit 1
fi

echo "Detected OS: $OS"
echo

case $OS in
    ubuntu|debian)
        echo "📦 Installing PortAudio and Opus for Ubuntu/Debian..."
        sudo apt-get update
        sudo apt-get install -y portaudio19-dev libopus-dev
        ;;

    fedora|rhel|centos)
        echo "📦 Installing PortAudio and Opus for Fedora/RHEL..."
        sudo dnf install -y portaudio-devel opus-devel
        ;;

    arch|manjaro)
        echo "📦 Installing PortAudio and Opus for Arch..."
        sudo pacman -S --noconfirm portaudio opus
        ;;

    *)
        echo "❌ Unsupported OS: $OS"
        echo "Please install manually:"
        echo "  - PortAudio development files"
        echo "  - Opus development files"
        exit 1
        ;;
esac

echo
echo "✅ Audio libraries installed successfully!"
echo
echo "Next steps:"
echo "  1. Rebuild MeshRadio: make clean && make build"
echo "  2. Test with real audio: ./meshradio"
echo

# Verify installation
echo "🔍 Verifying installation..."
if pkg-config --exists portaudio-2.0; then
    echo "✅ PortAudio: $(pkg-config --modversion portaudio-2.0)"
else
    echo "⚠️  PortAudio not detected by pkg-config"
fi

if pkg-config --exists opus; then
    echo "✅ Opus: $(pkg-config --modversion opus)"
else
    echo "⚠️  Opus not detected by pkg-config"
fi

echo
echo "🎉 Ready for real audio!"
