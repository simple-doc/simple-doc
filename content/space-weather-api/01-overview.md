---
title: Overview
---

# Space Weather API Overview

![SolarFlux Platform Architecture](static/images/architecture.svg)

The SolarFlux Space Weather API provides programmatic access to real-time and historical space weather data collected from ground-based observatories and satellite instruments worldwide.

## What is Space Weather?

Space weather refers to conditions on the Sun and in the solar wind, magnetosphere, ionosphere, and thermosphere that can affect the performance of space-borne and ground-based technological systems. Major space weather events include:

- **Solar flares** — sudden bursts of electromagnetic radiation from the Sun's surface
- **Coronal mass ejections (CMEs)** — large expulsions of plasma and magnetic field from the solar corona
- **Geomagnetic storms** — disturbances in Earth's magnetosphere caused by solar wind
- **Solar energetic particles (SEPs)** — high-energy particles accelerated by solar events

## API Capabilities

The Space Weather API allows you to:

| Capability | Description |
|---|---|
| Query solar events | Retrieve solar flare, CME, and SEP event data with filtering by date, intensity, and region |
| Monitor conditions | Get current Kp index, solar wind speed, interplanetary magnetic field, and proton flux readings |
| Forecast access | Retrieve 1-hour, 3-day, and 27-day space weather forecasts |
| Historical data | Access archived observations dating back to Solar Cycle 19 (1954) |
| Aurora forecasts | Get hemispheric aurora probability maps updated every 30 minutes |

## Base URL

All API requests are made to:

```
https://api.solarflux.dev/v2
```

A sandbox environment is available for testing:

```
https://sandbox.api.solarflux.dev/v2
```

The sandbox returns synthetic data and does not count against your rate limits.

## Response Format

All responses are returned in JSON. Timestamps use ISO 8601 format in UTC. Numeric measurements include their unit in a companion field.

```json
{
  "event_id": "FLR-2025-04-12T09:22Z",
  "type": "solar_flare",
  "class": "X2.1",
  "peak_time": "2025-04-12T09:22:00Z",
  "source_region": "AR3842",
  "location": {
    "latitude": -14.2,
    "longitude": 32.8
  },
  "duration_minutes": 47,
  "associated_cme": true
}
```

## Versioning

The API uses URL-based versioning. The current stable version is `v2`. The previous version `v1` remains available but is deprecated and will be retired on 2026-06-01.
