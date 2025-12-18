# Alertmanager Discord Webhook

![Go Report Card](https://goreportcard.com/badge/github.com/kolesaev/alertmanager-discord?style=flat-square)

## Goal

The goal of this application is to serve as a customizable Discord webhook for Alertmanager with enhanced features for modern monitoring workflows.

Based on the original [alertmanager-discord](https://github.com/masgustavos/alertmanager-discord) by masgustavos, this fork adds powerful new capabilities while maintaining compatibility with existing configurations.

## Enhanced Features

### 🎯 Smart Alert Processing
- **Dashboard Links**: Automatically add Grafana dashboard links when URL is provided in alert labels
- **Prometheus Query Links**: Include direct links to Prometheus/VictoriaMetrics queries for quick investigation
- **Time Information Display**: Show alert start/end times and duration with configurable labels
- **Intelligent Title Extraction**: Follows Telegram template logic for consistent alert titles

### 🎨 Flexible Configuration
- **Multiple Message Types**: Choose between `status` or `severity` based message formatting
- **Configurable Severity Levels**: Customize colors, emojis, and priorities for different severity levels
- **Mention Control**: Configure which roles to mention based on severity or alert count thresholds
- **Channel-specific Overrides**: Different configurations for different Discord channels

### 🔗 External Integration
- **Dashboard Link Configuration**: Customize dashboard URL labels and link text
- **Generator URL Support**: Add links to Prometheus query pages automatically
- **Time Display Options**: Control how and when to show alert timing information

## How It Works

### Alert Flow
1. **Prometheus** evaluates alert rules and sends firing alerts to **Alertmanager**
2. **Alertmanager** routes alerts based on label matching to the appropriate webhook receiver
3. **This application** processes the alerts and sends formatted messages to **Discord**

### Routing Example

```yaml
group_by:
  - severity
  - alertname
  - summary
  - url

routes:
  - match:
      severity: critical
    receiver: "discord-critical"
    continue: true
  - match:
      team: platform
    receiver: "discord-platform"
    continue: false
  - receiver: "discord-default"

receivers:
  - name: "discord-default"
    webhook_configs:
      - send_resolved: true
        url: "http://app:8080/default"
  - name: "discord-critical"
    webhook_configs:
      - send_resolved: true
        url: "http://app:8080/critical"
  - name: "discord-platform"
    webhook_configs:
      - send_resolved: true
        url: "http://app:8080/platform"
```

## New Configuration Options

### Dashboard Links
```yaml
dashboardLink:
  enabled: true                    # Enable dashboard links
  label: "url"                     # Label containing dashboard URL (commonly "url" or "grafana_url")
  text: "Open in Dashboard"        # Link text to display
```

### Generator Links
```yaml
generatorLink:
  enabled: true                    # Enable generator links
  text: "Open in PromQL"           # Link text for Prometheus queries
```

### Time Display
```yaml
timeDisplay:
  enabled: true                    # Show time information in alerts
  startsAtText: "Started at:"      # Label for alert start time
  endsAtText: "Ended at:"          # Label for alert end time (resolved alerts only)
  durationText: "Duration:"        # Label for alert duration (resolved alerts only)
```

## Example Message Output

### Message Header
```
[Open in Dashboard](https://grafana.example.com/d/abc123)
[Open in PromQL](https://prometheus.example.com/graph?g0.expr=...)
```

### Alert Embed (Firing)
**Title:** `⚠️ High CPU usage on server-web-01`

**Description:**
```
🔔
CPU usage is above 80% for more than 5 minutes

🕑
Started at: 18.12.2025 14:30:15 UTC
```

### Alert Embed (Resolved)
**Title:** `✅ High CPU usage on server-web-01`

**Description:**
```
🔔
CPU usage is above 80% for more than 5 minutes

🕑
Started at: 18.12.2025 14:30:15 UTC
Ended at: 18.12.2025 14:45:30 UTC
Duration: 15m 15s
```

## Message Types

### Status-Based Messages
When `messageType` is set to `status` (default), alerts are grouped by name and styled according to their status (firing/resolved).

### Severity-Based Messages
When `messageType` is set to `severity`, additional features are enabled:
- Customizable severity label name (default: `severity`)
- Different colors and emojis for each severity level
- Priority-based ordering (higher priority alerts shown first)
- Configurable mentions based on severity
- Option to ignore certain severities when alone

## Installation

### Docker Compose
1. Copy `config.example.yaml` to `config.yaml`
2. Update with your Discord webhook URLs and configuration
3. Run with `docker-compose up -d`

### Kubernetes with Helm
A Helm chart is available for easy Kubernetes deployment:

```bash
helm repo add masgustavos https://masgustavos.github.io/helm
helm install alertmanager-discord --values values.yaml masgustavos/alertmanager-discord
```

## Configuration File

The application supports both YAML and JSON configuration files. See `config.example.yaml` for a fully commented example with all available options.

### Key Configuration Sections
- **`severity`**: Define severity levels, colors, emojis, and priorities
- **`status`**: Configure appearance for firing and resolved alerts
- **`channels`**: Discord channel-specific configurations and webhook URLs
- **`dashboardLink`/`generatorLink`**: Configure external link generation
- **`timeDisplay`**: Control time information display in alerts

## Development and Experimentation

Use the provided docker-compose setup to experiment with different configurations:
- `mock/prometheus/`: Sample Prometheus configuration with alert rules
- `mock/alertmanager/`: Alertmanager configuration with routing rules
- `mock/helm/`: Helm chart values for Kubernetes deployment

## Discord Colors Reference

When configuring severity colors, use these Discord color codes:

```yaml
EmbedColorAqua: 1752220
EmbedColorGreen: 3066993
EmbedColorBlue: 3447003
EmbedColorPurple: 10181046
EmbedColorGold: 15844367
EmbedColorOrange: 15105570
EmbedColorRed: 15158332
EmbedColorGrey: 9807270
EmbedColorDarkerGrey: 8359053
EmbedColorNavy: 3426654
EmbedColorDarkAqua: 1146986
EmbedColorDarkGreen: 2067276
EmbedColorDarkBlue: 2123412
EmbedColorDarkPurple: 7419530
EmbedColorDarkGold: 12745742
EmbedColorDarkOrange: 11027200
EmbedColorDarkRed: 10038562
EmbedColorDarkGrey: 9936031
EmbedColorLightGrey: 12370112
EmbedColorDarkNavy: 2899536
EmbedColorLuminousVividPink: 16580705
EmbedColorDarkVividPink: 12320855
```

## Compatibility

This fork maintains backward compatibility with existing configurations from the original project. All original features work as expected, with new features being opt-in through configuration.

## Credits

This project is based on [alertmanager-discord](https://github.com/masgustavos/alertmanager-discord) by masgustavos, extended with additional features for modern monitoring workflows.