# MeshRadio Web UI — Text Illustration

Rough textual sketch of the intended web UI (desktop-first, responsive down to mobile).

```
+----------------------------------------------------------------------------------+
| MESHRADIO [CALLSIGN]                         Status: ● Live/Idle   IPv6: xxxx::1 |
+----------------------------------------------------------------------------------+
| MeshRadio on Yggdrasil                                      IPv6 as frequency    |
| [ Broadcast ]   [ Listen ]   [ Scan (soon) ]                                     |
+-----------------------------------+----------------------------------------------+
| Broadcast                         | Listen                                       |
|-----------------------------------+----------------------------------------------|
| Mic: [ Default ▼ ]                | Tune IPv6: [ xxxx:xxxx::1    ] [ Tune ]      |
| [ Start / Stop ]  (● Live 00:12)  | Last: [ xxxx:beef::cafe ]                    |
| Audio Level: |||█||█|             | Signal: ▂▄▆█                                 |
| IPv6 is your frequency            | Packets/sec: 42                              |
+-----------------------------------+----------------------------------------------+
| Nearby Stations (beacons)         | Activity Log                                 |
|-----------------------------------+----------------------------------------------|
| [CALL] xxxx:...:1234  [Tune]      | 12:01 beacon recv from xxxx::1               |
| [NODE] yyyy:...:abcd  [Tune]      | 12:02 tuned yyyy::abcd                       |
| (empty state: "No beacons yet")   | 12:03 warn: audio simulated                  |
+-----------------------------------+----------------------------------------------+
| Ports: GUI 7999 · Broadcaster 8799 · Listener 9799 ("799" theme)                 |
| Mesh-first; Yggdrasil required; audio currently simulated                        |
+----------------------------------------------------------------------------------+
```

Mobile: stack cards vertically; hero compresses; log collapses into a tab; buttons become full-width.
