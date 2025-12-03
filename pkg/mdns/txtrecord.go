package mdns

import (
	"fmt"
	"strconv"
	"strings"
)

// CreateTXTRecord creates TXT record strings from ServiceInfo
func CreateTXTRecord(info ServiceInfo) []string {
	txt := []string{
		fmt.Sprintf("group=%s", info.Group),
		fmt.Sprintf("channel=%s", info.Channel),
		fmt.Sprintf("callsign=%s", info.Callsign),
		fmt.Sprintf("priority=%s", info.Priority),
		fmt.Sprintf("codec=%s", info.Codec),
		fmt.Sprintf("bitrate=%d", info.Bitrate),
	}
	return txt
}

// ParseTXTRecord parses TXT record strings into ServiceInfo
func ParseTXTRecord(txt []string) (ServiceInfo, error) {
	info := ServiceInfo{
		Codec:   "opus", // Default codec
		Bitrate: 64,     // Default bitrate
	}

	for _, record := range txt {
		parts := strings.SplitN(record, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "group":
			info.Group = value
		case "channel":
			info.Channel = value
		case "callsign":
			info.Callsign = value
		case "priority":
			info.Priority = value
		case "codec":
			info.Codec = value
		case "bitrate":
			if bitrate, err := strconv.Atoi(value); err == nil {
				info.Bitrate = bitrate
			}
		}
	}

	return info, nil
}
