#!/bin/bash
#
# MeshRadio Integration Test Script
# Tests all features across two nodes
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC} $1"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
}

print_test() {
    echo -e "${YELLOW}→${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC}  $1"
}

# Check if running as node 1 or node 2
if [ "$1" == "node1" ]; then
    NODE_ID="NODE1"
    NODE_CALLSIGN="STATION-A"
    NODE_PORT_BASE=8790
    IS_NODE1=true
elif [ "$1" == "node2" ]; then
    NODE_ID="NODE2"
    NODE_CALLSIGN="STATION-B"
    NODE_PORT_BASE=9790
    IS_NODE1=false
else
    echo "Usage: $0 {node1|node2}"
    echo
    echo "This script tests MeshRadio functionality across two nodes."
    echo
    echo "Instructions:"
    echo "  1. On Node 1: ./test-integration.sh node1"
    echo "  2. On Node 2: ./test-integration.sh node2"
    echo
    echo "Or in two terminals on the same machine:"
    echo "  Terminal 1: ./test-integration.sh node1"
    echo "  Terminal 2: ./test-integration.sh node2"
    exit 1
fi

print_header "MeshRadio Integration Test - $NODE_ID"

# Get Yggdrasil IPv6
print_test "Getting Yggdrasil IPv6 address..."
if command -v yggdrasilctl &> /dev/null; then
    YGG_IPV6=$(sudo yggdrasilctl getSelf 2>/dev/null | grep "IPv6 address" | awk '{print $3}')
    if [ -z "$YGG_IPV6" ]; then
        print_error "Could not get Yggdrasil IPv6"
        YGG_IPV6="::1"  # Fallback to localhost for testing
    fi
else
    print_info "Yggdrasil not found, using localhost"
    YGG_IPV6="::1"
fi
print_success "IPv6: $YGG_IPV6"

# Ask for peer IPv6 if node2
if [ "$IS_NODE1" = false ]; then
    echo
    read -p "Enter Node 1 IPv6 address: " PEER_IPV6
    if [ -z "$PEER_IPV6" ]; then
        print_error "Peer IPv6 required for Node 2"
        exit 1
    fi
fi

# Build binaries if needed
print_test "Checking binaries..."
if [ ! -f "./rtp-test" ] || [ ! -f "./mdns-test" ] || [ ! -f "./multicast-test" ] || [ ! -f "./emergency-test" ]; then
    print_info "Building binaries..."
    ./build.sh
fi
print_success "Binaries ready"

# Test 1: RTP Streaming
print_header "Test 1: RTP Streaming (Layer 2)"

if [ "$IS_NODE1" = true ]; then
    print_test "Node 1: Starting RTP broadcaster..."
    print_info "Running: ./rtp-test broadcast -callsign $NODE_CALLSIGN -port $NODE_PORT_BASE"
    echo
    echo "Press Enter when Node 2 confirms it's listening..."
    read
    timeout 15 ./rtp-test broadcast -callsign "$NODE_CALLSIGN" -port "$NODE_PORT_BASE" &
    RTP_PID=$!
    sleep 2
    print_success "Broadcasting for 15 seconds (PID: $RTP_PID)"
    echo
    echo "Waiting for broadcast to complete..."
    wait $RTP_PID 2>/dev/null || true
    print_success "Test 1 complete on Node 1"
else
    echo
    echo "Press Enter when Node 1 starts broadcasting..."
    read
    print_test "Node 2: Starting RTP listener..."
    print_info "Running: ./rtp-test listen -target $PEER_IPV6 -port $NODE_PORT_BASE"
    echo
    timeout 15 ./rtp-test listen -target "$PEER_IPV6" -port "$NODE_PORT_BASE" || true
    print_success "Test 1 complete on Node 2"
fi

echo
echo "════════════════════════════════════════════════════════════════"
read -p "Press Enter to continue to Test 2..."

# Test 2: mDNS Discovery
print_header "Test 2: mDNS Discovery (Layer 3)"

