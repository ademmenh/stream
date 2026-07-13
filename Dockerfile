FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git curl ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

RUN curl -sSfL https://release.ariga.io/atlas/atlas-linux-amd64-latest -o /app/atlas && \
    chmod +x /app/atlas

FROM alpine:3.20

RUN apk add --no-cache ca-certificates ffmpeg

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/atlas .
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8000

CMD ["./server"]
