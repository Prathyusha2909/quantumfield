# QuantumField API

Base URL: `http://localhost:8080/api`

Protected endpoints require:

```http
Authorization: Bearer <jwt>
Content-Type: application/json
```

Errors use:

```json
{
  "error": "human-readable message"
}
```

Authentication routes share a limit of 10 requests per client IP per minute. Scan-trigger routes allow 10 requests per authenticated user per 10 minutes. A `429` response includes `Retry-After`, `X-RateLimit-Limit`, and `X-RateLimit-Remaining` headers.

## Authentication

### `POST /auth/register`

```json
{
  "name": "Demo User",
  "email": "demo@example.com",
  "password": "StrongPassword123!"
}
```

Response: `201 Created`

```json
{
  "token": "<jwt>",
  "user": {
    "id": "uuid",
    "name": "Demo User",
    "email": "demo@example.com",
    "role": "analyst"
  }
}
```

### `POST /auth/login`

```json
{
  "email": "demo@example.com",
  "password": "StrongPassword123!"
}
```

Response: `200 OK` with the same session shape as registration.

### `GET /auth/me`

Returns the authenticated user.

## Assets

### `POST /assets`

```json
{
  "domain": "https://example.com/path",
  "port": 443,
  "label": "Public website"
}
```

The API normalizes the input to a fully qualified hostname and port.

### `GET /assets`

Returns the authenticated user’s active assets.

### `GET /assets/:id`

Returns an asset and up to 20 recent scans with certificate and crypto-agility assessment data.

### `DELETE /assets/:id`

Soft-deletes the asset. Response: `204 No Content`.

### `POST /assets/:id/scan`

Creates a scan record and pushes a job to Redis.

Response: `202 Accepted`

```json
{
  "id": "scan-uuid",
  "asset_id": "asset-uuid",
  "status": "queued",
  "retry_count": 0,
  "max_retries": 3
}
```

Only one queued/running scan is allowed per asset.

## Scan evidence

### `GET /scans`

Returns up to 200 recent user-scoped scan jobs.

### `GET /scans/:id`

Returns the full scan, including:

- negotiated TLS version and cipher suite;
- retry count, last error, and terminal failure timestamp;
- leaf certificate;
- findings;
- crypto-agility assessment.

### `GET /assets/:id/certificate`

Returns the latest observed leaf certificate.

### `GET /assets/:id/findings`

Returns finding history for one asset.

### `GET /assets/:id/pqc-assessment`

Returns the latest rule-based crypto-agility assessment. The legacy route name is retained for API compatibility; it does not claim implemented PQC.

## Portfolio intelligence

### `GET /dashboard`

Returns summary metrics, assets, recent scans, and priority findings.

### `GET /findings?severity=critical`

Returns portfolio findings with optional severity filtering.

### `GET /certificates`

Returns certificate inventory across the user’s assets.

### `GET /pqc-assessments`

Returns crypto-agility assessment history. The route name is retained for compatibility.

### `GET /reports/summary`

Returns findings-by-severity, certificate algorithms, and certificates expiring within 90 days.

### `GET /reports/export`

Returns the current JSON report and records a `REPORT_EXPORTED` audit event.

### `GET /audit-logs`

Returns up to 200 recent audit events belonging to the authenticated user.

## Health

### `GET /health`

Checks PostgreSQL and Redis connectivity. Returns `503` when a required dependency is unavailable.
