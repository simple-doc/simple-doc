---
title: Code Examples
---

# Code Examples

Practical examples for common data feed integration patterns.

## Python — Real-Time Dashboard

Connect to the WebSocket feed and plot solar wind speed in real time:

```python
import asyncio
import json
import websockets
from collections import deque
from datetime import datetime

API_KEY = "sf_live_abc123"
FEED_URL = "wss://feeds.solarflux.dev/v2/stream"

readings = deque(maxlen=300)  # Keep last 5 hours at 1-min cadence

async def monitor():
    async with websockets.connect(FEED_URL) as ws:
        await ws.send(json.dumps({
            "action": "subscribe",
            "feeds": ["solar_wind"],
            "api_key": API_KEY
        }))

        async for message in ws:
            data = json.loads(message)
            if data.get("feed") == "solar_wind":
                readings.append({
                    "time": data["timestamp"],
                    "speed": data["values"]["speed_km_s"],
                    "bz": data["values"]["bz_nt"]
                })
                speed = data["values"]["speed_km_s"]
                bz = data["values"]["bz_nt"]
                print(f"[{data['timestamp']}] Wind: {speed} km/s | Bz: {bz} nT")

asyncio.run(monitor())
```

## Go — Alert Webhook Server

Handle incoming SolarFlux webhook alerts:

```go
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "io"
    "log"
    "net/http"
)

const webhookSecret = "whsec_your_signing_secret"

type Alert struct {
    AlertID   string `json:"alert_id"`
    RuleName  string `json:"rule_name"`
    Severity  string `json:"severity"`
    Summary   string `json:"summary"`
    Condition struct {
        Metric      string  `json:"metric"`
        Threshold   float64 `json:"threshold"`
        ActualValue float64 `json:"actual_value"`
    } `json:"condition"`
}

func verifySignature(body []byte, signature string) bool {
    mac := hmac.New(sha256.New, []byte(webhookSecret))
    mac.Write(body)
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}

func handleAlert(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    sig := r.Header.Get("X-SolarFlux-Signature")

    if !verifySignature(body, sig) {
        http.Error(w, "invalid signature", http.StatusUnauthorized)
        return
    }

    var alert Alert
    json.Unmarshal(body, &alert)

    log.Printf("[%s] %s: %s (value: %.2f, threshold: %.2f)",
        alert.Severity, alert.RuleName, alert.Summary,
        alert.Condition.ActualValue, alert.Condition.Threshold)

    w.WriteHeader(http.StatusOK)
}

func main() {
    http.HandleFunc("/hooks/space-weather", handleAlert)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## JavaScript — Historical Data Analysis

Fetch and analyze Kp index data for the last 30 days:

```javascript
const API_KEY = "sf_live_abc123";
const BASE_URL = "https://api.solarflux.dev/v2";

async function analyzeKpActivity() {
  const endDate = new Date().toISOString().split("T")[0];
  const startDate = new Date(Date.now() - 30 * 86400000)
    .toISOString()
    .split("T")[0];

  const response = await fetch(
    `${BASE_URL}/feeds/history/kp_index?start=${startDate}&end=${endDate}&resolution=3h`,
    { headers: { "X-API-Key": API_KEY } }
  );

  const result = await response.json();

  const kpValues = result.data.map((d) => d.kp);
  const maxKp = Math.max(...kpValues);
  const avgKp = kpValues.reduce((a, b) => a + b, 0) / kpValues.length;
  const stormHours = kpValues.filter((kp) => kp >= 5).length * 3;

  console.log(`Period: ${startDate} to ${endDate}`);
  console.log(`Max Kp: ${maxKp}`);
  console.log(`Average Kp: ${avgKp.toFixed(2)}`);
  console.log(`Storm hours (Kp >= 5): ${stormHours}h`);
}

analyzeKpActivity();
```

## curl — Quick Checks

Check current space weather conditions:

```bash
curl -s -H "X-API-Key: sf_live_abc123" \
  https://api.solarflux.dev/v2/conditions/current | jq .
```

Get today's solar flares:

```bash
curl -s -H "X-API-Key: sf_live_abc123" \
  "https://api.solarflux.dev/v2/events/solar-flares?start_date=$(date +%Y-%m-%d)" | jq '.data[] | {class, peak_time, source_region}'
```

List your active alert rules:

```bash
curl -s -H "X-API-Key: sf_live_abc123" \
  https://api.solarflux.dev/v2/alerts/rules | jq '.data[] | {name, severity, enabled}'
```
