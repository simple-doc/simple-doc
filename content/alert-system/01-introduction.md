---
title: Introduction
---

# Alert System Introduction

The SolarFlux Alert System enables automated notifications when space weather conditions meet your specified criteria. Alerts help operators of satellites, power grids, aviation systems, and communication networks respond quickly to potentially disruptive solar events.

![Alert System Use Case](static/images/alert-use-case.svg)

## Why Alerts Matter

Space weather events can escalate rapidly. A solar flare reaches peak intensity within minutes, and a coronal mass ejection can arrive at Earth in as little as 15 hours. The Alert System delivers notifications within seconds of condition changes, giving your team maximum lead time to activate protective measures.

## Key Features

- **Threshold-based triggers** — Set alerts on any measurable parameter (Kp index, proton flux, X-ray flux, solar wind speed)
- **Event-based triggers** — Get notified when specific event types occur (X-class flares, Earth-directed CMEs, radiation storms)
- **Multiple delivery channels** — Webhook, email, SMS, Slack, and PagerDuty integrations
- **Escalation rules** — Define tiered responses: informational at lower thresholds, urgent pages at critical levels
- **Quiet hours** — Suppress non-critical alerts during specified time windows
- **Alert grouping** — Avoid notification floods by grouping related events within a configurable time window

## How It Works

1. **Create an alert rule** — Define the condition, threshold, and delivery channel
2. **SolarFlux monitors** — Our pipeline ingests data from GOES, DSCOVR, ACE, and SDO satellites in near-real-time
3. **Condition matched** — When incoming data satisfies your rule, an alert is triggered
4. **Notification delivered** — The alert payload is sent to your configured endpoints
5. **Acknowledgment** — Optionally require manual acknowledgment to track response times

## Quick Start

Create your first alert with a single API call:

```bash
curl -X POST https://api.solarflux.dev/v2/alerts/rules \
  -H "X-API-Key: sf_live_abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "High Kp Alert",
    "condition": {
      "metric": "kp_index",
      "operator": "gte",
      "value": 7
    },
    "channel": {
      "type": "webhook",
      "url": "https://ops.example.com/hooks/space-weather"
    },
    "severity": "critical"
  }'
```
