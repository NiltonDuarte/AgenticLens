# Multi-stage Docker build for AgenticLens standalone single-binary distribution

# Stage 1: Build the TypeScript frontend bundle
FROM node:20-alpine AS frontend-builder
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install
COPY web/ ./
RUN npm run build

# Stage 2: Build the Go backend binary with embedded assets
FROM golang:1.26-alpine AS backend-builder
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-builder /web/dist ./internal/adapters/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/agentic-lens ./cmd/agentic-lens

# Stage 3: Minimal runtime image with zero Node.js and zero Python dependencies
FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=backend-builder /bin/agentic-lens /bin/agentic-lens

EXPOSE 8085

ENTRYPOINT ["/bin/agentic-lens", "server", "--port", "8085", "--workspaces", "/workspaces"]
