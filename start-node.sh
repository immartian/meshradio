#!/bin/bash
#
# MeshRadio Node Setup Script
# Quickly configure a node as broadcaster, music station, or listener
#

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_header() {
    echo
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC} $1"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
}

print_info() {
    echo -e "${CYAN}ℹ${NC}  $1"
}

print_success() {
    echo -e "${GREEN}✅${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

print_header "MeshRadio Node Configuration"

# Check if binaries exist
if [ ! -f "./meshradio" ]; then
    print_error "Binaries not found. Building..."
    ./build.sh
    echo
fi

# Get IPv6
print_info "Detecting Yggdrasil IPv6..."
if command -v yggdrasilctl &> /dev/null; then
    IPV6=$(yggdrasilctl getSelf | grep "IPv6 address" | awk '{print $3}')
    if [ -n "$IPV6" ]; then
        print_success "Your IPv6: $IPV6"
    else
        print_info "Using localhost (::1) for testing"
        IPV6="::1"
    fi
else
    print_info "Yggdrasil not found, using localhost"
    IPV6="::1"
fi
echo

# Show menu
echo "Select node type:"
echo
echo "  ${GREEN}1)${NC} Music Broadcaster  - Scan and broadcast MP3 files"
echo "  ${GREEN}2)${NC} Voice Broadcaster  - Broadcast from microphone"
echo "  ${GREEN}3)${NC} Listener          - Listen to broadcasts"
echo "  ${GREEN}4)${NC} Emergency Test    - Test emergency priorities"
echo "  ${GREEN}5)${NC} Discovery Test    - Test mDNS service discovery"
echo "  ${GREEN}6)${NC} Integration Test  - Full two-node test"
echo
echo "  ${YELLOW}0)${NC} Exit"
echo

read -p "Choice [1-6]: " choice

case $choice in
    1)
        print_header "Music Broadcaster Setup"

        # Get music directory
        DEFAULT_MUSIC="$HOME/Music"
        echo "Music directory (default: $DEFAULT_MUSIC):"
        read -p "> " MUSIC_DIR
        MUSIC_DIR=${MUSIC_DIR:-$DEFAULT_MUSIC}

        # Check if directory exists
        if [ ! -d "$MUSIC_DIR" ]; then
            print_error "Directory not found: $MUSIC_DIR"
            exit 1
        fi

        # Get callsign
        echo
        echo "Station callsign (default: MUSIC-DJ):"
        read -p "> " CALLSIGN
        CALLSIGN=${CALLSIGN:-MUSIC-DJ}

        # Get channel
        echo
        echo "Channel (default: talk):"
        echo "  Options: talk, community, test"
        read -p "> " CHANNEL
        CHANNEL=${CHANNEL:-talk}

        # Get port
        case $CHANNEL in
            talk)
                PORT=8798
                ;;
            community)
                PORT=8795
                ;;
            test)
                PORT=8799
                ;;
            *)
                PORT=8798
                ;;
        esac

        echo
        print_header "Starting Music Station"
        echo "Callsign: $CALLSIGN"
        echo "Channel:  $CHANNEL"
        echo "Port:     $PORT"
        echo "Music:    $MUSIC_DIR"
        echo "IPv6:     $IPV6"
        echo
        print_info "Share this with listeners: $IPV6:$PORT"
        echo
        read -p "Press Enter to start..."

        ./music-broadcast \
            --callsign "$CALLSIGN" \
            --dir "$MUSIC_DIR" \
            --group "$CHANNEL" \
            --port "$PORT"
        ;;

    2)
        print_header "Voice Broadcaster Setup"

        # Get callsign
        echo "Station callsign (default: STATION-TX):"
        read -p "> " CALLSIGN
        CALLSIGN=${CALLSIGN:-STATION-TX}

        # Get channel
        echo
        echo "Channel (default: talk):"
        echo "  Options: talk, community, test, emergency"
        read -p "> " CHANNEL
        CHANNEL=${CHANNEL:-talk}

        # Get port based on channel
        case $CHANNEL in
            emergency)
                PORT=8790
                ;;
            netcontrol)
                PORT=8791
                ;;
            medical)
                PORT=8792
                ;;
            weather)
                PORT=8793
                ;;
            sar)
                PORT=8794
                ;;
            community)
                PORT=8795
                ;;
            talk)
                PORT=8798
                ;;
            test)
                PORT=8799
                ;;
            *)
                PORT=8798
                ;;
        esac

        echo
        print_header "Starting Voice Broadcaster"
        echo "Callsign: $CALLSIGN"
        echo "Channel:  $CHANNEL"
        echo "Port:     $PORT"
        echo "IPv6:     $IPV6"
        echo
        print_info "Share this with listeners: $IPV6:$PORT"
        echo

        # Check if emergency channel
        if [ "$CHANNEL" = "emergency" ] || [ "$CHANNEL" = "netcontrol" ] || [ "$CHANNEL" = "medical" ] || [ "$CHANNEL" = "sar" ]; then
            print_info "Using emergency channel - high priority broadcast"
            read -p "Press Enter to start EMERGENCY broadcast..."

            case $CHANNEL in
                emergency)
                    ./emergency-test broadcast-critical
                    ;;
                netcontrol|medical|sar)
                    ./emergency-test broadcast-emergency
                    ;;
            esac
        else
            read -p "Press Enter to start..."
            ./rtp-test broadcast -callsign "$CALLSIGN" -port "$PORT"
        fi
        ;;

    3)
        print_header "Listener Setup"

        # Get target IPv6
        echo "Broadcaster IPv6 address:"
        echo "  (Get this from the broadcaster's terminal)"
        read -p "> " TARGET_IPV6

        if [ -z "$TARGET_IPV6" ]; then
            print_error "IPv6 address required"
            exit 1
        fi

        # Get port
        echo
        echo "Broadcaster port (default: 8798):"
        echo "  8790 = emergency"
        echo "  8791 = netcontrol"
        echo "  8795 = community"
        echo "  8798 = talk"
        echo "  8799 = test"
        read -p "> " PORT
        PORT=${PORT:-8798}

        # Get group based on port
        case $PORT in
            8790)
                GROUP="emergency"
                ;;
            8791)
                GROUP="netcontrol"
                ;;
            8792)
                GROUP="medical"
                ;;
            8793)
                GROUP="weather"
                ;;
            8794)
                GROUP="sar"
                ;;
            8795)
                GROUP="community"
                ;;
            8798)
                GROUP="talk"
                ;;
            8799)
                GROUP="test"
                ;;
            *)
                GROUP="talk"
                ;;
        esac

        echo
        print_header "Starting Listener"
        echo "Target:  $TARGET_IPV6:$PORT"
        echo "Group:   $GROUP"
        echo
        print_info "Watch for priority alerts (🚨 ⚠️ 📢)"
        echo
        read -p "Press Enter to start listening..."

        ./emergency-test listen-manual \
            -target "$TARGET_IPV6" \
            -port "$PORT" \
            -group "$GROUP"
        ;;

    4)
        print_header "Emergency Priority Test"

        echo "Select priority level:"
        echo
        echo "  ${GREEN}1)${NC} Normal     - No alerts"
        echo "  ${GREEN}2)${NC} High       - 📢 High priority"
        echo "  ${GREEN}3)${NC} Emergency  - ⚠️  Emergency"
        echo "  ${GREEN}4)${NC} Critical   - 🚨 Critical"
        echo
        read -p "Choice [1-4]: " PRIORITY

        case $PRIORITY in
            1)
                echo
                print_info "Broadcasting NORMAL priority on port 8795"
                ./emergency-test broadcast-normal
                ;;
            2)
                echo
                print_info "Broadcasting HIGH priority on port 8793"
                ./emergency-test broadcast-high
                ;;
            3)
                echo
                print_info "Broadcasting EMERGENCY priority on port 8791"
                ./emergency-test broadcast-emergency
                ;;
            4)
                echo
                print_info "Broadcasting CRITICAL priority on port 8790"
                ./emergency-test broadcast-critical
                ;;
            *)
                print_error "Invalid choice"
                exit 1
                ;;
        esac
        ;;

    5)
        print_header "Service Discovery Test"

        echo "Select mode:"
        echo
        echo "  ${GREEN}1)${NC} Advertise - Make this station discoverable"
        echo "  ${GREEN}2)${NC} Browse    - Find other stations"
        echo
        read -p "Choice [1-2]: " MODE

        case $MODE in
            1)
                echo
                echo "Station callsign (default: BEACON-1):"
                read -p "> " CALLSIGN
                CALLSIGN=${CALLSIGN:-BEACON-1}

                echo
                echo "Channel (default: talk):"
                read -p "> " CHANNEL
                CHANNEL=${CHANNEL:-talk}

                PORT=8798

                echo
                print_header "Advertising Station"
                echo "Callsign: $CALLSIGN"
                echo "Channel:  $CHANNEL"
                echo "Port:     $PORT"
                echo "IPv6:     $IPV6"
                echo
                print_info "Other nodes can now discover this station"
                echo

                ./mdns-test advertise -callsign "$CALLSIGN" -port "$PORT" -group "$CHANNEL"
                ;;
            2)
                echo
                print_header "Browsing for Stations"
                print_info "Looking for MeshRadio stations on the network..."
                echo

                ./mdns-test browse
                ;;
            *)
                print_error "Invalid choice"
                exit 1
                ;;
        esac
        ;;

    6)
        print_header "Integration Test"

        echo "This runs the full two-node integration test."
        echo "Run this on TWO machines (or two terminals)."
        echo
        echo "Select node role:"
        echo
        echo "  ${GREEN}1)${NC} Node 1 (Broadcaster)"
        echo "  ${GREEN}2)${NC} Node 2 (Listener)"
        echo
        read -p "Choice [1-2]: " NODE

        case $NODE in
            1)
                echo
                print_header "Starting Integration Test - Node 1"
                print_info "Your IPv6: $IPV6"
                echo
                print_info "On Node 2, run: ./start-node.sh and select option 6, then Node 2"
                echo
                read -p "Press Enter when ready..."

                ./test-integration.sh node1
                ;;
            2)
                echo
                print_header "Starting Integration Test - Node 2"
                echo
                print_info "On Node 2, run: ./start-node.sh and select option 6, then Node 2"
                echo
                read -p "Press Enter when ready..."

                ./test-integration.sh node2
                ;;
            *)
                print_error "Invalid choice"
                exit 1
                ;;
        esac
        ;;

    0)
        echo "Goodbye!"
        exit 0
        ;;

    *)
        print_error "Invalid choice"
        exit 1
        ;;
esac
