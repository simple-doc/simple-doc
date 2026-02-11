---
title: Alert Types
---

# Alert Types

SolarFlux categorizes space weather alerts into five domains, each corresponding to a class of operational impact.

## Solar Radiation Alerts

Triggered when energetic particle flux exceeds defined thresholds. These alerts are critical for aviation (polar routes), crewed spaceflight, and satellite operations.

**Levels:**

| Level | Threshold (>10 MeV pfu) | Typical Impact |
|---|---|---|
| S1 — Minor | 10 | Minimal impact on HF radio |
| S2 — Moderate | 100 | Infrequent single-event upsets in satellites |
| S3 — Strong | 1,000 | Elevated radiation risk on polar flights |
| S4 — Severe | 10,000 | Blackout of HF radio in polar regions |
| S5 — Extreme | 100,000 | Satellite damage possible, complete HF blackout |

## Geomagnetic Storm Alerts

Triggered by sustained disturbances in Earth's magnetic field, typically caused by CME impacts or high-speed solar wind streams.

**Levels:**

| Level | Kp Index | Typical Impact |
|---|---|---|
| G1 — Minor | 5 | Weak power grid fluctuations, aurora visible at high latitudes |
| G2 — Moderate | 6 | Voltage alarms in high-latitude power systems |
| G3 — Strong | 7 | Intermittent satellite navigation problems |
| G4 — Severe | 8 | Widespread voltage control problems, GPS degraded |
| G5 — Extreme | 9 | Possible widespread grid collapse, GPS unusable |

## Radio Blackout Alerts

Triggered by solar X-ray flares that ionize Earth's dayside ionosphere, disrupting HF radio propagation.

**Levels:**

| Level | Flare Class | Typical Impact |
|---|---|---|
| R1 — Minor | M1 | Weak degradation of HF signals on sunlit side |
| R2 — Moderate | M5 | Limited HF blackout on sunlit side |
| R3 — Strong | X1 | Wide-area HF blackout for ~1 hour |
| R4 — Severe | X10 | HF blackout on entire sunlit side for ~2 hours |
| R5 — Extreme | X20+ | Complete HF blackout for several hours |

## CME Arrival Alerts

Triggered when a coronal mass ejection is detected heading toward Earth. Includes estimated arrival time and expected impact severity.

**Alert payload includes:**

```json
{
  "alert_type": "cme_arrival",
  "cme_id": "CME-2025-04-12T10:48Z",
  "estimated_arrival": "2025-04-14T06:00:00Z",
  "confidence": 0.82,
  "estimated_speed_km_s": 1200,
  "expected_kp_range": [6, 8],
  "lead_time_hours": 43.2
}
```

## Aurora Alerts

Triggered when aurora activity is predicted to be visible at specified geographic latitudes. Useful for aurora tourism operators and photography enthusiasts.

**Configuration:**

```json
{
  "name": "Aurora visible in Portland",
  "condition": {
    "metric": "aurora_visibility_latitude",
    "operator": "lte",
    "value": 45.5
  },
  "channel": {
    "type": "email",
    "address": "alerts@example.com"
  },
  "severity": "info"
}
```
