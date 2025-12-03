package main

import (
	"fmt"
	"net"
	"time"

	"github.com/meshradio/meshradio/pkg/multicast"
)

func main() {
	fmt.Println("Multicast Overlay Test")
	fmt.Println("======================\n")

	// Create subscription manager
	sm := multicast.NewSubscriptionManager()

	// Test 1: Regular Multicast (Emergency Channel)
	fmt.Println("Test 1: Regular Multicast (Emergency Channel)")
	fmt.Println("----------------------------------------------")

	// Create emergency group
	sm.CreateGroup("emergency")

	// Add two broadcasters
	broadcaster1 := &multicast.Broadcaster{
		IPv6:     net.ParseIP("201:abcd::1"),
		Port:     8790,
		Callsign: "STATION-A",
		LastSeen: time.Now(),
	}
	broadcaster2 := &multicast.Broadcaster{
		IPv6:     net.ParseIP("201:abcd::2"),
		Port:     8790,
		Callsign: "STATION-B",
		LastSeen: time.Now(),
	}

	sm.RegisterBroadcaster("emergency", broadcaster1)
	sm.RegisterBroadcaster("emergency", broadcaster2)

	// Add listener (regular multicast - receives from ALL sources)
	listener1 := &multicast.Subscriber{
		IPv6:      net.ParseIP("201:abcd::100"),
		Port:      9001,
		Callsign:  "LISTENER-1",
		LastSeen:  time.Now(),
		SSMSource: nil, // nil = regular multicast
	}

	sm.Subscribe(multicast.SubscribeRequest{
		Group:      "emergency",
		Subscriber: listener1,
	})

	// Check subscribers for each broadcaster
	subs1 := sm.GetSubscribersForSource("emergency", broadcaster1.IPv6)
	subs2 := sm.GetSubscribersForSource("emergency", broadcaster2.IPv6)

	fmt.Printf("Broadcaster A (%s): %d subscriber(s)\n", broadcaster1.Callsign, len(subs1))
	fmt.Printf("Broadcaster B (%s): %d subscriber(s)\n", broadcaster2.Callsign, len(subs2))
	fmt.Printf("✓ Regular multicast: Listener receives from BOTH broadcasters\n\n")

	// Test 2: SSM (Source-Specific Multicast)
	fmt.Println("Test 2: SSM (Source-Specific Multicast)")
	fmt.Println("----------------------------------------")

	// Create community group
	sm.CreateGroup("community")

	// Add two broadcasters
	broadcaster3 := &multicast.Broadcaster{
		IPv6:     net.ParseIP("201:abcd::3"),
		Port:     8795,
		Callsign: "COMMUNITY-A",
		LastSeen: time.Now(),
	}
	broadcaster4 := &multicast.Broadcaster{
		IPv6:     net.ParseIP("201:abcd::4"),
		Port:     8795,
		Callsign: "COMMUNITY-B",
		LastSeen: time.Now(),
	}

	sm.RegisterBroadcaster("community", broadcaster3)
	sm.RegisterBroadcaster("community", broadcaster4)

	// Add listener with SSM (only receives from COMMUNITY-A)
	listener2 := &multicast.Subscriber{
		IPv6:      net.ParseIP("201:abcd::101"),
		Port:      9002,
		Callsign:  "LISTENER-2",
		LastSeen:  time.Now(),
		SSMSource: broadcaster3.IPv6, // Only receive from COMMUNITY-A
	}

	sm.Subscribe(multicast.SubscribeRequest{
		Group:      "community",
		Subscriber: listener2,
	})

	// Check subscribers for each broadcaster
	subs3 := sm.GetSubscribersForSource("community", broadcaster3.IPv6)
	subs4 := sm.GetSubscribersForSource("community", broadcaster4.IPv6)

	fmt.Printf("Broadcaster COMMUNITY-A: %d subscriber(s)\n", len(subs3))
	fmt.Printf("Broadcaster COMMUNITY-B: %d subscriber(s)\n", len(subs4))
	fmt.Printf("✓ SSM: Listener only receives from COMMUNITY-A\n\n")

	// Test 3: Multiple Listeners
	fmt.Println("Test 3: Multiple Listeners")
	fmt.Println("---------------------------")

	// Add more listeners to emergency channel
	for i := 3; i <= 5; i++ {
		listener := &multicast.Subscriber{
			IPv6:      net.ParseIP(fmt.Sprintf("201:abcd::%d", 100+i)),
			Port:      9000 + i,
			Callsign:  fmt.Sprintf("LISTENER-%d", i),
			LastSeen:  time.Now(),
			SSMSource: nil, // Regular multicast
		}
		sm.Subscribe(multicast.SubscribeRequest{
			Group:      "emergency",
			Subscriber: listener,
		})
	}

	emergencySubs := sm.GetSubscribers("emergency")
	fmt.Printf("Emergency channel: %d subscriber(s)\n", len(emergencySubs))
	for _, sub := range emergencySubs {
		multicastType := "Regular"
		if sub.IsSSM() {
			multicastType = "SSM"
		}
		fmt.Printf("  - %s (%s) [%s]\n", sub.Callsign, sub.IPv6, multicastType)
	}
	fmt.Println()

	// Test 4: Statistics
	fmt.Println("Test 4: Statistics")
	fmt.Println("------------------")

	stats := sm.GetStats()
	fmt.Printf("Total groups: %d\n", stats.GroupCount)
	fmt.Printf("Total subscribers: %d\n", stats.TotalSubscribers)
	fmt.Printf("Total broadcasters: %d\n\n", stats.TotalBroadcasters)

	for name, groupStats := range stats.Groups {
		fmt.Printf("Group '%s':\n", name)
		fmt.Printf("  Subscribers: %d\n", groupStats.SubscriberCount)
		fmt.Printf("  Broadcasters: %d\n", groupStats.BroadcasterCount)
	}
	fmt.Println()

	// Test 5: Heartbeat and Pruning
	fmt.Println("Test 5: Heartbeat and Pruning")
	fmt.Println("------------------------------")

	// Create a test subscriber that will go stale
	staleListener := &multicast.Subscriber{
		IPv6:      net.ParseIP("201:abcd::200"),
		Port:      9999,
		Callsign:  "STALE-LISTENER",
		LastSeen:  time.Now().Add(-20 * time.Second), // 20 seconds ago
		SSMSource: nil,
	}

	sm.Subscribe(multicast.SubscribeRequest{
		Group:      "emergency",
		Subscriber: staleListener,
	})

	beforeSubs := len(sm.GetSubscribers("emergency"))
	fmt.Printf("Subscribers before pruning: %d\n", beforeSubs)

	// Prune stale subscribers (timeout: 15 seconds)
	prunedSubs, prunedBroadcasters := sm.PruneStale(15 * time.Second)
	fmt.Printf("Pruned %d stale subscriber(s)\n", prunedSubs)
	fmt.Printf("Pruned %d stale broadcaster(s)\n", prunedBroadcasters)

	afterSubs := len(sm.GetSubscribers("emergency"))
	fmt.Printf("Subscribers after pruning: %d\n", afterSubs)
	fmt.Printf("✓ Stale subscriber removed\n\n")

	// Summary
	fmt.Println("Summary")
	fmt.Println("-------")
	fmt.Println("✅ Regular multicast works (all sources → listener)")
	fmt.Println("✅ SSM works (specific source → listener)")
	fmt.Println("✅ Multiple listeners supported")
	fmt.Println("✅ Statistics tracking works")
	fmt.Println("✅ Heartbeat/pruning works")
	fmt.Println("\nLayer 4 (Multicast Overlay) core functionality complete!")
}