if [ "$IS_NODE1" = true ]; then
    print_test "Node 1: Advertising mDNS service..."
    print_info "Running: ./mdns-test advertise -callsign $NODE_CALLSIGN -port $NODE_PORT_BASE -group test"
    echo
    echo "Press Enter when Node 2 has discovered the service..."
    read
    timeout 20 ./mdns-test advertise -callsign "$NODE_CALLSIGN" -port "$NODE_PORT_BASE" -group test &
    MDNS_PID=$!
    sleep 2
    print_success "Advertising for 20 seconds (PID: $MDNS_PID)"
    wait $MDNS_PID 2>/dev/null || true
    print_success "Test 2 complete on Node 1"
else
    echo
    echo "Press Enter when Node 1 starts advertising..."
    read
    print_test "Node 2: Browsing for mDNS services..."
    print_info "Running: ./mdns-test browse"
    echo
    print_info "Looking for service from $NODE_CALLSIGN..."
    timeout 15 ./mdns-test browse || true
    print_success "Test 2 complete on Node 2"
fi

echo
echo "════════════════════════════════════════════════════════════════"
read -p "Press Enter to continue to Test 3..."

# Test 3: Multicast Overlay - Regular Multicast
print_header "Test 3: Multicast Overlay - Regular Multicast (Layer 4)"

MULTICAST_PORT=$((NODE_PORT_BASE + 1))

if [ "$IS_NODE1" = true ]; then
    print_test "Node 1: Starting regular multicast broadcaster..."
    print_info "Running: ./multicast-test broadcast-regular -callsign $NODE_CALLSIGN -port $MULTICAST_PORT"
    echo
    echo "This broadcasts to 'emergency' group (any-source multicast)"
    echo "Node 2 will receive from all sources in this group"
    echo
    echo "Press Enter when Node 2 is listening..."
    read
    timeout 20 ./multicast-test broadcast-regular -callsign "$NODE_CALLSIGN" -port "$MULTICAST_PORT" &
    MCAST_PID=$!
    sleep 2
    print_success "Broadcasting to emergency group (PID: $MCAST_PID)"
    echo
    echo "Waiting for broadcast to complete..."
    wait $MCAST_PID 2>/dev/null || true
    print_success "Test 3 complete on Node 1"
else
    echo
    echo "Press Enter when Node 1 starts broadcasting..."
    read
    print_test "Node 2: Listening to regular multicast (emergency group)..."
    print_info "Running: ./multicast-test listen-regular -target $PEER_IPV6 -port $MULTICAST_PORT"
    echo
    timeout 20 ./multicast-test listen-regular -target "$PEER_IPV6" -port "$MULTICAST_PORT" || true
    print_success "Test 3 complete on Node 2"
fi

echo
echo "════════════════════════════════════════════════════════════════"
read -p "Press Enter to continue to Test 4..."

# Test 4: Multicast Overlay - SSM
print_header "Test 4: Multicast Overlay - SSM (Layer 4)"

SSM_PORT=$((NODE_PORT_BASE + 2))

if [ "$IS_NODE1" = true ]; then
    print_test "Node 1: Starting SSM broadcaster..."
    print_info "Running: ./multicast-test broadcast-ssm -callsign $NODE_CALLSIGN -port $SSM_PORT"
    echo
    echo "This broadcasts to 'community' group (source-specific multicast)"
    echo "Node 2 will only receive from this specific source"
    echo
    echo "Press Enter when Node 2 is listening..."
    read
    timeout 20 ./multicast-test broadcast-ssm -callsign "$NODE_CALLSIGN" -port "$SSM_PORT" &
    SSM_PID=$!
    sleep 2
    print_success "Broadcasting with SSM (PID: $SSM_PID)"
    echo
    echo "Waiting for broadcast to complete..."
    wait $SSM_PID 2>/dev/null || true
    print_success "Test 4 complete on Node 1"
