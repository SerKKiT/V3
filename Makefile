# Makefile - Database migration management
# Usage: make <target>

# =================================================================
# CONFIGURATION
# =================================================================
include .env
export

# Database URLs
AUTH_DB_URL := postgresql://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(AUTH_DB_NAME)?sslmode=disable
STREAMS_DB_URL := postgresql://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(STREAMS_DB_NAME)?sslmode=disable
VOD_DB_URL := postgresql://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(VOD_DB_NAME)?sslmode=disable

# Migrations path
MIGRATIONS_PATH := infrastructure/postgres/migrations

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# =================================================================
# HELP
# =================================================================
.PHONY: help
help: ## Show this help message
	@echo "$(GREEN)Available targets:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-25s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make migrate-up              # Apply all pending migrations"
	@echo "  make migrate-down DB=streams # Rollback last migration in streams_db"
	@echo "  make migrate-create NAME=add_followers DB=auth"
	@echo ""

.DEFAULT_GOAL := help

# =================================================================
# DOCKER COMPOSE COMMANDS
# =================================================================
.PHONY: up
up: ## Start all services (with migrations)
	@echo "$(GREEN)🚀 Starting all services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✅ All services started$(NC)"

.PHONY: down
down: ## Stop all services
	@echo "$(YELLOW)⏹️  Stopping all services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✅ All services stopped$(NC)"

.PHONY: restart
restart: down up ## Restart all services

.PHONY: logs
logs: ## Show logs (usage: make logs [SERVICE=service-name])
	@if [ -z "$(SERVICE)" ]; then \
		docker-compose logs -f; \
	else \
		docker-compose logs -f $(SERVICE); \
	fi

.PHONY: logs-migrations
logs-migrations: ## Show migration logs
	docker-compose logs migrations

# =================================================================
# MIGRATION COMMANDS
# =================================================================
.PHONY: migrate-up
migrate-up: ## Apply all pending migrations to all databases
	@echo "$(GREEN)🚀 Applying migrations to all databases...$(NC)"
	@echo ""
	@echo "$(YELLOW)📦 Migrating auth_db...$(NC)"
	@migrate -path $(MIGRATIONS_PATH)/auth_db -database "$(AUTH_DB_URL)" up || true
	@echo ""
	@echo "$(YELLOW)📦 Migrating streams_db...$(NC)"
	@migrate -path $(MIGRATIONS_PATH)/streams_db -database "$(STREAMS_DB_URL)" up || true
	@echo ""
	@echo "$(YELLOW)📦 Migrating vod_db...$(NC)"
	@migrate -path $(MIGRATIONS_PATH)/vod_db -database "$(VOD_DB_URL)" up || true
	@echo ""
	@echo "$(GREEN)🎉 All migrations completed!$(NC)"

.PHONY: migrate-down
migrate-down: ## Rollback last migration (usage: make migrate-down DB=streams)
	@if [ -z "$(DB)" ]; then \
		echo "$(RED)❌ Error: DB parameter is required$(NC)"; \
		echo "$(YELLOW)Usage: make migrate-down DB=auth|streams|vod$(NC)"; \
		exit 1; \
	fi
	@if [ "$(DB)" = "auth" ]; then \
		echo "$(YELLOW)⏪ Rolling back auth_db...$(NC)"; \
		migrate -path $(MIGRATIONS_PATH)/auth_db -database "$(AUTH_DB_URL)" down 1; \
	elif [ "$(DB)" = "streams" ]; then \
		echo "$(YELLOW)⏪ Rolling back streams_db...$(NC)"; \
		migrate -path $(MIGRATIONS_PATH)/streams_db -database "$(STREAMS_DB_URL)" down 1; \
	elif [ "$(DB)" = "vod" ]; then \
		echo "$(YELLOW)⏪ Rolling back vod_db...$(NC)"; \
		migrate -path $(MIGRATIONS_PATH)/vod_db -database "$(VOD_DB_URL)" down 1; \
	else \
		echo "$(RED)❌ Error: Invalid DB parameter$(NC)"; \
		echo "$(YELLOW)Valid values: auth, streams, vod$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Rollback completed$(NC)"

