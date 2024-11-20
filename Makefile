# Makefile for grpc-todo

PROTO_DIR=proto
SERVER_DIR=server
MAIN_FILE=main.go
BIN_DIR=bin

GOLANGCI_LINT=golangci-lint
GRPC_URL=grpcurl
GO=go

GOLANGCI_LINT_VERSION=v1.62.0


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
	@PID=$$(lsof -ti tcp:50051) && \
	if [ -n "$$PID" ]; then \
		kill $$PID && echo "Server stopped."; \
	else \
		echo "No server is running on port 50051."; \
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

# Примеры команд grpcurl

# Создание задачи
.PHONY: create-task
create-task:
	$(GRPC_URL) -plaintext -d '{"title": "Sample Task", "description": "This is a sample task"}' localhost:50051 todo.ToDoService/CreateTask

# Получение всех задач
.PHONY: get-all-tasks
get-all-tasks:
	$(GRPC_URL) -plaintext -d '{}' localhost:50051 todo.ToDoService/GetAllTasks

# Обновление статуса задачи
.PHONY: update-task-status
update-task-status:
	$(GRPC_URL) -plaintext -d '{"id": 1, "status": "IN_PROGRESS"}' localhost:50051 todo.ToDoService/UpdateTaskStatus

# Удаление задачи
.PHONY: delete-task
delete-task:
	$(GRPC_URL) -plaintext -d '{"id": 1}' localhost:50051 todo.ToDoService/DeleteTask

# Проверка цикломатической сложности
.PHONY: cyclo
cyclo:
	$(GOLANGCI_LINT) run --enable=gocyclo



# Просмотр всех доступных команд
.PHONY: help
help:
	@echo "Available make commands:"
	@echo "  make deps             Install and tidy dependencies"
	@echo "  make generate         Generate Go code from .proto files"
	@echo "  make build            Build the project"
	@echo "  make lint             Run linters"
	@echo "  make test             Run tests"
	@echo "  make run              Run the server"
	@echo "  make clean            Clean build artifacts and generated files"
	@echo "  make install-linter   Install golangci-lint if not installed"
	@echo "  make create-task      Example: Create a new task using grpcurl"
	@echo "  make get-all-tasks    Example: Get all tasks using grpcurl"
	@echo "  make update-task-status Example: Update task status using grpcurl"
	@echo "  make delete-task      Example: Delete a task using grpcurl"
	@echo "  make cyclo            Check cyclomatic complexity"
	@echo "  make help             Show this help message"