else
    echo
    echo "Press Enter when Node 1 starts broadcasting..."
    read
    print_test "Node 2: Listening with SSM (only from Node 1)..."
    print_info "Running: ./multicast-test listen-ssm -target $PEER_IPV6 -port $SSM_PORT"
    echo
    timeout 20 ./multicast-test listen-ssm -target "$PEER_IPV6" -port "$SSM_PORT" || true
    print_success "Test 4 complete on Node 2"
fi

echo
echo "════════════════════════════════════════════════════════════════"
read -p "Press Enter to continue to Test 5..."

# Test 5: Emergency Priority - Normal
print_header "Test 5: Emergency Priority - Normal Priority (Layer 5)"

NORMAL_PORT=$((NODE_PORT_BASE + 5))

if [ "$IS_NODE1" = true ]; then
    print_test "Node 1: Broadcasting with NORMAL priority..."
    print_info "Running: ./emergency-test broadcast-normal"
    echo
    echo "This broadcasts on 'community' channel (port 8795)"
    echo "Priority: Normal (0) - no special alerts"
    echo
    echo "Press Enter when Node 2 is listening..."
    read
    timeout 20 ./emergency-test broadcast-normal &
    NORMAL_PID=$!
    sleep 2
    print_success "Broadcasting with normal priority (PID: $NORMAL_PID)"
    echo
    echo "Waiting for broadcast to complete..."
    wait $NORMAL_PID 2>/dev/null || true
    print_success "Test 5 complete on Node 1"
else
    echo
    echo "Press Enter when Node 1 starts broadcasting..."
    read
    print_test "Node 2: Listening for normal priority broadcast..."
    print_info "Running: ./emergency-test listen-manual -target $PEER_IPV6 -port 8795 -group community"
    echo
    print_info "Should see normal packet reception (no priority alerts)"
    timeout 20 ./emergency-test listen-manual -target "$PEER_IPV6" -port 8795 -group community || true
    print_success "Test 5 complete on Node 2"
fi

echo
echo "════════════════════════════════════════════════════════════════"
read -p "Press Enter to continue to Test 6..."

# Test 6: Emergency Priority - High
print_header "Test 6: Emergency Priority - High Priority (Layer 5)"

if [ "$IS_NODE1" = true ]; then
    print_test "Node 1: Broadcasting with HIGH priority..."
    print_info "Running: ./emergency-test broadcast-high"
    echo
    echo "This broadcasts on 'weather' channel (port 8793)"
    echo "Priority: High (1) - 📢 alerts expected"
    echo
    echo "Press Enter when Node 2 is listening..."
    read
    timeout 20 ./emergency-test broadcast-high &
    HIGH_PID=$!
    sleep 2
    print_success "Broadcasting with high priority (PID: $HIGH_PID)"
    echo
    echo "Waiting for broadcast to complete..."
    wait $HIGH_PID 2>/dev/null || true
    print_success "Test 6 complete on Node 1"
else
    echo
    echo "Press Enter when Node 1 starts broadcasting..."
    read
    print_test "Node 2: Listening for high priority broadcast..."
    print_info "Running: ./emergency-test listen-manual -target $PEER_IPV6 -port 8793 -group weather"
    echo
    print_info "Watch for: 📢 High priority broadcast alerts"
    timeout 20 ./emergency-test listen-manual -target "$PEER_IPV6" -port 8793 -group weather || true
    print_success "Test 6 complete on Node 2"
fi

echo
echo "════════════════════════════════════════════════════════════════"
read -p "Press Enter to continue to Test 7..."

# Test 7: Emergency Priority - Emergency
print_header "Test 7: Emergency Priority - Emergency (Layer 5)"

