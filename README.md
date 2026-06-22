# Go Starter API

A starter template for a Go/Echo backend with Domain-Driven Design (DDD) architecture.

## Stack

- **Language**: Go 1.23+
- **Framework**: Echo v4
- **Database**: PostgreSQL 16
- **Query Builder**: database/sql + lib/pq
- **Auth**: JWT (golang-jwt) + bcrypt (golang.org/x/crypto)
- **Reverse Proxy**: NGINX

## Project Structure

```
cmd/
  server/              # App entry point
internal/
  config/
    domain/            # IConfig interface
    infrastructure/    # ConfigAdapter (env vars)
  shared/
    domain/            # Id, Email, Phone value objects + Pagination
    infrastructure/    # IDGenerator, Logger
    presentation/      # Auth middleware, error handlers, response types
  auth/
    domain/            # I/JwtAdapter, IPasswordAdapter interfaces
    application/       # Login, Register, RefreshToken use cases
    infrastructure/    # JWT adapter, password adapter
    presentation/      # HTTP handlers, DTOs
  users/
    domain/            # UserEntity, errors, ports
    application/       # GetUser, ListUsers, UpdateUser, DeleteUser
    infrastructure/    # PostgreSQL repository, mapper
    presentation/      # HTTP handlers, DTOs
  app/                 # Dependency injection + Echo setup
migrations/            # SQL migration files
```

## DDD Layer Convention (per module)

- `domain/` — entities, value objects, port interfaces, domain errors
- `application/` — use cases (one file per use case)
- `infrastructure/` — db repositories, mappers, external adapters
- `presentation/` — Echo handlers, DTOs (request)

## API Endpoints

| Method | Path | Access |
|--------|------|--------|
| GET | /api/v1/health | Public |
| POST | /api/v1/auth/login | Public |
| POST | /api/v1/auth/register | Public |
| POST | /api/v1/auth/refresh | Public (reads `refresh_token` cookie) |
| GET | /api/v1/users/@me | Authenticated |
| PUT | /api/v1/users/@me | Authenticated (no role change) |
| DELETE | /api/v1/users/@me | Authenticated |
| GET | /api/v1/users | Admin |
| GET | /api/v1/users/:id | Admin |
| PUT | /api/v1/users/:id | Admin (can change role) |
| DELETE | /api/v1/users/:id | Admin |
| POST | /api/v1/users/:id/ban | Admin |
| POST | /api/v1/users/:id/unban | Admin |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENV` | `dev` | Environment (dev/prod) |
| `PORT` | `8000` | Server port |
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `waslini` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `JWT_ACCESS_TOKEN_SECRET` | — | Access token signing secret |
| `JWT_REFRESH_TOKEN_SECRET` | — | Refresh token signing secret |
| `JWT_ACCESS_TOKEN_EXPIRY` | `3600` | Access token TTL (seconds) |
| `JWT_REFRESH_TOKEN_EXPIRY` | `604800` | Refresh token TTL (seconds) |
| `CORS_ORIGINS` | `["*"]` | Allowed CORS origins |
| `COOKIES_SECURE` | `true` | Secure cookie flag |
| `COOKIES_SAME_SITE` | `lax` | SameSite cookie policy |

## Running

### Local Development

```bash
make deps
make dev
```

### With Docker

```bash
cp .env.example .env
make build:dev
make start:dev
```

### Production

```bash
make build:prod
# docker compose up -d
```

## License

This project is licensed under the GPL v3 License.
