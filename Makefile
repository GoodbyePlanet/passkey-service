APP_NAME = passkey-service
DOCKER_COMPOSE = docker compose
GO = go

# Default target
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make build         Build the Go binary"
	@echo "  make run           Run the app locally"
	@echo "  make docker-build  Build the Docker image"
	@echo "  make up            Start all Docker services"
	@echo "  make db-up         Start only PostgreSQL"
	@echo "  make down          Stop and remove containers"
	@echo "  make logs          Tail API logs"
	@echo "  make db            Open a Postgres shell"
	@echo "  make clean         Remove build artifacts and Docker volumes"

# -------------------------------
# Local development targets
# -------------------------------
build:
	$(GO) build -o $(APP_NAME) . && chmod +x $(APP_NAME)

run: build
	./$(APP_NAME)

# -------------------------------
# Docker targets
# -------------------------------
docker-build:
	$(DOCKER_COMPOSE) build --no-cache

up:
	$(DOCKER_COMPOSE) up

db-up:
	$(DOCKER_COMPOSE) up -d db

down:
	$(DOCKER_COMPOSE) down

logs:
	$(DOCKER_COMPOSE) logs -f passkeys-service

db:
	docker exec -it passkey_db psql -U postgres -d passkey_db

clean:
	rm -f $(APP_NAME)
	$(DOCKER_COMPOSE) down -v