.PHONY: migrate-create
migrate-create: ## Create new migration (usage: make migrate-create NAME=add_followers DB=auth)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)❌ Error: NAME parameter is required$(NC)"; \
		echo "$(YELLOW)Usage: make migrate-create NAME=add_followers DB=auth$(NC)"; \
		exit 1; \
	fi
	@if [ -z "$(DB)" ]; then \
		echo "$(RED)❌ Error: DB parameter is required$(NC)"; \
		echo "$(YELLOW)Usage: make migrate-create NAME=add_followers DB=auth|streams|vod$(NC)"; \
		exit 1; \
	fi
	@if [ "$(DB)" = "auth" ]; then \
		echo "$(GREEN)📝 Creating migration for auth_db: $(NAME)$(NC)"; \
		migrate create -ext sql -dir $(MIGRATIONS_PATH)/auth_db -seq $(NAME); \
	elif [ "$(DB)" = "streams" ]; then \
		echo "$(GREEN)📝 Creating migration for streams_db: $(NAME)$(NC)"; \
		migrate create -ext sql -dir $(MIGRATIONS_PATH)/streams_db -seq $(NAME); \
	elif [ "$(DB)" = "vod" ]; then \
		echo "$(GREEN)📝 Creating migration for vod_db: $(NAME)$(NC)"; \
		migrate create -ext sql -dir $(MIGRATIONS_PATH)/vod_db -seq $(NAME); \
	else \
		echo "$(RED)❌ Error: Invalid DB parameter$(NC)"; \
		exit 1; \
	fi

.PHONY: migrate-version
migrate-version: ## Show current migration version (usage: make migrate-version DB=streams)
	@if [ -z "$(DB)" ]; then \
		echo "$(YELLOW)Auth DB version:$(NC)"; \
		migrate -path $(MIGRATIONS_PATH)/auth_db -database "$(AUTH_DB_URL)" version || echo "No migrations applied"; \
		echo ""; \
		echo "$(YELLOW)Streams DB version:$(NC)"; \
		migrate -path $(MIGRATIONS_PATH)/streams_db -database "$(STREAMS_DB_URL)" version || echo "No migrations applied"; \
		echo ""; \
		echo "$(YELLOW)VOD DB version:$(NC)"; \
		migrate -path $(MIGRATIONS_PATH)/vod_db -database "$(VOD_DB_URL)" version || echo "No migrations applied"; \
	else \
		if [ "$(DB)" = "auth" ]; then \
			migrate -path $(MIGRATIONS_PATH)/auth_db -database "$(AUTH_DB_URL)" version; \
		elif [ "$(DB)" = "streams" ]; then \
			migrate -path $(MIGRATIONS_PATH)/streams_db -database "$(STREAMS_DB_URL)" version; \
		elif [ "$(DB)" = "vod" ]; then \
			migrate -path $(MIGRATIONS_PATH)/vod_db -database "$(VOD_DB_URL)" version; \
		fi \
	fi

.PHONY: migrate-force
migrate-force: ## Force migration version (usage: make migrate-force DB=streams VER=1)
	@if [ -z "$(DB)" ] || [ -z "$(VER)" ]; then \
		echo "$(RED)❌ Error: DB and VER parameters are required$(NC)"; \
		echo "$(YELLOW)Usage: make migrate-force DB=streams VER=1$(NC)"; \
		exit 1; \
	fi
	@echo "$(RED)⚠️  WARNING: This will force the migration version!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [ "$$REPLY" = "y" ] || [ "$$REPLY" = "Y" ]; then \
		if [ "$(DB)" = "auth" ]; then \
			migrate -path $(MIGRATIONS_PATH)/auth_db -database "$(AUTH_DB_URL)" force $(VER); \
		elif [ "$(DB)" = "streams" ]; then \
			migrate -path $(MIGRATIONS_PATH)/streams_db -database "$(STREAMS_DB_URL)" force $(VER); \
		elif [ "$(DB)" = "vod" ]; then \
			migrate -path $(MIGRATIONS_PATH)/vod_db -database "$(VOD_DB_URL)" force $(VER); \
		fi; \
		echo "$(GREEN)✅ Version forced to $(VER)$(NC)"; \
	else \
		echo "$(YELLOW)❌ Cancelled$(NC)"; \
	fi

