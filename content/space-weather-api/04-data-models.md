---
title: Data Models
---

# Data Models

This section describes the core data structures returned by the Space Weather API.

## SolarFlare

Represents a solar flare event detected by GOES X-ray sensors.

| Field | Type | Description |
|---|---|---|
| `event_id` | string | Unique event identifier (e.g. `FLR-2025-04-12T09:22Z`) |
| `type` | string | Always `solar_flare` |
| `class` | string | Flare class: A, B, C, M, or X with intensity (e.g. `X2.1`) |
| `begin_time` | string | Flare start time (ISO 8601) |
| `peak_time` | string | Peak intensity time |
| `end_time` | string | Flare end time |
| `source_region` | string | NOAA active region number |
| `location` | object | Heliographic coordinates (`latitude`, `longitude`) |
| `duration_minutes` | integer | Total duration in minutes |
| `peak_flux_w_m2` | number | Peak X-ray flux in W/m^2 |
| `associated_cme` | boolean | Whether a CME was associated with this flare |

## CoronalMassEjection

Represents a CME event tracked by coronagraph imagery.

| Field | Type | Description |
|---|---|---|
| `event_id` | string | Unique identifier (e.g. `CME-2025-04-12T10:48Z`) |
| `type` | string | Always `cme` |
| `start_time` | string | First appearance in coronagraph |
| `speed_km_s` | number | Linear speed in km/s |
| `half_angle_deg` | number | Angular half-width in degrees |
| `is_earth_directed` | boolean | Whether the CME is headed toward Earth |
| `estimated_arrival` | string | Predicted Earth arrival time (if Earth-directed) |
| `source_event` | string | Related solar flare event_id, if any |
| `analysis_level` | integer | 0 = preliminary, 1 = refined, 2 = final |

## GeomagneticStorm

Represents a geomagnetic storm period.

| Field | Type | Description |
|---|---|---|
| `event_id` | string | Unique identifier |
| `type` | string | Always `geomagnetic_storm` |
| `begin_time` | string | Storm onset time |
| `end_time` | string | Storm recovery time |
| `peak_kp` | integer | Maximum Kp index during the storm (0-9) |
| `category` | string | NOAA category: `G1` (minor) through `G5` (extreme) |
| `dst_min_nt` | number | Minimum Dst index in nanotesla |
| `trigger_cme` | string | CME event_id that triggered the storm, if identified |

## Conditions

Real-time space environment readings.

| Field | Type | Description |
|---|---|---|
| `timestamp` | string | Measurement time |
| `solar_wind.speed_km_s` | number | Solar wind bulk speed |
| `solar_wind.density_p_cm3` | number | Proton density |
| `solar_wind.temperature_k` | number | Proton temperature |
| `interplanetary_magnetic_field.bt_nt` | number | Total IMF magnitude in nanotesla |
| `interplanetary_magnetic_field.bz_nt` | number | IMF Bz component (negative = geoeffective) |
| `kp_index.current` | integer | Current 3-hour Kp index |
| `kp_index.predicted_3h` | integer | Predicted next Kp value |
| `proton_flux.gt_10mev` | number | Proton flux > 10 MeV (pfu) |
| `xray_flux.long_channel` | number | 1-8 Angstrom X-ray flux (W/m^2) |
