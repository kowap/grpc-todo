PROTO_DIR=proto
SERVER_DIR=server
MAIN_FILE=main.go
BIN_DIR=bin

GOLANGCI_LINT=golangci-lint
GRPC_URL=grpcurl
GO=go

GOLANGCI_LINT_VERSION=v1.62.0

PORT=50051

DOCKER_COMPOSE=docker-compose
DOCKER=docker

.PHONY: all
all: build

.PHONY: deps
deps:
	$(GO) mod tidy

.PHONY: generate
generate:
	protoc \
	  --go_out=paths=source_relative:. \
	  --go-grpc_out=paths=source_relative:. \
	  $(PROTO_DIR)/todo.proto

.PHONY: build
build: generate
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/grpc-todo $(MAIN_FILE)

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run

.PHONY: test
test:
	$(GO) test ./...

.PHONY: run
run: build
	./$(BIN_DIR)/grpc-todo

.PHONY: stop
stop:
	@echo "Stopping grpc-todo server..."
	@PID=$$(lsof -ti tcp:$(PORT)) && \
	if [ -n "$$PID" ]; then \
		kill $$PID && echo "Server stopped."; \
	else \
		echo "No server is running on port $(PORT)."; \
	fi

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	rm -f $(PROTO_DIR)/*.pb.go

.PHONY: install-linter
install-linter:
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || { \
		echo "Installing $(GOLANGCI_LINT)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	}

# Docker команды

# Запуск приложения с помощью Docker Compose
.PHONY: docker-up
docker-up:
	$(DOCKER_COMPOSE) up -d

# Остановка и удаление контейнеров
.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down

# Пересборка образов и перезапуск контейнеров
.PHONY: docker-rebuild
docker-rebuild:
	$(DOCKER_COMPOSE) up -d --build

# Просмотр логов приложения
.PHONY: logs
logs:
	$(DOCKER_COMPOSE) logs -f app

# Примеры команд grpcurl

# Создание задачи
.PHONY: create-task
create-task:
	$(GRPC_URL) -plaintext -d '{"title": "Sample Task", "description": "This is a sample task"}' localhost:$(PORT) todo.ToDoService/CreateTask

# Получение всех задач
.PHONY: get-all-tasks
get-all-tasks:
	$(GRPC_URL) -plaintext -d '{}' localhost:$(PORT) todo.ToDoService/GetAllTasks

# Обновление статуса задачи (требуется указать ID)
.PHONY: update-task-status
update-task-status:
	@if [ -z "$(ID)" ]; then echo "Please set ID variable: make update-task-status ID=<task_id>"; exit 1; fi
	$(GRPC_URL) -plaintext -d '{"id": "$(ID)", "status": "DONE"}' localhost:$(PORT) todo.ToDoService/UpdateTaskStatus

# Удаление задачи (требуется указать ID)
.PHONY: delete-task
delete-task:
	@if [ -z "$(ID)" ]; then echo "Please set ID variable: make delete-task ID=<task_id>"; exit 1; fi
	$(GRPC_URL) -plaintext -d '{"id": "$(ID)"}' localhost:$(PORT) todo.ToDoService/DeleteTask

# Проверка цикломатической сложности
.PHONY: cyclo
cyclo:
	$(GOLANGCI_LINT) run --enable=gocyclo

# Просмотр всех доступных команд
.PHONY: help
help:
	@echo "Available make commands:"
	@echo "  make deps               Install and tidy dependencies"
	@echo "  make generate           Generate Go code from .proto files"
	@echo "  make build              Build the project"
	@echo "  make lint               Run linters"
	@echo "  make test               Run tests"
	@echo "  make run                Run the server locally"
	@echo "  make stop               Stop the locally running server"
	@echo "  make docker-up          Start the application using Docker Compose"
	@echo "  make docker-down        Stop and remove Docker containers"
	@echo "  make docker-rebuild     Rebuild images and restart containers"
	@echo "  make logs               View application logs"
	@echo "  make clean              Clean build artifacts and generated files"
	@echo "  make install-linter     Install golangci-lint if not installed"
	@echo "  make create-task        Example: Create a new task using grpcurl"
	@echo "  make get-all-tasks      Example: Get all tasks using grpcurl"
	@echo "  make update-task-status ID=<task_id>  Update task status using grpcurl"
	@echo "  make delete-task        ID=<task_id>  Delete a task using grpcurl"
	@echo "  make cyclo              Check cyclomatic complexity"
	@echo "  make help               Show this help message"