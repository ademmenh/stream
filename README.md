# STREAM API

```bash
An application that streams videos with ffmpeg hls and minio s3.
```

## Tech Stack

- **Language**: Go 1.25+
- **Framework**: Echo v4
- **Database**: PostgreSQL 16 + Ent ORM
- **Migrations**: Atlas
- **Storage**: MinIO (S3-compatible)
- **Auth**: JWT (golang-jwt) + bcrypt (golang.org/x/crypto)
- **Video Processing**: FFmpeg (fluent-ffmpeg) + HLS
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
    infrastructure/    # Ent schemas, generated client, DB connection
    presentation/      # Auth middleware, error handlers, response types
    tests/             # Shared test utilities (LoadDotEnv, BuildDSN)
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
  videos/
    domain/            # Video entity, VideoStatus/VideoType/VideoQuality, ports, errors
    application/       # (pending)
    infrastructure/    # PostgreSQL repository, mapper, InMemoryVideoRepository
    presentation/      # (pending)
  app/                 # Dependency injection + Echo setup
migrations/            # Atlas SQL migration files
```

## DDD Layer Convention (per module)

- `domain/` — entities, value objects, port interfaces, domain errors
- `application/` — use cases (one file per use case)
- `infrastructure/` — db repositories, mappers, external adapters
- `presentation/` — Echo handlers, DTOs (request)

## API Endpoints

| Method | Path                    | Access                                |
| ------ | ----------------------- | ------------------------------------- |
| GET    | /api/v1/health          | Public                                |
| POST   | /api/v1/auth/login      | Public                                |
| POST   | /api/v1/auth/register   | Public                                |
| POST   | /api/v1/auth/refresh    | Public (reads `refresh_token` cookie) |
| GET    | /api/v1/users/@me       | Authenticated                         |
| PUT    | /api/v1/users/@me       | Authenticated (no role change)        |
| DELETE | /api/v1/users/@me       | Authenticated                         |
| GET    | /api/v1/users           | Admin                                 |
| GET    | /api/v1/users/:id       | Admin                                 |
| PUT    | /api/v1/users/:id       | Admin (can change role)               |
| DELETE | /api/v1/users/:id       | Admin                                 |
| POST   | /api/v1/users/:id/ban   | Admin                                 |
| POST   | /api/v1/users/:id/unban | Admin                                 |

### Video Endpoints (planned)

| Method | Path                             | Access | Description                             |
| ------ | -------------------------------- | ------ | --------------------------------------- |
| POST   | /api/admin/videos                | Admin  | Create metadata + presigned upload URLs |
| POST   | /api/admin/videos/:id/process    | Admin  | Trigger HLS transcoding worker          |
| PUT    | /api/admin/videos/:id/replace    | Admin  | Replace video file, keep metadata       |
| POST   | /api/admin/videos/:id/regenerate | Admin  | Add encoding quality (e.g. 720p)        |
| DELETE | /api/admin/videos/:id            | Admin  | Delete video and all S3 assets          |
| GET    | /api/admin/videos                | Admin  | List with status/type filters           |
| GET    | /api/videos                      | Public | Catalog of Ready videos                 |
| GET    | /api/videos/:id                  | Public | Stream HLS master playlist              |

## Environment Variables

| Variable                    | Description                 |
| -------------------------- | ---------------------------- |
| `ENV`                      | Environment (dev/prod)       |
| `APP_NAME`                 | Application name             |
| `PORT`                     | Server port                  |
| `DB_HOST`                  | Database host                |
| `DB_PORT`                  | Database port                |
| `DB_USER`                  | Database user                |
| `DB_PASSWORD`              | Database password            |
| `DB_NAME`                  | Database name                |
| `DB_SSLMODE`               | SSL mode                     |
| `JWT_ACCESS_TOKEN_SECRET`  | Access token signing secret  |
| `JWT_REFRESH_TOKEN_SECRET` | Refresh token signing secret |
| `JWT_ACCESS_TOKEN_EXPIRY`  | Access token TTL (seconds)   |
| `JWT_REFRESH_TOKEN_EXPIRY` | Refresh token TTL (seconds)  |
| `CORS_ORIGINS`             | Allowed CORS origins         |
| `COOKIES_SECURE`           | Secure cookie flag           |
| `COOKIES_SAME_SITE`        | SameSite cookie policy       |
| `S3_HOST`                  | MinIO/S3 host                |
| `S3_PORT`                  | MinIO/S3 API port            |
| `S3_REGION`                | S3 region                    |
| `S3_ACCESS_KEY`            | S3 access key                |
| `S3_SECRET_KEY`            | S3 secret key                |
| `S3_BUCKET`                | S3 bucket name               |
| `S3_PUBLIC_ENDPOINT`       | Public S3 endpoint           |

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

### Migrations

```bash
# Create a new migration from schema changes
make migrate:create

# Apply pending migrations
make migrate:apply
```

### Code Generation

```bash
# Regenerate Ent code from schema files
make generate
```

## Testing

```bash
make test               # All tests
make test:unit          # Unit tests (no DB required)
make test:integration   # Integration tests (requires Postgres)
make test:e2e           # End-to-end tests
```

Set up `.env.test` for integration test DB connection:

## License

This project is licensed under the GPL v3 License.
