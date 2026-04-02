# insider-one

A multi-channel notification microservice supporting Email, SMS, and Push notifications with priority-based routing, scheduling, batch sending, and idempotent delivery.

---

## Overview

insider-one provides an async notification platform built on an event-driven architecture. Clients submit notifications via a REST API; delivery happens asynchronously through dedicated consumer processes. Key capabilities:

- **Three channels:** Email, SMS, Push
- **Priority routing:** HIGH / MEDIUM / LOW — scale consumers per tier independently
- **Scheduling:** Deliver notifications at a future time (up to 1 year ahead)
- **Batch sending:** Up to 1,000 notifications per request
- **Idempotency:** Duplicate submissions are silently deduplicated
- **Observability:** Structured JSON logging, correlation IDs, Prometheus metrics, health check endpoint

---

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                  Presentation Layer                  │
│        REST API  (Gin)  +  Swagger UI                │
└────────────────────────┬─────────────────────────────┘
                         │  Commands / Queries
┌────────────────────────▼─────────────────────────────┐
│                 Application Layer                    │
│         CQRS — Command Handlers / Query Handlers     │
└────────────────────────┬─────────────────────────────┘
                         │  Domain interfaces
┌────────────────────────▼─────────────────────────────┐
│                   Domain Layer                       │
│     Entities · Repository interfaces · Events        │
└────────────────────────┬─────────────────────────────┘
                         │  Adapters
┌────────────────────────▼─────────────────────────────┐
│                Infrastructure Layer                  │
│  PostgreSQL · RabbitMQ · Provider Clients · Redis    │
└──────────────────────────────────────────────────────┘
```

### Event-Driven Flow

```
API ──[CreateEmailEvent]──► Create Consumer
                                 │  saves to DB
                                 └─[EmailCreatedEvent]──► Email Provider Consumer
                                                               │  calls provider API
                                                               └─ marks DELIVERED

Scheduled Job ──► polls SCHEDULED notifications ──► transitions to PENDING
```

### Deployable Components

| Component | Count | Description |
|---|---|---|
| `notification-management-api` | 1 | REST API |
| `create-{email,push,sms}-consumer` | 3 | Persist new notifications |
| `{email,push,sms}-created-consumer` | 3 | Deliver via provider |
| `cancel-{email,push,sms}-consumer` | 3 | Process cancellations |
| `{email,push,sms}-scheduled-job` | 3 | Promote scheduled → pending |
| `test-api` | 1 | Load-test helper — fires random notifications directly to the API |

### RabbitMQ Topology

Each channel has three exchanges following the same naming pattern:

```
Insider.One.Notification.Create.Email      (fanout)
Insider.One.Notification.Email.Created     (fanout)
Insider.One.Notification.Cancel.Email      (fanout)
```

Queues are bound per priority: `Insider.One.Notification.Create.Email.HIGH`, `.MEDIUM`, `.LOW`

---

## Prerequisites

- Go 1.25+
- PostgreSQL
- RabbitMQ
- Redis (distributed lock for scheduled jobs)

---

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/your-org/insider-one.git
cd insider-one
```

### 2. Configure environment files

Each component reads its config from `projects/{component-name}/{env}.env`. Create a file for each component you intend to run.

**Example: `projects/notification-management-api/qa.env`**

```env
APP.NAME=notification-management-api
APP.PORT=8080
APP.DEBUG=false

DB.HOST=localhost
DB.PORT=5432
DB.NAME=mydatabase
DB.USER=myuser
DB.PASSWORD=mypassword

RABBIT.HOST=localhost
RABBIT.PORT=5672
RABBIT.USER=admin
RABBIT.PASSWORD=admin123

EMAILPROVIDER.HOST=https://email-provider.example.com
EMAILPROVIDER.USER=apiuser
EMAILPROVIDER.PASSWORD=apipassword

SMSPROVIDER.HOST=https://sms-provider.example.com
SMSPROVIDER.USER=apiuser
SMSPROVIDER.PASSWORD=apipassword

PUSHPROVIDER.HOST=https://push-provider.example.com
PUSHPROVIDER.USER=apiuser
PUSHPROVIDER.PASSWORD=apipassword

REDIS.HOST=localhost
REDIS.PORT=6379
REDIS.DB=0
REDIS.USER=
REDIS.PASSWORD=
```

