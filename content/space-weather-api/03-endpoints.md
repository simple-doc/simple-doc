---
title: API Endpoints
---

# API Endpoints

![API Endpoints Overview](static/images/api-endpoints.svg)

## Solar Events

### List Solar Flares

```
GET /v2/events/solar-flares
```

Returns a paginated list of solar flare events.

**Query Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `start_date` | string | Start of date range (ISO 8601) |
| `end_date` | string | End of date range (ISO 8601) |
| `min_class` | string | Minimum flare class (e.g. `C1.0`, `M5.0`, `X1.0`) |
| `source_region` | string | Active region number (e.g. `AR3842`) |
| `page` | integer | Page number (default: 1) |
| `per_page` | integer | Results per page (default: 50, max: 500) |

**Example:**

```bash
curl -H "X-API-Key: sf_live_abc123" \
  "https://api.solarflux.dev/v2/events/solar-flares?start_date=2025-04-01&min_class=M1.0"
```

### List Coronal Mass Ejections

```
GET /v2/events/cmes
```

Returns CME events with trajectory analysis.

**Query Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `start_date` | string | Start of date range |
| `end_date` | string | End of date range |
| `min_speed` | integer | Minimum speed in km/s |
| `earth_directed` | boolean | Filter to Earth-directed CMEs only |

### List Geomagnetic Storms

```
GET /v2/events/geomagnetic-storms
```

Returns geomagnetic storm events with Kp index readings.

**Query Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `start_date` | string | Start of date range |
| `end_date` | string | End of date range |
| `min_kp` | integer | Minimum Kp index (0-9) |
| `category` | string | NOAA storm category: `G1`, `G2`, `G3`, `G4`, `G5` |

## Current Conditions

### Get Current Conditions

```
GET /v2/conditions/current
```

Returns the latest space weather conditions snapshot.

**Response:**

```json
{
  "timestamp": "2025-04-12T14:30:00Z",
  "solar_wind": {
    "speed_km_s": 423.7,
    "density_p_cm3": 5.2,
    "temperature_k": 98400
  },
  "interplanetary_magnetic_field": {
    "bt_nt": 6.8,
    "bz_nt": -3.2
  },
  "kp_index": {
    "current": 4,
    "predicted_3h": 5
  },
  "proton_flux": {
    "gt_10mev": 2.3,
    "gt_100mev": 0.04
  },
  "xray_flux": {
    "short_channel": 3.2e-7,
    "long_channel": 1.8e-6
  }
}
```

### Get Kp Index History

```
GET /v2/conditions/kp-index
```

Returns Kp index values over a specified time range.

## Forecasts

### Get 3-Day Forecast

```
GET /v2/forecasts/3-day
```

Returns a 3-day probabilistic space weather forecast including solar flare probabilities, geomagnetic storm likelihood, and radiation storm risk.

### Get 27-Day Outlook

```
GET /v2/forecasts/27-day
```

Returns a solar rotation (27-day) forecast based on recurrent patterns.

### Get Aurora Forecast

```
GET /v2/forecasts/aurora
```

Returns hemispheric aurora probability maps. Specify hemisphere with `?hemisphere=north` or `?hemisphere=south`.

## Pagination

All list endpoints support cursor-based pagination:

```json
{
  "data": [...],
  "pagination": {
    "total": 342,
    "page": 1,
    "per_page": 50,
    "total_pages": 7
  }
}
```

Maximum `per_page` is 500. Default is 50.
