# QuantumField

**TLS, PKI & Crypto-Risk Intelligence Platform**

QuantumField inventories internet-facing TLS certificates, turns protocol and PKI weaknesses into prioritized findings, and adds a rule-based crypto-agility assessment for future cryptographic migrations.

> **Project maturity:** portfolio-grade security engineering prototype with an explicit production-hardening roadmap. It is not presented as a production security service.

## Scope and honest positioning

QuantumField performs TLS/PKI discovery and risk analysis. Its crypto-agility score can help organize a future post-quantum migration backlog, but the project does **not** implement:

- ML-KEM/Kyber key exchange;
- ML-DSA/Dilithium signatures;
- hybrid post-quantum TLS;
- post-quantum X.509 certificates;
- a production CA, HSM, or standards-compliant PQC cryptosystem.

The scoring engine is deliberately deterministic and explainable. There is no AI model hidden behind the score, and the repository should not be described as an AI project.

## Why this project exists

Most TLS tools stop at “valid” or “expired.” QuantumField connects three concerns that security teams increasingly need to manage together:

- operational TLS hygiene: expiry, trust chain, hostname coverage, HSTS, protocol versions, and key strength;
- PKI inventory: issuer, subject, SANs, serial number, fingerprint, signature algorithm, public-key algorithm, and lifecycle;
- crypto agility: classical RSA/ECC dependency, TLS 1.3 adoption, legacy protocol removal, and certificate rotation readiness.

This makes the project interview-defensible: it performs real network inspection, processes jobs asynchronously, persists evidence, applies explicit scoring models, and presents results in an analyst-focused dashboard.

## Architecture

```mermaid
flowchart LR
    U[React + TypeScript UI] -->|JWT / REST| A[Go + Gin API]
    A --> P[(PostgreSQL)]
    A -->|LPUSH scan job| R[(Redis)]
    W[Go scan worker] -->|BRPOP job| R
    W -->|TLS handshake / HTTPS probe| T[Public domain]
    W -->|Certificate, findings, scores| P
    A -->|Dashboard and inventory data| U
```

### Scan lifecycle

1. An authenticated user adds a fully qualified domain and TLS port.
2. The API creates a `queued` scan and pushes a job into Redis.
3. The worker resolves the target and rejects private/reserved network destinations.
4. It retrieves the certificate chain, then verifies trust and hostname coverage independently.
5. Additional probes inspect the negotiated TLS version, cipher suite, TLS 1.0/1.1 acceptance, and HSTS.
6. The scoring engine creates findings, a 0–100 risk score, and a 0–100 crypto-agility score.
7. PostgreSQL stores immutable scan evidence and updates the asset’s current posture.

## Features

- JWT registration, login, session restoration, and tenant-scoped data access
- analyst/admin role claims and reusable role middleware
- domain inventory with labels, ports, status, and scan history
- Redis-backed asynchronous scan queue with a separately scalable worker
- X.509 subject, issuer, CN, SAN, serial, validity, fingerprint, key/signature algorithms, and key size
- independent certificate-chain and hostname validation
- TLS version, cipher suite, TLS 1.0/1.1, and HSTS inspection
- persisted findings with severity, evidence, and remediation
- explainable risk and crypto-agility readiness models
- dashboard, asset drill-down, scan jobs, findings, certificate inventory, crypto-agility readiness, and reports
- JSON report export
- Docker Compose, GitHub Actions, and an optional Kubernetes template
- resolve-once DNS pinning that blocks loopback, private, link-local, multicast, documentation, benchmark, and reserved networks

## Technology

| Layer | Technology |
|---|---|
| API and worker | Go 1.23, Gin, GORM |
| Durable data | PostgreSQL 16 |
| Job queue | Redis 7 |
| Authentication | JWT HS256, bcrypt |
| Scanner | Go `crypto/tls`, `crypto/x509`, `net/http` |
| Frontend | React 18, TypeScript, Vite, Tailwind CSS |
| Runtime | Docker Compose, Nginx |
| CI | GitHub Actions |
| Optional deployment | Kubernetes |

## Quick start with Docker

Requirements: Docker Engine and Docker Compose v2.

```bash
cp .env.example .env
docker compose up --build
```

Open:

- UI: <http://localhost:3000>
- API health: <http://localhost:8080/health>

When `SEED_DEMO=true`, the development-only account is:

