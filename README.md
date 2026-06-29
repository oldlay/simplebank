# Simple Bank

A full-stack banking web application built with Go, gRPC, and Vue 3. Provides user management, account operations, money transfers, and email verification — all served through a gRPC API with a RESTful HTTP gateway.

## Architecture Overview

```text
Client (Vue 3 + PrimeVue)
        │
        ▼
┌───────────────────────────────┐
│  HTTP Gateway (gRPC-Gateway)  │  :8080
│  Swagger UI                   │
└───────────┬───────────────────┘
            │
            ▼
┌───────────────────────────────┐
│       gRPC Server             │  :9090
│  (auth, accounts, transfers)  │
└──────┬────────────┬───────────┘
       │            │
       ▼            ▼
┌──────────┐  ┌──────────┐
│PostgreSQL│  │  Redis   │
│          │  │ (Asynq)  │
└──────────┘  └──────────┘
```

## Tech Stack

### Backend

- **Language:** Go 1.26
- **API:** gRPC + gRPC-Gateway (RESTful HTTP)
- **Database:** PostgreSQL 18 with [pgx](https://github.com/jackc/pgx) driver
- **SQL Generation:** [sqlc](https://sqlc.dev/) — type-safe Go code from SQL
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate)
- **Authentication:** JWT (HMAC-SHA256) & PASETO tokens
- **Async Tasks:** [Asynq](https://github.com/hibiken/asynq) (Redis-backed task queue)
- **Email:** Google Gmail SMTP via [jordan-wright/email](https://github.com/jordan-wright/email)
- **Configuration:** [Viper](https://github.com/spf13/viper)
- **Logging:** [zerolog](https://github.com/rs/zerolog)
- **API Docs:** Protocol Buffers → OpenAPI v2 (Swagger)

### Frontend

- **Framework:** Vue 3 (Composition API) + TypeScript
- **Build:** Vite 8
- **UI Library:** PrimeVue 4 + PrimeFlex
- **HTTP Client:** Axios
- **Routing:** Vue Router 5

## Features

- **User Management** — Sign up, login, update profile, role-based authorization
- **Account Management** — Create, list, get, update, and delete bank accounts
- **Money Transfers** — Transfer funds between accounts with transactional integrity
- **Email Verification** — Async email verification via Redis task queue
- **Exchange Rates** — Real-time currency exchange rate lookup
- **JWT & PASETO Auth** — Dual token system with access + refresh tokens
- **Swagger UI** — Auto-generated API documentation served at `/swagger/`

## Project Structure

```text
.
├── main.go                  # Application entry point
├── api/                     # Gin HTTP handlers (legacy, optional)
├── gapi/                    # gRPC server & RPC implementations
├── db/
│   ├── migration/           # Database migration files (up/down SQL)
│   ├── query/               # Raw SQL queries for sqlc
│   └── sqlc/                # Generated type-safe Go code from SQL
├── proto/                   # Protocol Buffer definitions
├── pb/                      # Generated protobuf Go code
├── token/                   # JWT & PASETO token makers
├── util/                    # Config, password hashing, random generators
├── worker/                  # Asynq task processor & distributor
├── mail/                    # Email sender (Gmail)
├── doc/
│   ├── swagger/             # Swagger UI & OpenAPI spec
│   ├── statik/              # Embedded static files
│   ├── db.dbml              # DBML database documentation
│   └── schema.sql           # Generated SQL schema
├── frontend/                # Vue 3 frontend application
├── Dockerfile               # Multi-stage Docker build
├── docker-compose.yaml      # PostgreSQL + Redis + API services
├── Makefile                 # Dev workflow shortcuts
└── app.env                  # Environment configuration
```

## Getting Started

### Prerequisites

- Go 1.26+
- Node.js 22+ (for frontend)
- PostgreSQL 18+
- Redis 8+
- [migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [sqlc](https://sqlc.dev/)
- [protoc](https://grpc.io/docs/protoc-installation/) + plugins

### Quick Start with Docker

```bash
docker compose up
```

This starts PostgreSQL, Redis, and the API server. The API will be available at:

- HTTP Gateway: `http://localhost:8080`
- gRPC Server: `localhost:9090`

### Manual Setup

1. **Start infrastructure**

   ```bash
   make postgres    # Start PostgreSQL container
   make redis       # Start Redis container
   make createdb    # Create the database
   ```

1. **Run database migrations**

   ```bash
   make migrateup
   ```

1. **Start the server**

   ```bash
   make server
   ```

1. **Start the frontend** (in a separate terminal)

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## Configuration

Configuration is loaded from `app.env` via Viper. Key settings:

| Variable | Description | Default |
|---|---|---|
| `ENVIRONMENT` | `development` or `production` | `development` |
| `DB_SOURCE` | PostgreSQL connection string | `postgresql://root:...@localhost:5432/simple_bank` |
| `MIGRATION_URL` | Migration files location | `file://db/migration` |
| `HTTP_SERVER_ADDRESS` | HTTP gateway listen address | `0.0.0.0:8080` |
| `GRPC_SERVER_ADDRESS` | gRPC server listen address | `0.0.0.0:9090` |
| `TOKEN_SYMMETRIC_KEY` | Symmetric key for PASETO/JWT | (32-char hex string) |
| `ACCESS_TOKEN_DURATION` | Access token TTL | `15m` |
| `REDIS_ADDRESS` | Redis server address | `0.0.0.0:6379` |
| `EXCHANGE_API` | Exchange rate API endpoint | exchangerate-api.com |

## Makefile Targets

| Command | Description |
|---|---|
| `make postgres` | Start PostgreSQL via Docker |
| `make createdb` | Create the application database |
| `make migrateup` | Run all pending migrations |
| `make migratedown` | Rollback all migrations |
| `make sqlc` | Generate Go code from SQL queries |
| `make proto` | Generate protobuf & gRPC code |
| `make test` | Run all tests with coverage |
| `make server` | Start the application |
| `make mock` | Generate mock implementations for testing |
| `make redis` | Start Redis via Docker |

## API Documentation

Swagger UI is available at `http://localhost:8080/swagger/` when the server is running.

### gRPC Services

Defined in [proto/service_simple_bank.proto](proto/service_simple_bank.proto):

- `CreateUser` / `LoginUser` / `UpdateUser` / `VerifyEmail`
- `CreateAccount` / `GetAccount` / `ListAccount` / `UpdateAccount` / `DeleteAccount`
- `CreateTransfer`
- `GetExchangeRate`

## Database

### Schema

- **accounts** — Bank accounts with owner, currency, and balance
- **users** — Application users with hashed passwords and roles
- **transfers** — Money transfer records between accounts
- **entries** — Double-entry bookkeeping ledger
- **sessions** — User session records for refresh tokens
- **verify_emails** — Email verification codes and status

See [doc/db.dbml](doc/db.dbml) for the DBML diagram and [doc/schema.sql](doc/schema.sql) for the full SQL schema.

### Creating New Migrations

```bash
make new_migration name=your_migration_name
```

## License

[MIT](LICENSE)
