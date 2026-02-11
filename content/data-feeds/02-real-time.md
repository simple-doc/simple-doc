---
title: Real-Time Feeds
---

# Real-Time Feeds

Real-time feeds deliver measurements within seconds of acquisition from satellite instruments. All real-time data goes through quality control checks before dissemination.

## Solar Wind Feed

Reports plasma parameters and interplanetary magnetic field components from the L1 Lagrange point (~1.5 million km from Earth).

**Message format:**

```json
{
  "feed": "solar_wind",
  "timestamp": "2025-04-12T14:30:00Z",
  "source": "DSCOVR",
  "values": {
    "speed_km_s": 423.7,
    "density_p_cm3": 5.2,
    "temperature_k": 98400,
    "bt_nt": 6.8,
    "bx_nt": 2.1,
    "by_nt": -4.3,
    "bz_nt": -3.2
  },
  "quality_flag": "definitive"
}
```

**Quality flags:** `preliminary` (< 5 min old), `definitive` (verified), `estimated` (gap-filled)

## X-Ray Flux Feed

Reports solar X-ray irradiance in two wavelength bands, used for flare detection and classification.

```json
{
  "feed": "xray_flux",
  "timestamp": "2025-04-12T14:30:00Z",
  "source": "GOES-18",
  "values": {
    "short_channel_w_m2": 3.2e-7,
    "long_channel_w_m2": 1.8e-6,
    "flare_class": "B1.8",
    "background_level": "B1.2"
  }
}
```

## Proton Flux Feed

High-energy proton measurements used for radiation storm assessment.

```json
{
  "feed": "proton_flux",
  "timestamp": "2025-04-12T14:30:00Z",
  "source": "GOES-18",
  "values": {
    "gt_1mev_pfu": 1200.5,
    "gt_5mev_pfu": 45.3,
    "gt_10mev_pfu": 2.3,
    "gt_30mev_pfu": 0.18,
    "gt_50mev_pfu": 0.07,
    "gt_100mev_pfu": 0.04
  }
}
```

## Kp Index Feed

Planetary geomagnetic activity index, ranging from 0 (quiet) to 9 (extreme storm).

```json
{
  "feed": "kp_index",
  "timestamp": "2025-04-12T12:00:00Z",
  "values": {
    "kp": 4,
    "kp_fraction": 4.33,
    "ap": 32,
    "status": "estimated"
  }
}
```

## Subscribing to Multiple Feeds

You can subscribe to any combination of feeds on a single connection:

```json
{
  "action": "subscribe",
  "feeds": ["solar_wind", "xray_flux", "proton_flux", "kp_index"]
}
```

To unsubscribe from a feed without disconnecting:

```json
{
  "action": "unsubscribe",
  "feeds": ["proton_flux"]
}
```

## Heartbeat

The server sends a heartbeat message every 30 seconds on idle connections:

```json
{
  "type": "heartbeat",
  "timestamp": "2025-04-12T14:30:30Z"
}
```

If no heartbeat is received within 60 seconds, the client should reconnect.
