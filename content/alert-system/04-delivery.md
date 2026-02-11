---
title: Delivery Channels
---

# Delivery Channels

Alerts can be routed to one or more delivery channels. Each alert rule specifies a primary channel; additional channels can be added via escalation policies.

## Webhook

Delivers a JSON payload via HTTP POST to your endpoint.

```json
{
  "type": "webhook",
  "url": "https://ops.example.com/hooks/space-weather",
  "headers": {
    "Authorization": "Bearer your-token"
  },
  "timeout_seconds": 10,
  "retry_count": 3
}
```

**Webhook payload:**

```json
{
  "alert_id": "alt_9c3d4e5f",
  "rule_id": "rule_8f2a3b4c",
  "rule_name": "High Kp Alert",
  "severity": "critical",
  "triggered_at": "2025-04-14T06:12:00Z",
  "condition": {
    "metric": "kp_index",
    "operator": "gte",
    "threshold": 7,
    "actual_value": 8
  },
  "event_id": "GST-2025-04-14T05:30Z",
  "summary": "Kp index reached 8 (G4 — Severe geomagnetic storm)"
}
```

SolarFlux signs webhook payloads with HMAC-SHA256. Verify the `X-SolarFlux-Signature` header using your webhook signing secret.

## Email

Sends formatted HTML email alerts.

```json
{
  "type": "email",
  "address": "ops-team@example.com"
}
```

Email alerts include a summary, current readings, and a link to the event detail page.

## SMS

Sends short text alerts to a phone number.

```json
{
  "type": "sms",
  "phone_number": "+14155551234"
}
```

SMS alerts are limited to 160 characters and include the severity, metric, and current value.

## Slack

Posts to a Slack channel via incoming webhook.

```json
{
  "type": "slack",
  "webhook_url": "https://hooks.slack.com/services/T00/B00/xxxxx"
}
```

Slack messages include rich formatting with severity color coding and direct links to the SolarFlux dashboard.

## PagerDuty

Triggers incidents in PagerDuty for on-call routing.

```json
{
  "type": "pagerduty",
  "routing_key": "your-pagerduty-integration-key",
  "dedup_key_prefix": "solarflux"
}
```

Alerts with `critical` and `emergency` severity create PagerDuty incidents. Lower severities are sent as informational events.

## Delivery Guarantees

- Webhooks are retried up to 3 times with exponential backoff (1s, 5s, 25s)
- If all retries fail, the alert is logged and visible in the Alert History dashboard
- Email and SMS delivery is best-effort with a 99.5% SLA
- All delivered alerts are recorded in the audit log accessible via `GET /v2/alerts/history`
