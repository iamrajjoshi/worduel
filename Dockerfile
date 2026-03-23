# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

# Stage 2: Build backend
FROM golang:1.21.5-alpine AS backend-builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download && go mod verify
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o /app/worduel \
    ./main.go

# Stage 3: Production runtime (minimal)
FROM scratch AS production
COPY --from=backend-builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend-builder /etc/passwd /etc/passwd
COPY --from=backend-builder /app/worduel /worduel
COPY --from=frontend-builder /app/frontend/dist /frontend/dist
USER nobody
EXPOSE 8080
ENTRYPOINT ["/worduel"]

# Stage 4: Debug runtime (has shell + curl)
FROM alpine:3.19 AS debug
RUN apk --no-cache add ca-certificates tzdata curl
RUN addgroup -g 1001 worduel && adduser -u 1001 -G worduel -s /bin/sh -D worduel
WORKDIR /app
COPY --from=backend-builder /app/worduel /app/worduel
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist
RUN chown -R worduel:worduel /app
USER worduel
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:${PORT:-8080}/health || exit 1
ENTRYPOINT ["/app/worduel"]
