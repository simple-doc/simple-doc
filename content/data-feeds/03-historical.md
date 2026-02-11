---
title: Historical Data
---

# Historical Data

The SolarFlux historical data archive contains space weather observations spanning over 70 years, from Solar Cycle 19 (1954) to the present. Historical data is available via bulk download and query APIs.

## Bulk Data Downloads

Large datasets are available as compressed CSV or Parquet files via authenticated HTTPS download.

### Request a Bulk Export

```bash
curl -X POST https://api.solarflux.dev/v2/feeds/exports \
  -H "X-API-Key: sf_live_abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "feed": "solar_wind",
    "start_date": "2024-01-01",
    "end_date": "2024-12-31",
    "format": "parquet",
    "resolution": "1min"
  }'
```

**Response:**

```json
{
  "export_id": "exp_4a5b6c7d",
  "status": "processing",
  "estimated_size_mb": 847,
  "estimated_ready_at": "2025-04-12T15:10:00Z"
}
```

### Check Export Status

```bash
curl -H "X-API-Key: sf_live_abc123" \
  "https://api.solarflux.dev/v2/feeds/exports/exp_4a5b6c7d"
```

When `status` is `ready`, a signed `download_url` will be included (valid for 24 hours).

## Available Resolutions

| Resolution | Available From | Use Case |
|---|---|---|
| 1 second | 2016-present | High-cadence analysis |
| 1 minute | 1995-present | Standard research |
| 5 minute | 1986-present | Long-term studies |
| 1 hour | 1964-present | Solar cycle analysis |
| 1 day | 1954-present | Multi-cycle trends |

## Query API

For smaller date ranges, query historical data directly:

```bash
curl -H "X-API-Key: sf_live_abc123" \
  "https://api.solarflux.dev/v2/feeds/history/solar_wind?start=2025-04-10&end=2025-04-12&resolution=5min"
```

**Response:**

```json
{
  "feed": "solar_wind",
  "resolution": "5min",
  "count": 576,
  "data": [
    {
      "timestamp": "2025-04-10T00:00:00Z",
      "speed_km_s": 387.2,
      "density_p_cm3": 4.8,
      "bt_nt": 5.1,
      "bz_nt": 1.3
    },
    ...
  ]
}
```

Maximum query range depends on resolution:

| Resolution | Max Range Per Query |
|---|---|
| 1 second | 1 hour |
| 1 minute | 7 days |
| 5 minute | 30 days |
| 1 hour | 1 year |
| 1 day | Full archive |

## Data Quality

Historical data is classified by quality:

- **Definitive** — Final, quality-controlled values. Available ~6 months after observation.
- **Provisional** — Preliminary quality control applied. Available within 1 week.
- **Real-time** — Minimal quality control. Available immediately.

The `quality` field in each data point indicates its classification.
