---
title: Error Handling
---

# Error Handling

The API uses standard HTTP status codes and returns structured error responses.

## Error Response Format

```json
{
  "error": {
    "code": "invalid_parameter",
    "message": "The min_class parameter must be a valid flare class (e.g. C1.0, M5.0, X1.0).",
    "param": "min_class",
    "request_id": "req_7f3a2b1c"
  }
}
```

## Status Codes

| Code | Meaning |
|---|---|
| `200` | Success |
| `400` | Bad Request — invalid parameters |
| `401` | Unauthorized — missing or invalid authentication |
| `403` | Forbidden — insufficient scope |
| `404` | Not Found — resource does not exist |
| `429` | Too Many Requests — rate limit exceeded |
| `500` | Internal Server Error |
| `503` | Service Unavailable — upstream data source offline |

## Error Codes

| Error Code | Description |
|---|---|
| `invalid_parameter` | A query parameter has an invalid value |
| `missing_parameter` | A required parameter was not provided |
| `invalid_date_range` | The start_date is after the end_date |
| `date_range_too_wide` | Maximum date range is 365 days per request |
| `resource_not_found` | The requested event or resource does not exist |
| `rate_limit_exceeded` | Too many requests; retry after the `Retry-After` period |
| `upstream_unavailable` | A data source (e.g., GOES satellite feed) is temporarily offline |

## Retry Strategy

For `429` and `503` errors, implement exponential backoff:

1. Wait for the duration specified in the `Retry-After` header
2. If no header is present, start with a 1-second delay
3. Double the delay on each subsequent retry
4. Cap the maximum delay at 60 seconds
5. Add random jitter (0-500ms) to avoid thundering herd

## Request IDs

Every response includes an `X-Request-Id` header. Include this ID when contacting support about a specific request.
