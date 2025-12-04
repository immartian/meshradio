#!/bin/bash
#
# MeshRadio Quick Test Script
# Runs automated tests on a single machine
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

cleanup() {
    print_info "Cleaning up background processes..."
    jobs -p | xargs -r kill 2>/dev/null || true
}

trap cleanup EXIT

print_header "MeshRadio Quick Test Suite"

# Check binaries
print_test "Checking binaries..."
if [ ! -f "./rtp-test" ] || [ ! -f "./emergency-test" ]; then
    print_info "Building binaries..."
    ./build.sh || exit 1
fi
print_success "Binaries ready"

# Use localhost for testing
IPV6="::1"
print_info "Testing on localhost: $IPV6"

PASSED=0
FAILED=0

# Test 1: RTP Streaming
print_header "Test 1: RTP Streaming"
print_test "Starting RTP broadcaster..."

./rtp-test broadcast -callsign TEST-TX -port 8799 > /tmp/test-rtp-tx.log 2>&1 &
TX_PID=$!
sleep 2

print_test "Starting RTP listener..."
timeout 10 ./rtp-test listen -target "$IPV6" -port 8799 > /tmp/test-rtp-rx.log 2>&1 &
RX_PID=$!
sleep 8

kill $TX_PID 2>/dev/null || true
wait $RX_PID 2>/dev/null || true

if grep -q "Received packet" /tmp/test-rtp-rx.log && grep -q "Broadcasting" /tmp/test-rtp-tx.log; then
    print_success "Test 1 PASSED: RTP packets transmitted and received"
    PASSED=$((PASSED + 1))
else
    print_error "Test 1 FAILED: No RTP packets detected"
    FAILED=$((FAILED + 1))
fi

# Test 2: Emergency Priority - Normal
print_header "Test 2: Emergency Priority - Normal"
print_test "Starting normal priority broadcaster..."

./emergency-test broadcast-normal > /tmp/test-normal-tx.log 2>&1 &
NORMAL_TX_PID=$!
sleep 2

print_test "Starting listener..."
timeout 10 ./emergency-test listen-manual -target "$IPV6" -port 8795 -group community > /tmp/test-normal-rx.log 2>&1 &
NORMAL_RX_PID=$!
sleep 8

kill $NORMAL_TX_PID 2>/dev/null || true
wait $NORMAL_RX_PID 2>/dev/null || true

if grep -q "Received:" /tmp/test-normal-rx.log && ! grep -q "EMERGENCY" /tmp/test-normal-rx.log; then
    print_success "Test 2 PASSED: Normal priority (no alerts)"
    PASSED=$((PASSED + 1))
else
    print_error "Test 2 FAILED: Normal priority test failed"
    FAILED=$((FAILED + 1))
fi

# Test 3: Emergency Priority - High
print_header "Test 3: Emergency Priority - High"
print_test "Starting high priority broadcaster..."

./emergency-test broadcast-high > /tmp/test-high-tx.log 2>&1 &
HIGH_TX_PID=$!
sleep 2

print_test "Starting listener..."
timeout 10 ./emergency-test listen-manual -target "$IPV6" -port 8793 -group weather > /tmp/test-high-rx.log 2>&1 &
HIGH_RX_PID=$!
sleep 8

kill $HIGH_TX_PID 2>/dev/null || true
wait $HIGH_RX_PID 2>/dev/null || true

if grep -q "High priority broadcast" /tmp/test-high-rx.log; then
    print_success "Test 3 PASSED: High priority detected (📢)"
    PASSED=$((PASSED + 1))
else
    print_error "Test 3 FAILED: High priority alert not detected"
    FAILED=$((FAILED + 1))
fi

# Test 4: Emergency Priority - Emergency
print_header "Test 4: Emergency Priority - Emergency"
print_test "Starting emergency priority broadcaster..."

./emergency-test broadcast-emergency > /tmp/test-emergency-tx.log 2>&1 &
EMERG_TX_PID=$!
sleep 2

