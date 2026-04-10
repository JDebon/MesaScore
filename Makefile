DEV_COMPOSE = docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile dev

.PHONY: dev deploy logs down rebuild ps

## Start all services for local development
dev:
	$(DEV_COMPOSE) up -d
	@echo ""
	@echo "  App     → http://localhost"
	@echo "  Mailpit → http://localhost:8025"
	@echo "  DB      → localhost:5432 (user: mesascore / pass: changeme)"
	@echo ""

## Rebuild a service and restart it  (usage: make rebuild s=backend)
rebuild:
	$(DEV_COMPOSE) build $(s)
	$(DEV_COMPOSE) up -d --no-deps $(s)

## Build images and deploy to production (uses Cloudflare tunnel)
deploy:
	docker compose --profile prod up -d --build

## Tail logs, optionally filter by service  (usage: make logs s=backend)
logs:
	$(DEV_COMPOSE) logs -f $(s)

## Show running containers
ps:
	$(DEV_COMPOSE) ps

## Stop and remove all containers
down:
	docker compose --profile dev --profile prod down