Consumer env files follow the same structure; only `APP.NAME` differs.

### 3. Build

```bash
go build -o insider-one .
```

### 4. Run

See [Running the Services](#running-the-services) below.

---

## Running the Services

All components share the same binary and are selected via a subcommand.

### Flags

| Flag | Values | Description |
|---|---|---|
| `--env` | `qa`, `prod` | Environment — selects the matching `.env` file |
| `--priority` | `HIGH`, `MEDIUM`, `LOW`, `*` | Priority queue to consume (`*` = all) |

> `--priority` is not applicable to the API or scheduled jobs.

### Commands

```bash
# REST API
./insider-one notification-management-api --env qa

# Create consumers (persist incoming notifications)
./insider-one create-email-consumer --env qa --priority HIGH
./insider-one create-push-consumer  --env qa --priority LOW
./insider-one create-sms-consumer   --env qa --priority MEDIUM

# Delivery consumers (call external providers)
./insider-one email-created-consumer --env qa --priority HIGH
./insider-one push-created-consumer  --env qa --priority MEDIUM
./insider-one sms-created-consumer   --env qa --priority LOW

# Cancellation consumers
./insider-one cancel-email-consumer --env qa
./insider-one cancel-push-consumer  --env qa
./insider-one cancel-sms-consumer   --env qa

# Scheduled jobs (promote SCHEDULED → PENDING)
./insider-one email-scheduled-job --env qa
./insider-one push-scheduled-job  --env qa
./insider-one sms-scheduled-job   --env qa

# Load-test helper (fires random notifications against the API)
./insider-one test-api --env qa
```

---

## API Reference

**Base URL:** `http://localhost:8080`
**Authentication:** All `/api/v1/*` routes require `Authorization: Bearer <token>`
**Interactive docs:** `GET /swagger/index.html`

### System Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check (DB + RabbitMQ status) |
| `GET` | `/metrics` | Prometheus metrics |
| `GET` | `/swagger/*` | Swagger UI |

---

### Email Endpoints

The SMS and Push channels expose identical endpoints under `/api/v1/sms/` and `/api/v1/push/` respectively, with the same request/response shapes.

---

#### POST `/api/v1/email/` — Send a single email

**Request**
```json
{
  "to": "recipient@example.com",
  "from": "sender@example.com",
  "subject": "Welcome",
  "content": "Hello, world!",
  "type": "marketing",
  "priority": "HIGH",
  "scheduled_at": "2026-05-01T10:00:00Z"
}
```

| Field | Type | Required | Constraints |
|---|---|---|---|
| `to` | string | yes | valid email, max 255 |
| `from` | string | yes | valid email, max 255 |
| `subject` | string | yes | 1–150 chars |
| `content` | string | yes | 1–10,000 chars |
| `type` | string | yes | 1–150 chars |
| `priority` | string | yes | `LOW` \| `MEDIUM` \| `HIGH` |
| `scheduled_at` | RFC3339 | no | must be in the future, within 1 year |

**Response:** `200 OK` (empty body)

---

#### POST `/api/v1/email/batch` — Send a batch of emails

**Request**
```json
{
  "emails": [
    {
      "to": "alice@example.com",
      "from": "sender@example.com",
      "subject": "Hello Alice",
      "content": "...",
      "type": "transactional",
      "priority": "MEDIUM"
    },
    {
      "to": "bob@example.com",
      "from": "sender@example.com",
      "subject": "Hello Bob",
      "content": "...",
      "type": "transactional",
      "priority": "LOW"
    }
  ]
}
```

- `emails`: array, 1–1,000 items; each item has the same fields as the single-send request.

**Response:** `200 OK` (empty body)

---

#### GET `/api/v1/email/` — List emails

**Query parameters**

| Parameter | Type | Required | Values / Constraints     |
|---|---|---|--------------------------|
| `status` | string | yes | `PENDING` \| `DELIVERED` |
| `page` | int | yes | min 1                    |
| `page_size` | int | yes | 0–50                     |
| `create_date` | RFC3339 | no | start of date range      |
| `end_date` | RFC3339 | no | end of date range        |

**Example:** `GET /api/v1/email/?status=PENDING&page=1&page_size=10`

**Response**
```json
{
  "emails": [
    {
      "id": 42,
      "to": "recipient@example.com",
      "from": "sender@example.com",
      "subject": "Welcome",
      "content": "Hello, world!",
      "status": "PENDING",
      "type": "marketing",
      "priority": "HIGH",
      "created_at": 1743465600,
      "scheduled_at": 0,
      "sent_at": 0
    }
  ],
  "total_count": 1,
  "page": 1,
  "page_size": 10
}
```

Timestamps are Unix epoch seconds. Zero value means the field is unset.

---

#### PUT `/api/v1/email/cancel` — Cancel emails by status

**Query parameters**

| Parameter | Type | Required | Values |
|---|---|---|---|
| `status` | string | yes | `PENDING` |

**Example:** `PUT /api/v1/email/cancel?status=PENDING`

**Response:** `200 OK` (empty body)

---

#### GET `/api/v1/email/status` — Get status by ID

**Query parameters**

| Parameter | Type | Required |
|---|---|---|
| `id` | uint64 | yes |

**Example:** `GET /api/v1/email/status?id=42`

**Response**
```json
{
  "email_id": 42,
  "status": "PENDING"
}
```

---

#### POST `/api/v1/email/status/batch` — Get statuses for multiple IDs

**Request**
```json
{
  "ids": [42, 43, 44]
}
```

**Response**
```json
{
  "emails": [
    { "email_id": 42, "status": "PENDING" },
    { "email_id": 43, "status": "DELIVERED" },
    { "email_id": 44, "status": "SCHEDULED" },
    { "email_id": 44, "status": "CANCELLED" }
  ]
}
```

---

### SMS Endpoints

Same as email, replacing `/api/v1/email/` with `/api/v1/sms/`. The `subject` field is absent from SMS requests; `to` and `from` accept phone numbers.

---

### Push Endpoints

Same as email, replacing `/api/v1/email/` with `/api/v1/push/`.

---

## Configuration Reference

| Key | Description |
|---|---|
| `APP.NAME` | Service name (used in logs) |
| `APP.PORT` | HTTP port (API only) |
| `APP.DEBUG` | Enable debug logging (`true`/`false`) |
| `DB.HOST` | PostgreSQL host |
| `DB.PORT` | PostgreSQL port |
| `DB.NAME` | Database name |
| `DB.USER` | Database user |
| `DB.PASSWORD` | Database password |
| `RABBIT.HOST` | RabbitMQ host |
| `RABBIT.PORT` | RabbitMQ AMQP port |
| `RABBIT.USER` | RabbitMQ user |
| `RABBIT.PASSWORD` | RabbitMQ password |
| `EMAILPROVIDER.HOST` | Email provider base URL |
| `EMAILPROVIDER.USER` | Email provider credentials |
| `EMAILPROVIDER.PASSWORD` | Email provider credentials |
| `SMSPROVIDER.HOST` | SMS provider base URL |
| `SMSPROVIDER.USER` | SMS provider credentials |
| `SMSPROVIDER.PASSWORD` | SMS provider credentials |
| `PUSHPROVIDER.HOST` | Push provider base URL |
| `PUSHPROVIDER.USER` | Push provider credentials |
| `PUSHPROVIDER.PASSWORD` | Push provider credentials |
| `REDIS.HOST` | Redis host (scheduled jobs only) |
| `REDIS.PORT` | Redis port |
| `REDIS.DB` | Redis database index |
| `REDIS.USER` | Redis user (optional) |
| `REDIS.PASSWORD` | Redis password (optional) |


## Run test with single command
```bash
go test ./...              # tüm testler
go test -v ./...           # verbose (her test adı görünür)
go test -cover ./...       # coverage yüzdesi
go test -race ./...        # race condition tespiti
```