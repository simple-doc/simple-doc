---
title: Authentication
---

# Authentication

All requests to the SolarFlux API must be authenticated using an API key or OAuth 2.0 bearer token.

## API Key Authentication

The simplest method. Include your API key in the `X-API-Key` header:

```bash
curl -H "X-API-Key: sf_live_abc123def456" \
  https://api.solarflux.dev/v2/conditions/current
```

API keys can be generated in the [Developer Console](https://console.solarflux.dev). Each key is scoped to a specific project and environment (live or sandbox).

### Key Rotation

Keys can be rotated without downtime. When you generate a new key, the previous key remains valid for 24 hours.

## OAuth 2.0 (Client Credentials)

For server-to-server integrations, use the OAuth 2.0 client credentials flow:

```bash
curl -X POST https://auth.solarflux.dev/oauth/token \
  -d grant_type=client_credentials \
  -d client_id=YOUR_CLIENT_ID \
  -d client_secret=YOUR_CLIENT_SECRET \
  -d scope=read:events read:conditions
```

Response:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "read:events read:conditions"
}
```

Use the token in the `Authorization` header:

```bash
curl -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIs..." \
  https://api.solarflux.dev/v2/events/solar-flares
```

### Available Scopes

| Scope | Description |
|---|---|
| `read:events` | Query solar event data |
| `read:conditions` | Access current conditions and forecasts |
| `read:feeds` | Subscribe to real-time data feeds |
| `write:alerts` | Create and manage alert configurations |
| `admin:account` | Manage account settings and API keys |

## Rate Limits

| Plan | Requests / minute | Burst limit |
|---|---|---|
| Free | 60 | 10 |
| Professional | 600 | 100 |
| Enterprise | 6,000 | 1,000 |

Rate limit headers are included in every response:

```
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 594
X-RateLimit-Reset: 1714003260
```

When rate limited, the API returns `429 Too Many Requests` with a `Retry-After` header.
