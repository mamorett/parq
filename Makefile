# Build targets
.PHONY: build-backend build-frontend build docker dev lint test

# Go settings
GO_BIN = parq-server

# Frontend settings
WEB_DIR = web
WEB_DIST = $(WEB_DIR)/dist

build: build-frontend build-backend

build-backend:
	go build -ldflags="-s -w" -o $(GO_BIN) .

build-frontend:
	cd $(WEB_DIR) && npm run build

dev:
	# Run Go backend in one terminal and Vite dev server in another
	# For simplicity in this Makefile, we just point to the instructions
	@echo "Run 'go run main.go' for backend and 'cd web && npm run dev' for frontend"

docker:
	docker build -t parq .

lint:
	go vet ./...
	cd $(WEB_DIR) && npm run lint

test:
	go test -v ./...
	cd $(WEB_DIR) && npm test