print_test "Starting listener..."
timeout 10 ./emergency-test listen-manual -target "$IPV6" -port 8791 -group netcontrol > /tmp/test-emergency-rx.log 2>&1 &
EMERG_RX_PID=$!
sleep 8

kill $EMERG_TX_PID 2>/dev/null || true
wait $EMERG_RX_PID 2>/dev/null || true

if grep -q "EMERGENCY BROADCAST" /tmp/test-emergency-rx.log; then
    print_success "Test 4 PASSED: Emergency priority detected (⚠️)"
    PASSED=$((PASSED + 1))
else
    print_error "Test 4 FAILED: Emergency priority alert not detected"
    FAILED=$((FAILED + 1))
fi

# Test 5: Emergency Priority - Critical
print_header "Test 5: Emergency Priority - Critical"
print_test "Starting critical priority broadcaster..."

./emergency-test broadcast-critical > /tmp/test-critical-tx.log 2>&1 &
CRIT_TX_PID=$!
sleep 2

print_test "Starting listener..."
timeout 10 ./emergency-test listen-manual -target "$IPV6" -port 8790 -group emergency > /tmp/test-critical-rx.log 2>&1 &
CRIT_RX_PID=$!
sleep 8

kill $CRIT_TX_PID 2>/dev/null || true
wait $CRIT_RX_PID 2>/dev/null || true

if grep -q "CRITICAL EMERGENCY" /tmp/test-critical-rx.log; then
    print_success "Test 5 PASSED: Critical priority detected (🚨)"
    PASSED=$((PASSED + 1))
else
    print_error "Test 5 FAILED: Critical priority alert not detected"
    FAILED=$((FAILED + 1))
fi

# Test 6: Protocol Priority Encoding
print_header "Test 6: Protocol Priority Encoding"
print_test "Checking priority encoding in packets..."

if grep -q "Priority: critical" /tmp/test-critical-tx.log && \
   grep -q "Priority: emergency" /tmp/test-emergency-tx.log && \
   grep -q "Priority: high" /tmp/test-high-tx.log && \
   grep -q "Priority: normal" /tmp/test-normal-tx.log; then
    print_success "Test 6 PASSED: All priority levels encoded correctly"
    PASSED=$((PASSED + 1))
else
    print_error "Test 6 FAILED: Priority encoding issue"
    FAILED=$((FAILED + 1))
fi

# Test 7: Subscription Management
print_header "Test 7: Subscription Management"
print_test "Checking subscription flow..."

if grep -q "Sent SUBSCRIBE" /tmp/test-critical-rx.log && \
   grep -q "New subscriber" /tmp/test-critical-tx.log; then
    print_success "Test 7 PASSED: Subscription protocol works"
    PASSED=$((PASSED + 1))
else
    print_error "Test 7 FAILED: Subscription flow issue"
    FAILED=$((FAILED + 1))
fi

# Final Summary
print_header "Test Summary"

echo "═══════════════════════════════════════════════════════════════"
echo -e "  ${GREEN}Passed:${NC} $PASSED"
echo -e "  ${RED}Failed:${NC} $FAILED"
echo "═══════════════════════════════════════════════════════════════"
echo

if [ $FAILED -eq 0 ]; then
    print_success "ALL TESTS PASSED! ✅"
    echo
    echo "MeshRadio core functionality verified:"
    echo "  ✅ Layer 2: RTP Streaming"
    echo "  ✅ Layer 4: Subscription Management"
    echo "  ✅ Layer 5: Emergency Priority (Normal)"
    echo "  ✅ Layer 5: Emergency Priority (High)"
    echo "  ✅ Layer 5: Emergency Priority (Emergency)"
    echo "  ✅ Layer 5: Emergency Priority (Critical)"
    echo "  ✅ Protocol: Priority Encoding"
    echo
    echo "Run './test-integration.sh' for full two-node testing."
    echo "See TESTING.md for detailed test procedures."
    exit 0
else
    print_error "SOME TESTS FAILED ❌"
    echo
    echo "Check logs in /tmp/test-*.log for details"
    echo "See TESTING.md for troubleshooting"
    exit 1
fi
