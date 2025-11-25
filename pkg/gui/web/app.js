// MeshRadio Web GUI
class MeshRadioGUI {
    constructor() {
        this.ws = null;
        this.mode = 'idle';
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;

        this.init();
    }

    init() {
        this.connectWebSocket();
        this.setupEventListeners();
        this.startAnimations();
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
            this.updateNetworkStatus(true);
            this.addLog('MeshRadio interface ready', 'success');
        };

        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.updateStatus(data);
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.updateNetworkStatus(false);
            this.attemptReconnect();
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.addLog('Connection error', 'error');
        };
    }

    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            this.addLog(`Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`, 'info');
            setTimeout(() => this.connectWebSocket(), 2000 * this.reconnectAttempts);
        } else {
            this.addLog('Connection lost. Please refresh the page.', 'error');
        }
    }

    setupEventListeners() {
        // Broadcast button
        document.getElementById('broadcast-btn').addEventListener('click', () => {
            this.toggleBroadcast();
        });

        // Listen button
        document.getElementById('listen-btn').addEventListener('click', () => {
            this.toggleListen();
        });
    }

    async toggleBroadcast() {
        const btn = document.getElementById('broadcast-btn');

        if (this.mode === 'broadcasting') {
            // Stop broadcasting
            try {
                const response = await fetch('/api/broadcast/stop', { method: 'POST' });
                const data = await response.json();
                this.addLog('Broadcasting stopped', 'info');
                btn.textContent = 'Start Broadcasting';
                btn.classList.remove('active');
                document.getElementById('broadcast-info').style.display = 'none';
            } catch (error) {
                this.addLog('Error stopping broadcast: ' + error.message, 'error');
            }
        } else {
            // Start broadcasting
            try {
                const response = await fetch('/api/broadcast/start', { method: 'POST' });
                const data = await response.json();

                if (data.error) {
                    this.addLog('Error: ' + data.error, 'error');
                } else {
                    this.addLog('Broadcasting started', 'success');
                    btn.textContent = 'Stop Broadcasting';
                    btn.classList.add('active');
                    document.getElementById('broadcast-info').style.display = 'block';
                }
            } catch (error) {
                this.addLog('Error starting broadcast: ' + error.message, 'error');
            }
        }
    }

    async toggleListen() {
        const btn = document.getElementById('listen-btn');

        if (this.mode === 'listening') {
            // Stop listening
            try {
                const response = await fetch('/api/listen/stop', { method: 'POST' });
                const data = await response.json();
                this.addLog('Listening stopped', 'info');
                btn.textContent = 'Start Listening';
                btn.classList.remove('active');
                document.getElementById('listen-info').style.display = 'none';
                document.getElementById('listen-input').style.display = 'block';
            } catch (error) {
                this.addLog('Error stopping listen: ' + error.message, 'error');
            }
        } else {
            // Start listening
            const ipv6Input = document.getElementById('target-ipv6');
            const ipv6 = ipv6Input.value.trim();

            if (!ipv6) {
                this.addLog('Please enter an IPv6 address', 'error');
                return;
            }

            try {
                const response = await fetch('/api/listen/start', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ ipv6 })
                });
                const data = await response.json();

                if (data.error) {
                    this.addLog('Error: ' + data.error, 'error');
                } else {
                    this.addLog(`Listening to ${ipv6}`, 'success');
                    btn.textContent = 'Stop Listening';
                    btn.classList.add('active');
                    document.getElementById('listen-info').style.display = 'block';
                    document.getElementById('listen-input').style.display = 'none';
                }
            } catch (error) {
                this.addLog('Error starting listen: ' + error.message, 'error');
            }
        }
    }

    updateStatus(status) {
        // Update callsign and IPv6
        document.getElementById('callsign').textContent = status.callsign || '-';
        document.getElementById('ipv6').textContent = status.ipv6 || '-';

        // Update mode
        const modeBadge = document.getElementById('mode');
        modeBadge.textContent = status.mode.charAt(0).toUpperCase() + status.mode.slice(1);
        modeBadge.className = 'value status-badge';

        if (status.mode === 'broadcasting') {
            modeBadge.classList.add('broadcasting');
            document.getElementById('broadcast-addr').textContent = status.ipv6 + ':9001';
        } else if (status.mode === 'listening') {
            modeBadge.classList.add('listening');
            document.getElementById('station-name').textContent = status.station || 'Unknown';
            document.getElementById('packet-count').textContent = status.packetCount || 0;

            // Update signal strength
            const signalPercent = Math.min((status.packetCount % 100), 100);
            document.getElementById('signal-strength').style.width = signalPercent + '%';
        }

        this.mode = status.mode;
    }

    updateNetworkStatus(connected) {
        const indicator = document.getElementById('network-status');
        const text = document.getElementById('network-text');

        if (connected) {
            indicator.style.color = 'var(--success)';
            text.textContent = 'Connected';
        } else {
            indicator.style.color = 'var(--danger)';
            text.textContent = 'Disconnected';
        }
    }

    addLog(message, type = 'info') {
        const log = document.getElementById('activity-log');
        const entry = document.createElement('div');
        entry.className = `log-entry ${type}`;

        const timestamp = new Date().toLocaleTimeString();
        entry.textContent = `[${timestamp}] ${message}`;

        log.insertBefore(entry, log.firstChild);

        // Keep only last 20 entries
        while (log.children.length > 20) {
            log.removeChild(log.lastChild);
        }
    }

    startAnimations() {
        // Animate audio level meter
        setInterval(() => {
            if (this.mode === 'broadcasting') {
                const audioLevel = document.getElementById('audio-level');
                const randomLevel = 40 + Math.random() * 40; // 40-80%
                audioLevel.style.width = randomLevel + '%';
            }
        }, 200);
    }
}

// Initialize GUI when page loads
document.addEventListener('DOMContentLoaded', () => {
    window.meshRadio = new MeshRadioGUI();
});
