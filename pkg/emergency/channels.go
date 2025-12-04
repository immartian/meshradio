package emergency

// Channel represents an emergency channel definition
type Channel struct {
	Name        string       // Channel name (e.g., "emergency")
	Group       string       // Multicast group
	Port        int          // RTP port
	Priority    Priority     // Default priority
	AutoTune    AutoTuneMode // Default auto-tune behavior
	Description string       // Human-readable description
}

// Standard emergency port numbers (8790-8799)
const (
	PortEmergency   = 8790
	PortNetControl  = 8791
	PortMedical     = 8792
	PortWeather     = 8793
	PortSAR         = 8794
	PortCommunity   = 8795
	PortReserved1   = 8796
	PortReserved2   = 8797
	PortTalk        = 8798
	PortTest        = 8799
)

// StandardChannels defines the standard emergency channels
var StandardChannels = map[string]Channel{
	"emergency": {
		Name:        "emergency",
		Group:       "emergency",
		Port:        PortEmergency,
		Priority:    PriorityCritical,
		AutoTune:    AutoTuneAlways,
		Description: "General emergency broadcast - active emergency in progress",
	},
	"netcontrol": {
		Name:        "netcontrol",
		Group:       "netcontrol",
		Port:        PortNetControl,
		Priority:    PriorityEmergency,
		AutoTune:    AutoTunePrompt,
		Description: "Emergency net control - coordination and resource management",
	},
	"medical": {
		Name:        "medical",
		Group:       "medical",
		Port:        PortMedical,
		Priority:    PriorityEmergency,
		AutoTune:    AutoTunePrompt,
		Description: "Medical emergency coordination - health and safety",
	},
	"weather": {
		Name:        "weather",
		Group:       "weather",
		Port:        PortWeather,
		Priority:    PriorityHigh,
		AutoTune:    AutoTuneNever,
		Description: "Weather alerts and warnings - severe weather notifications",
	},
	"sar": {
		Name:        "sar",
		Group:       "sar",
		Port:        PortSAR,
		Priority:    PriorityEmergency,
		AutoTune:    AutoTunePrompt,
		Description: "Search and rescue - missing persons and rescue operations",
	},
	"community": {
		Name:        "community",
		Group:       "community",
		Port:        PortCommunity,
		Priority:    PriorityNormal,
		AutoTune:    AutoTuneNever,
		Description: "Community service - public announcements and community info",
	},
	"talk": {
		Name:        "talk",
		Group:       "talk",
		Port:        PortTalk,
		Priority:    PriorityNormal,
		AutoTune:    AutoTuneNever,
		Description: "General conversation - casual communication",
	},
	"test": {
		Name:        "test",
		Group:       "test",
		Port:        PortTest,
		Priority:    PriorityNormal,
		AutoTune:    AutoTuneNever,
		Description: "Testing - system testing and development",
	},
}

// ChannelRegistry manages emergency channels
type ChannelRegistry struct {
	channels map[string]Channel
}

// NewChannelRegistry creates a new channel registry
func NewChannelRegistry() *ChannelRegistry {
	// Copy standard channels
	channels := make(map[string]Channel)
	for k, v := range StandardChannels {
		channels[k] = v
	}

	return &ChannelRegistry{
		channels: channels,
	}
}

// Get returns a channel by name
func (r *ChannelRegistry) Get(name string) (Channel, bool) {
	ch, ok := r.channels[name]
	return ch, ok
}

// GetByPort returns a channel by port number
func (r *ChannelRegistry) GetByPort(port int) (Channel, bool) {
	for _, ch := range r.channels {
		if ch.Port == port {
			return ch, true
		}
	}
	return Channel{}, false
}

// GetByGroup returns a channel by group name
func (r *ChannelRegistry) GetByGroup(group string) (Channel, bool) {
	for _, ch := range r.channels {
		if ch.Group == group {
			return ch, true
		}
	}
	return Channel{}, false
}

// List returns all channels
func (r *ChannelRegistry) List() []Channel {
	channels := make([]Channel, 0, len(r.channels))
	for _, ch := range r.channels {
		channels = append(channels, ch)
	}
	return channels
}

// ListEmergency returns only emergency channels (priority >= Emergency)
func (r *ChannelRegistry) ListEmergency() []Channel {
	channels := make([]Channel, 0)
	for _, ch := range r.channels {
		if ch.Priority >= PriorityEmergency {
			channels = append(channels, ch)
		}
	}
	return channels
}

// Add adds a custom channel
func (r *ChannelRegistry) Add(ch Channel) {
	r.channels[ch.Name] = ch
}

// Remove removes a channel
func (r *ChannelRegistry) Remove(name string) {
	delete(r.channels, name)
}

// IsEmergencyChannel checks if a channel name is an emergency channel
func IsEmergencyChannel(name string) bool {
	ch, ok := StandardChannels[name]
	if !ok {
		return false
	}
	return ch.Priority >= PriorityEmergency
}

// IsEmergencyPort checks if a port is an emergency port
func IsEmergencyPort(port int) bool {
	return port >= PortEmergency && port <= PortSAR
}