```text
demo@quantumfield.dev
QuantumField123!
```

Change `JWT_SECRET`, the database password, and the demo setting before any shared deployment.

## Demo status

No hosted deployment is currently claimed. Run the Compose stack for the live application. A short, reproducible recording plan is available in [docs/DEMO.md](docs/DEMO.md); add a real GIF or MP4 link only after recording the running application.

## Local development

Requirements: Go 1.23+, Node.js 22+, PostgreSQL, and Redis.

```bash
# Terminal 1
cd backend
go mod download
go run ./cmd/api

# Terminal 2
cd backend
go run ./cmd/worker

# Terminal 3
cd frontend
npm install
npm run dev
```

For host-run services, set `DATABASE_URL` and `REDIS_ADDR` to localhost values. Vite proxies `/api` to `http://localhost:8080`.

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `DATABASE_URL` | local development URL | PostgreSQL DSN |
| `REDIS_ADDR` | `localhost:6379` | Redis host and port |
| `REDIS_PASSWORD` | empty | Optional Redis password |
| `REDIS_DB` | `0` | Redis logical database |
| `JWT_SECRET` | development value | JWT signing secret; use 32+ random characters |
| `JWT_TTL_HOURS` | `24` | Access-token lifetime |
| `API_PORT` | `8080` | API listen port |
| `CORS_ORIGINS` | `http://localhost:5173` | Comma-separated browser origins |
| `SCAN_TIMEOUT_SECONDS` | `15` | Per-network-operation timeout |
| `SEED_DEMO` | `false` | Create the development demo account |
| `VITE_API_URL` | `/api` | Frontend API base URL |

## API

All protected routes require `Authorization: Bearer <jwt>`.

### Authentication

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/auth/register` | Create an analyst account |
| `POST` | `/api/auth/login` | Issue a JWT |
| `GET` | `/api/auth/me` | Return the current user |

### Assets and scans

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/assets` | Add a domain and TLS port |
| `GET` | `/api/assets` | List the user’s assets |
| `GET` | `/api/assets/:id` | Asset details and scan history |
| `DELETE` | `/api/assets/:id` | Soft-delete an asset |
| `POST` | `/api/assets/:id/scan` | Queue a TLS scan |
| `GET` | `/api/scans` | List scan jobs |
| `GET` | `/api/scans/:id` | Full scan evidence |

