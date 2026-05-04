# ── Stage 1: Build React SPA ──
FROM node:22-alpine AS frontend
WORKDIR /src
COPY web/package.json web/package-lock.json ./
RUN npm install --no-audit --no-fund
COPY web/ ./
RUN npm run build

# ── Stage 2: Build Go binary ──
FROM golang:1.25-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /parq-server .

# ── Stage 3: Runtime ──
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=frontend /src/dist /web/dist
COPY --from=backend /parq-server /parq-server
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/api/meta || exit 1
ENTRYPOINT ["/parq-server", "-static-dir=/web/dist", "-auto-discover"]
