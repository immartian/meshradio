# Changelog

All notable changes to MeshRadio will be documented in this file.

## [Unreleased]

### Added
- Real Yggdrasil IPv6 detection via yggdrasilctl
- Fallback IPv6 detection from network interfaces
- Real network transmission - broadcaster sends to multicast
- Beacon broadcasting with station info
- Enhanced UI with real-time updates
- Animated status indicators (dots)
- Signal strength visualization
- Audio level meters (simulated)
- Dependency checker script
- Better error messages and warnings

### Changed
- Broadcaster now actually transmits packets via UDP multicast
- Listener receives on correct port
- UI updates every second for live stats
- Main menu shows Yggdrasil connection status
- Improved broadcast view with network info
- Improved listen view with signal quality

### Fixed
- Network transmission now works (was prepared but not sent)
- IPv6 detection works with real Yggdrasil daemon
- Port binding issues resolved

## [0.1.0-alpha] - 2025-11-25

### Added
- Initial MVP release
- Complete protocol implementation
- Audio streaming pipeline (simulated)
- Broadcaster and Listener components
- Cross-platform TUI with Bubbletea
- Network transport layer
- Comprehensive design documentation
- Bootstrap strategy
- GitHub repository created

### Known Limitations
- Audio I/O is simulated (no real microphone/speakers yet)
- Codec is pass-through (no compression yet)
- No scanning or discovery features
- Single connection at a time
- Hardcoded multicast address