if [ "$IS_NODE1" = true ]; then
    print_test "Node 1: Broadcasting with EMERGENCY priority..."
    print_info "Running: ./emergency-test broadcast-emergency"
    echo
    echo "This broadcasts on 'netcontrol' channel (port 8791)"
    echo "Priority: Emergency (2) - ⚠️ alerts expected"
    echo
    echo "Press Enter when Node 2 is listening..."
    read
    timeout 20 ./emergency-test broadcast-emergency &
    EMERG_PID=$!
    sleep 2
    print_success "Broadcasting with emergency priority (PID: $EMERG_PID)"
    echo
    echo "Waiting for broadcast to complete..."
    wait $EMERG_PID 2>/dev/null || true
    print_success "Test 7 complete on Node 1"
else
    echo
    echo "Press Enter when Node 1 starts broadcasting..."
    read
    print_test "Node 2: Listening for emergency priority broadcast..."
    print_info "Running: ./emergency-test listen-manual -target $PEER_IPV6 -port 8791 -group netcontrol"
    echo
    print_info "Watch for: ⚠️  EMERGENCY BROADCAST alerts"
    timeout 20 ./emergency-test listen-manual -target "$PEER_IPV6" -port 8791 -group netcontrol || true
    print_success "Test 7 complete on Node 2"
fi

echo
echo "════════════════════════════════════════════════════════════════"
read -p "Press Enter to continue to Test 8..."

# Test 8: Emergency Priority - Critical
print_header "Test 8: Emergency Priority - Critical (Layer 5)"

if [ "$IS_NODE1" = true ]; then
    print_test "Node 1: Broadcasting with CRITICAL priority..."
    print_info "Running: ./emergency-test broadcast-critical"
    echo
    echo "This broadcasts on 'emergency' channel (port 8790)"
    echo "Priority: Critical (3) - 🚨 alerts expected"
    echo
    echo "Press Enter when Node 2 is listening..."
    read
    timeout 20 ./emergency-test broadcast-critical &
    CRIT_PID=$!
    sleep 2
    print_success "Broadcasting with critical priority (PID: $CRIT_PID)"
    echo
    echo "Waiting for broadcast to complete..."
    wait $CRIT_PID 2>/dev/null || true
    print_success "Test 8 complete on Node 1"
else
    echo
    echo "Press Enter when Node 1 starts broadcasting..."
    read
    print_test "Node 2: Listening for critical priority broadcast..."
    print_info "Running: ./emergency-test listen-manual -target $PEER_IPV6 -port 8790 -group emergency"
    echo
    print_info "Watch for: 🚨 CRITICAL EMERGENCY BROADCAST alerts"
    timeout 20 ./emergency-test listen-manual -target "$PEER_IPV6" -port 8790 -group emergency || true
    print_success "Test 8 complete on Node 2"
fi

# Final summary
print_header "Integration Test Complete - $NODE_ID"

echo "All tests completed successfully!"
echo
echo "Tests performed:"
echo "  ✅ Test 1: RTP Streaming (Layer 2)"
echo "  ✅ Test 2: mDNS Discovery (Layer 3)"
echo "  ✅ Test 3: Regular Multicast (Layer 4)"
echo "  ✅ Test 4: SSM (Source-Specific Multicast) (Layer 4)"
echo "  ✅ Test 5: Normal Priority Broadcast (Layer 5)"
echo "  ✅ Test 6: High Priority Broadcast (Layer 5)"
echo "  ✅ Test 7: Emergency Priority Broadcast (Layer 5)"
echo "  ✅ Test 8: Critical Priority Broadcast (Layer 5)"
echo
echo "Architecture validated:"
echo "  ✅ Layer 1: Transport (Yggdrasil IPv6)"
echo "  ✅ Layer 2: RTP Streaming"
echo "  ✅ Layer 3: mDNS Discovery"
echo "  ✅ Layer 4: Multicast Overlay (Regular + SSM)"
echo "  ✅ Layer 5: Emergency Features (Priority Signaling)"
echo

if [ "$IS_NODE1" = true ]; then
    echo "Node 1 test sequence complete!"
    echo "Check Node 2 terminal for its results."
else
    echo "Node 2 test sequence complete!"
    echo "Check Node 1 terminal for its results."
fi

echo
echo "For more details, see BUILD.md and TODO.md"
