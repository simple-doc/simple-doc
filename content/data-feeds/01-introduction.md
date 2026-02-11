---
title: Introduction
---

# Data Feeds Introduction

![Data Feed Pipeline](static/images/data-flow.svg)

SolarFlux Data Feeds provide continuous streams of space weather measurements and processed datasets. Use feeds for applications that need low-latency data ingestion, such as real-time dashboards, automated trading systems for energy markets, or satellite anomaly correlation.

## Feed Types

| Feed | Cadence | Description |
|---|---|---|
| **Solar Wind** | 1-minute | ACE/DSCOVR L1 solar wind plasma and magnetic field |
| **X-Ray Flux** | 1-minute | GOES X-ray irradiance (short and long channels) |
| **Proton Flux** | 5-minute | GOES energetic particle sensor readings |
| **Kp Index** | 3-hour | Estimated and definitive planetary K-index |
| **Dst Index** | 1-hour | Real-time disturbance storm time index |
| **Aurora Probability** | 30-minute | Hemispheric aurora oval maps with probabilities |
| **Sunspot Number** | Daily | International sunspot number and solar flux index |

## Connection Methods

Data feeds are available via two transport mechanisms:

### WebSocket (Recommended)

Low-latency, bidirectional connection ideal for real-time applications.

```javascript
const ws = new WebSocket("wss://feeds.solarflux.dev/v2/stream");

ws.onopen = () => {
  ws.send(JSON.stringify({
    action: "subscribe",
    feeds: ["solar_wind", "xray_flux"],
    api_key: "sf_live_abc123"
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(`[${data.feed}] ${data.timestamp}:`, data.values);
};
```

### Server-Sent Events (SSE)

Simpler unidirectional stream, compatible with any HTTP client.

```bash
curl -N -H "X-API-Key: sf_live_abc123" \
  "https://feeds.solarflux.dev/v2/sse?feeds=solar_wind,kp_index"
```

## Data Retention

| Plan | Real-time feeds | Historical bulk data |
|---|---|---|
| Free | Last 24 hours | Last 30 days |
| Professional | Last 7 days replay | Last 5 years |
| Enterprise | Last 30 days replay | Full archive (1954-present) |
