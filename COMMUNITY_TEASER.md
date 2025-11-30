# MeshRadio × Yggdrasil — RFC (MeshRadio: The Last of Us)

Day-one “what if?” RFC for Yggdrasil spaces (Matrix/issue tracker). Intent: feel out feasibility/ideas/collaborators before writing code or asking for testers.

## Draft blurb (conversational)
```
What if we built a mesh radio/ham app on Yggdrasil? Picture a world where Cloudflare/AWS are dark (we’ve seen outages), but the network still exists. Instead of central services, we lean on mesh.

Idea: audio (maybe video later) over Yggdrasil. The “frequency” isn’t RF, it’s IPv6. You broadcast; others tune to your IPv6. Peers help each other navigate and communicate when the big clouds are down.

Reality: no repo yet. We’d start with fake audio and wire real PortAudio/Opus after feedback.

Ask: does this seem worthwhile/feasible to Ygg folks? Any warnings about admin socket/peer discovery, or yggdrasil-go issues to read first? If you want to pair on the networking/audio bits, please say hi. https://github.com/yggdrasil-network/yggdrasil-go/issues
```

## Notes for posters
- Tone: humble “what if?” RFC; explicitly ask for feasibility comments/concerns and invite collaborators before any tester call.
- Audience: Yggdrasil core contributors, Matrix users, and issue-tracker readers.
- Goal: gather comments on discovery/admin APIs, UX, and what to emphasize or avoid in the final teaser.
- Follow-ups to collect: working node pairs, platform/CPU data, routing quirks, and pointers to issues aligned with media/real-time use-cases.

## Next (if green-lit)
- Create a minimal repo with fake audio
- Wire PortAudio/Opus after early feedback
- Share early builds for folks who opt in
