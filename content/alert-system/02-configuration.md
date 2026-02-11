---
title: Configuration
---

# Alert Configuration

Alert rules define what conditions trigger a notification and where that notification is delivered.

## Alert Rule Structure

```json
{
  "id": "rule_8f2a3b4c",
  "name": "X-Class Flare Alert",
  "enabled": true,
  "condition": {
    "metric": "solar_flare_class",
    "operator": "gte",
    "value": "X1.0"
  },
  "channel": {
    "type": "webhook",
    "url": "https://ops.example.com/hooks/flares",
    "headers": {
      "X-Custom-Header": "space-weather"
    }
  },
  "severity": "critical",
  "cooldown_minutes": 30,
  "tags": ["operations", "satellite-team"],
  "created_at": "2025-03-15T10:00:00Z"
}
```

## Condition Types

### Threshold Conditions

Compare a real-time metric against a fixed value.

| Metric | Unit | Description |
|---|---|---|
| `kp_index` | 0-9 | Planetary K-index |
| `solar_wind_speed` | km/s | Solar wind bulk velocity |
| `proton_flux_10mev` | pfu | Proton flux > 10 MeV |
| `proton_flux_100mev` | pfu | Proton flux > 100 MeV |
| `xray_flux` | W/m^2 | GOES 1-8 Angstrom X-ray flux |
| `bz_component` | nT | IMF Bz component |
| `dst_index` | nT | Disturbance storm time index |

**Operators:** `gt`, `gte`, `lt`, `lte`, `eq`

### Event Conditions

Trigger on specific event types.

| Metric | Values | Description |
|---|---|---|
| `solar_flare_class` | `C1.0` - `X99` | Minimum flare class |
| `cme_earth_directed` | `true` | Any Earth-directed CME |
| `cme_speed` | km/s | CME speed threshold |
| `geomagnetic_storm` | `G1` - `G5` | Minimum storm category |
| `radiation_storm` | `S1` - `S5` | Minimum radiation storm level |

### Compound Conditions

Combine multiple conditions with `AND` / `OR` logic:

```json
{
  "condition": {
    "operator": "and",
    "conditions": [
      { "metric": "kp_index", "operator": "gte", "value": 5 },
      { "metric": "bz_component", "operator": "lt", "value": -10 }
    ]
  }
}
```

## Severity Levels

| Level | Intended Use |
|---|---|
| `info` | Routine updates, no action needed |
| `warning` | Elevated conditions, team should be aware |
| `critical` | Immediate action may be required |
| `emergency` | Severe event in progress, activate response plan |

## Cooldown

The `cooldown_minutes` field prevents alert fatigue by suppressing duplicate notifications for the same rule within the specified window. Default is 15 minutes. Set to `0` for no cooldown.

## Managing Rules

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/v2/alerts/rules` | Create a new alert rule |
| `GET` | `/v2/alerts/rules` | List all rules |
| `GET` | `/v2/alerts/rules/{id}` | Get a specific rule |
| `PATCH` | `/v2/alerts/rules/{id}` | Update a rule |
| `DELETE` | `/v2/alerts/rules/{id}` | Delete a rule |
| `POST` | `/v2/alerts/rules/{id}/test` | Send a test notification |