### Intelligence

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/assets/:id/certificate` | Latest asset certificate |
| `GET` | `/api/assets/:id/findings` | Asset finding history |
| `GET` | `/api/assets/:id/pqc-assessment` | Latest asset assessment |
| `GET` | `/api/certificates` | Portfolio certificate inventory |
| `GET` | `/api/findings` | Portfolio findings; optional `severity` query |
| `GET` | `/api/pqc-assessments` | Portfolio readiness history |
| `GET` | `/api/dashboard` | Dashboard aggregates |
| `GET` | `/api/reports/summary` | Reporting aggregates |

Example:

```bash
curl -X POST http://localhost:8080/api/assets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"domain":"example.com","port":443,"label":"Public website"}'
```

## Data model

| Entity | Important fields | Relationship |
|---|---|---|
| `users` | name, email, password hash, role | owns assets |
| `assets` | domain, port, status, current risk/agility scores | belongs to user; has scans |
| `scans` | status, timing, TLS version, cipher, scores, error | belongs to asset |
| `certificates` | identity, issuer, validity, algorithms, fingerprint, validation | one per completed scan |
| `findings` | type, severity, evidence, remediation, status | many per scan |
| `pqc_assessments` | score, grade, dependency flags, rationale | one per completed scan |

GORM runs idempotent schema migration at API/worker startup. UUIDs are application-generated and PostgreSQL data is retained in a named Docker volume.

## Risk scoring model

Risk starts at 0, adds the weight of each observed finding, and is capped at 100:

| Severity | Points |
|---|---:|
| Critical | 30 |
| High | 20 |
| Medium | 10 |
| Low | 5 |
| Informational | 0 |

Current detections include:

- expired or soon-to-expire certificates;
- invalid certificate chain;
- hostname mismatch;
- missing HSTS when an HTTPS response was observed;
- TLS 1.0 or TLS 1.1 support;
- RSA keys below 2048 bits or ECC keys below 256 bits;
- RSA/ECC/classical public-key dependency as a low-severity migration signal.

The model is intentionally explainable rather than “AI generated.” A production deployment should tune weights using asset criticality, exposure, compensating controls, policy, and finding age.

## Crypto-agility readiness model

Readiness starts at 100 and applies explicit deductions:

| Condition | Deduction |
|---|---:|
| RSA or ECC certificate dependency | 35 |
| Classical certificate signature | 20 |
| Negotiated TLS is below TLS 1.3 | 15 |
| TLS 1.0 or TLS 1.1 is accepted | 15 |
| Certificate lifecycle fails the crypto-agility baseline | 15 |

Grades: A = 80–100, B = 65–79, C = 50–64, D = 30–49, F = 0–29.

This is a migration-preparedness indicator, not proof of post-quantum cryptography. Conventional public Web PKI generally depends on RSA or elliptic curves, so low scores are expected. The useful output is the migration backlog: inventory dependencies, modernize TLS, automate rotation, and prepare for standards-based cryptographic changes.

## Security design

- Passwords are bcrypt-hashed and never returned.
- JWTs contain user ID, role, issue time, and expiry.
- Inventory queries are scoped to the authenticated user.
- Duplicate active scans per asset are rejected.
- Network targets are resolved exactly once per scan.
- The complete DNS answer set is rejected if any address is loopback, private, unspecified, link-local, multicast, documentation-only, benchmark, carrier-grade NAT, or otherwise reserved.
- The selected validated IP is pinned for the certificate handshake, HSTS request, and legacy-TLS probes while the original hostname is retained for SNI, HTTP Host, and certificate hostname verification.
- Certificate verification is explicit, allowing invalid certificates to be inventoried without treating them as trusted.
- Nginx adds basic browser hardening headers.
- Secrets are environment-injected and not committed.

### Production hardening still required

- replace auto-migration with versioned migrations;
- use an asymmetric identity provider or short-lived tokens with refresh rotation;
- rate-limit authentication and scan endpoints;
- add MFA, email verification, password reset, and audit events;
- enforce organization-level egress policy and scan authorization;
- use managed PostgreSQL/Redis with TLS, backups, and secret management;
- add distributed locks, retries, dead-letter queues, and worker concurrency controls;
- perform threat modeling, SAST/DAST, dependency scanning, and penetration testing.

## Repository layout

```text
.
├── backend
│   ├── cmd/api                 # Gin API entry point
│   ├── cmd/worker              # Redis scan worker
│   └── internal
│       ├── auth                # Password and JWT service
│       ├── database            # PostgreSQL and migration
│       ├── httpapi             # Handlers and routes
│       ├── models              # Persistent entities
│       ├── queue               # Redis job client
│       ├── scanner             # TLS/X.509/HSTS probes
│       ├── scoring             # Risk and crypto-agility models
│       ├── target              # Domain normalization
│       └── worker              # Scan orchestration
├── frontend/src
│   ├── components              # Shell and reusable UI
│   ├── context                 # Authentication state
│   ├── lib                     # API and formatting
│   └── pages                   # Product screens
├── deploy/kubernetes.yaml      # Optional deployment template
├── docker-compose.yml
└── .github/workflows/ci.yml
```

## Automated validation

The default suite covers password hashing, JWT creation/expiry, queue payload validation, domain normalization, risk scoring, reserved-address rejection, DNS pin selection, and pinned HSTS behavior. CI also starts PostgreSQL and Redis and exercises the complete register → create asset → enqueue scan API flow.

```bash
cd backend
gofmt -w ./cmd ./internal
go vet ./...
go test ./...
go build ./cmd/api ./cmd/worker

cd ../frontend
npm run lint
npm run build
```

## Resume positioning

**QuantumField — TLS, PKI & Crypto-Risk Intelligence Platform**

- Built a Go/React cybersecurity platform that asynchronously scans public TLS endpoints, validates X.509 trust and hostname posture, inventories certificate cryptography, and persists evidence in PostgreSQL.
- Designed a Redis-backed worker pipeline and explainable 0–100 risk/crypto-agility models covering expiry, chain errors, legacy TLS, HSTS, RSA/ECC dependency, TLS 1.3, and certificate rotation practices.
- Containerized API, worker, database, queue, and frontend services with Docker Compose; added tenant-scoped JWT authorization, CI checks, and an optional Kubernetes deployment template.

## License

MIT
