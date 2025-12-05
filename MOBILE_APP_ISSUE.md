# Mobile App, Anyone?

We're looking for iOS/Android developers interested in building a **MeshRadio mobile app**! 📱

## Why Mobile?

MeshRadio works great on desktop, but imagine:
- Listening to mesh radio stations on your phone while walking
- Portable emergency broadcasting device
- Mesh radio in your pocket, anywhere with Yggdrasil connectivity

The live test station (see #1) proves the core streaming works reliably. Now we need mobile clients!

## iOS Approach (Simplest Path)

Thanks to `gomobile`, we can reuse MeshRadio's Go code on iOS:

### Architecture
```
┌─────────────────────────────────────────┐
│         Swift UI (iOS App)              │
│  ┌────────────┐  ┌──────────────────┐   │
│  │ Station    │  │  Audio Player    │   │
│  │ Selector   │  │  (AVFoundation)  │   │
│  └─────┬──────┘  └────────┬─────────┘   │
├────────┴──────────────────┴──────────────┤
│      MeshRadio.framework (Go)            │
│  • Listener (receive UDP packets)       │
│  • Opus Decoder                          │
│  • RTP Protocol Handler                 │
├──────────────────────────────────────────┤
│    Yggdrasil.framework                   │
│  • Network Extension (VPN)               │
│  • IPv6 Mesh Routing                     │
└──────────────────────────────────────────┘
```

### Implementation Plan

**Phase 1: Minimal Viable App (Weekend Project)**
- Extract listener logic to `pkg/mobile/` with mobile-friendly API
- Build iOS framework: `gomobile bind -target=ios -o MeshRadio.framework ./pkg/mobile`
- Create simple Swift UI with:
  - TextField for station IPv6
  - Listen/Stop button
  - Audio playback via AVAudioEngine
- Test on simulator with localhost broadcaster (no Yggdrasil needed initially)

**Phase 2: Yggdrasil Integration**
- Integrate [yggdrasil-ios](https://github.com/yggdrasil-network/yggdrasil-ios)
- Network Extension for mesh connectivity
- Real mesh radio on mobile!

**Phase 3: Polish**
- Station discovery/favorites
- Background audio playback
- Now playing metadata
- App Store release

### Code Sketch

**Go Mobile Bindings** (`pkg/mobile/listener.go`):
```go
package mobile

import (
    "github.com/meshradio/meshradio/internal/listener"
    "github.com/meshradio/meshradio/pkg/audio"
)

type MobileListener struct {
    listener  *listener.Listener
    audioChan chan []byte
}

func NewListener(ipv6 string, port int) (*MobileListener, error) {
    // Create listener that outputs PCM to channel
}

func (m *MobileListener) Start() error { }
func (m *MobileListener) ReadAudio() []byte { }  // iOS reads PCM samples
func (m *MobileListener) Stop() { }
```

**Swift UI** (ContentView.swift):
```swift
import SwiftUI
import AVFoundation
import MeshRadio

struct ContentView: View {
    @State private var stationIPv6 = "206:6da3:9f2:60d2:9769:21c6:10fe:11d4"
    @State private var isListening = false
    @StateObject private var audioPlayer = AudioPlayer()

    var body: some View {
        VStack(spacing: 20) {
            Text("📻 MeshRadio")
                .font(.largeTitle)

            TextField("Station IPv6", text: $stationIPv6)
                .textFieldStyle(RoundedBorderTextFieldStyle())
                .padding()

            Button(action: toggleListening) {
                Text(isListening ? "Stop" : "Listen")
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(isListening ? Color.red : Color.blue)
                    .foregroundColor(.white)
                    .cornerRadius(10)
            }
        }
    }

    func toggleListening() {
        isListening.toggle()
        if isListening {
            audioPlayer.start(ipv6: stationIPv6, port: 8799)
        } else {
            audioPlayer.stop()
        }
    }
}
```

**Build Commands**:
```bash
# Install gomobile (requires macOS + Xcode)
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Build MeshRadio framework for iOS
gomobile bind -target=ios -o MeshRadio.framework ./pkg/mobile

# Open in Xcode, add framework, build and run!
```

## Android Approach

Android is even simpler with `gomobile`:

```bash
# Build Android library
gomobile bind -target=android -o meshradio.aar ./pkg/mobile
```

Kotlin UI can use the same Go bindings. AudioTrack handles PCM playback.

## What We Need

**iOS Developer:**
- Familiar with Swift/SwiftUI
- Experience with AVFoundation for audio
- Access to macOS + Xcode
- (Optional) Apple Developer account for device testing

**Android Developer:**
- Familiar with Kotlin/Jetpack Compose
- Experience with AudioTrack
- Android Studio

**Go Developer:**
- Help design mobile-friendly API in `pkg/mobile/`
- Handle Go ↔ mobile memory management
- Optimize for battery life

## Why This is Feasible

- MeshRadio's Go code is already modular and testable
- `gomobile` handles the hard part (language bindings)
- Listener-only app is simpler (no microphone permissions)
- Core streaming proven stable with 24/7 test station
- Yggdrasil already has iOS app as reference
- Estimated ~200 lines of Swift + Go bindings for MVP

## Getting Started

Interested? Let's discuss:
1. Which platform do you prefer (iOS/Android)?
2. What's your experience level with mobile + Go?
3. Want to pair program on this?

**Resources:**
- Yggdrasil iOS: https://github.com/yggdrasil-network/yggdrasil-ios
- gomobile docs: https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile
- Live test station: #1

Let's bring mesh radio to mobile! 📻📱
