---
title: Data Formats
---

# Data Formats

SolarFlux provides data in multiple formats to suit different workflows.

## JSON (Default)

All API responses and real-time feeds use JSON. Numbers use IEEE 754 double precision. Null values indicate missing measurements.

```json
{
  "timestamp": "2025-04-12T14:30:00Z",
  "speed_km_s": 423.7,
  "density_p_cm3": null,
  "bz_nt": -3.2
}
```

## CSV

Bulk exports and historical query results can be returned as CSV.

```
timestamp,speed_km_s,density_p_cm3,bt_nt,bz_nt,quality
2025-04-12T14:30:00Z,423.7,5.2,6.8,-3.2,definitive
2025-04-12T14:31:00Z,425.1,5.1,6.9,-3.4,definitive
2025-04-12T14:32:00Z,424.3,,7.0,-3.1,definitive
```

Missing values are represented as empty fields. Request CSV format by adding `?format=csv` or setting `Accept: text/csv`.

## Apache Parquet

Recommended for large bulk exports. Parquet provides columnar compression and is natively supported by pandas, Spark, DuckDB, and most data analysis tools.

```python
import pandas as pd

df = pd.read_parquet("solar_wind_2024.parquet")
print(df.describe())
```

Typical compression ratio is 8-12x compared to raw CSV.

## Units Reference

All measurements use SI units unless otherwise noted.

| Measurement | Unit | Symbol |
|---|---|---|
| Solar wind speed | kilometers per second | km/s |
| Proton density | protons per cubic centimeter | p/cm^3 |
| Temperature | Kelvin | K |
| Magnetic field | nanotesla | nT |
| X-ray flux | watts per square meter | W/m^2 |
| Proton flux | particle flux units | pfu |
| Kp index | dimensionless | — |
| Dst index | nanotesla | nT |

## Coordinate Systems

Solar event locations use heliographic coordinates:

| System | Description |
|---|---|
| **Stonyhurst** | Longitude measured from the central meridian as seen from Earth. Standard for most SolarFlux data. |
| **Carrington** | Longitude measured in the Sun's rotating frame. Used for 27-day recurrence analysis. |

The `coordinate_system` field in event data specifies which system is used. Default is Stonyhurst.

## Timestamps

All timestamps are ISO 8601 in UTC:

```
2025-04-12T14:30:00Z
2025-04-12T14:30:00.123Z  (with milliseconds)
```

Date-only parameters accept `YYYY-MM-DD` format and are interpreted as the start of that day in UTC.
