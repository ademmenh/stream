# Implementation Plan

## Step 1: Domain & Application Layer

Completed. See `internal/videos/domain/` and `internal/videos/application/`.

---

## Step 2: Infrastructure Layer

### Database

- Ent schema definition for `videos` table in `internal/shared/infrastructure/ent/schema/video.go`
- Fields: id (UUID), title, description, type (enum), status (enum), qualities (JSON array), uploaded_at
- Run `make generate` to regenerate Ent code
- Atlas migration to create the `videos` table
- `VideoRepository` in `internal/videos/infrastructure/repository.go` implementing `domain.IVideoRepository`

### Storage (MinIO / S3)

- `VideoStorageAdapter` in `internal/videos/infrastructure/storage_adapter.go` implementing `domain.IStorageAdapter`
- Wraps the existing `shared/infrastructure/S3Adapter`
- Methods: `GeneratePresignedUploadUrl`, `GeneratePresignedGetUrl`, `ObjectExists`, `DeletePrefixes`

### Message Queue

- `MessageQueueAdapter` in `internal/videos/infrastructure/queue_adapter.go` implementing `domain.IMessageQueue`
- Simple in-memory or Redis-backed queue adapter

### Dependency Wiring

- Update `internal/app/app.go` to create `VideoRepository`, `VideoStorageAdapter`, and `MessageQueueAdapter`
- Pass them to `videos.NewModule(deps)`

---

## Step 3: Presentation Layer — Admin Endpoints

- `internal/videos/presentation/handlers.go` — Admin handlers (CreateVideo, TriggerProcessing, ReplaceVideo, RegenerateQuality, DeleteVideo, ListVideos)
- `internal/videos/presentation/dtos.go` — Request DTOs
- `internal/videos/presentation/error_handler.go` — Domain errors → HTTP status codes
- Register `VideoErrorHandler` in `shared/presentation/error_handler.go`
- Wire admin routes in `internal/videos/module.go` via `RegisterRoutes`

---

## Step 4: Presentation Layer — Worker (Video Processing)

Long-running process consuming jobs from queue. Downloads raw, runs ffmpeg HLS transcode, uploads chunks, updates DB status.

---

## Step 5: Presentation Layer — User Endpoints

- ListCatalog handler (GET /api/videos)
- GetVideoStream handler (GET /api/videos/:id)
- Register routes in module.go