# =================================================================
# DATABASE MANAGEMENT
# =================================================================
.PHONY: db-reset
db-reset: ## Reset database (WARNING: deletes all data!)
	@echo "$(RED)⚠️  WARNING: This will DELETE ALL DATA!$(NC)"
	@read -p "Are you sure? Type 'yes' to confirm: " -r; \
	echo; \
	if [ "$$REPLY" = "yes" ]; then \
		echo "$(YELLOW)🗑️  Stopping services and removing volumes...$(NC)"; \
		docker-compose down -v; \
		echo "$(GREEN)✅ Volumes removed$(NC)"; \
		echo ""; \
		echo "$(YELLOW)🚀 Starting PostgreSQL...$(NC)"; \
		docker-compose up -d postgres; \
		echo "$(YELLOW)⏳ Waiting for PostgreSQL to be ready...$(NC)"; \
		sleep 10; \
		echo "$(GREEN)✅ PostgreSQL ready$(NC)"; \
		echo ""; \
		echo "$(YELLOW)📦 Running migrations...$(NC)"; \
		$(MAKE) migrate-up; \
		echo ""; \
		echo "$(GREEN)🎉 Database reset completed!$(NC)"; \
	else \
		echo "$(YELLOW)❌ Cancelled$(NC)"; \
	fi

.PHONY: db-backup
db-backup: ## Backup all databases
	@echo "$(YELLOW)💾 Creating database backups...$(NC)"
	@mkdir -p backups
	@TIMESTAMP=$$(date +%Y%m%d_%H%M%S); \
	docker exec streaming-postgres pg_dump -U $(DB_USER) $(AUTH_DB_NAME) > backups/auth_db_$$TIMESTAMP.sql; \
	docker exec streaming-postgres pg_dump -U $(DB_USER) $(STREAMS_DB_NAME) > backups/streams_db_$$TIMESTAMP.sql; \
	docker exec streaming-postgres pg_dump -U $(DB_USER) $(VOD_DB_NAME) > backups/vod_db_$$TIMESTAMP.sql; \
	echo "$(GREEN)✅ Backups created in ./backups/$(NC)"

.PHONY: db-shell
db-shell: ## Open PostgreSQL shell (usage: make db-shell DB=streams)
	@if [ -z "$(DB)" ]; then \
		echo "$(YELLOW)Opening default PostgreSQL shell...$(NC)"; \
		docker exec -it streaming-postgres psql -U $(DB_USER) -d $(STREAMS_DB_NAME); \
	else \
		if [ "$(DB)" = "auth" ]; then \
			docker exec -it streaming-postgres psql -U $(DB_USER) -d $(AUTH_DB_NAME); \
		elif [ "$(DB)" = "streams" ]; then \
			docker exec -it streaming-postgres psql -U $(DB_USER) -d $(STREAMS_DB_NAME); \
		elif [ "$(DB)" = "vod" ]; then \
			docker exec -it streaming-postgres psql -U $(DB_USER) -d $(VOD_DB_NAME); \
		fi \
	fi

# =================================================================
# DEVELOPMENT
# =================================================================
.PHONY: dev
dev: ## Start development environment
	@echo "$(GREEN)🔧 Starting development environment...$(NC)"
	docker-compose up postgres minio migrations
	@echo "$(GREEN)✅ Development environment ready$(NC)"
	@echo "$(YELLOW)PostgreSQL:$(NC) localhost:5432"
	@echo "$(YELLOW)MinIO Console:$(NC) http://localhost:9001"

.PHONY: clean
clean: ## Clean up containers and volumes
	@echo "$(YELLOW)🧹 Cleaning up...$(NC)"
	docker-compose down -v --remove-orphans
	docker system prune -f
	@echo "$(GREEN)✅ Cleanup completed$(NC)"

.PHONY: status
status: ## Show status of all services
	@docker-compose ps
