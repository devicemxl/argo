# Makefile para Arena + CGO projeto

# Variables
CC = gcc
CFLAGS = -std=c99 -O2 -Wall -Wextra -fPIC
LDFLAGS = -shared -lm

# Archivos 
GO_MAIN_SOURCE = ./example/main.go
BINARY = arena_cgo_example
C_SOURCES = ./example/cgo_utils.c
C_HEADERS = ./example/cgo_utils.h

# Objetivos principales
.PHONY: all clean build run test benchmark install debug release check info help

all: build

# Compilar el proyecto
build:
	@echo "Building Arena + CGO project..."
	@cd example && go build -o ../$(BINARY) .
	@echo "Build completed: $(BINARY)"

# Ejecutar el programa
run: build
	@echo "Running Arena + CGO examples..."
	@./$(BINARY)

# Ejecutar tests en la librería principal
test:
	@echo "Running tests..."
	@go test -v .

# Ejecutar tests en el ejemplo
test-example:
	@echo "Running example tests..."
	@cd example && go test -v .

# Ejecutar benchmarks
benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem .
	@cd example && go test -bench=. -benchmem .

# Compilar biblioteca C por separado (opcional)
libcgo_utils.so: $(C_SOURCES) $(C_HEADERS)
	@echo "Building C shared library..."
	@$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $(C_SOURCES)

# Instalar dependencias
install:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Limpiar archivos generados
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY)
	@rm -f libcgo_utils.so
	@rm -f example/$(BINARY)
	@echo "Cleanup completed"

# Compilación para debug
debug: CFLAGS += -g -DDEBUG
debug: build
	@echo "Debug build completed"

# Compilación optimizada para producción
release: CFLAGS += -O3 -DNDEBUG
release: build
	@echo "Release build completed"

# Verificar que todos los archivos estén presentes
check:
	@echo "Checking project files..."
	@test -f example/cgo_utils.h || (echo "Missing: example/cgo_utils.h" && exit 1)
	@test -f example/cgo_utils.c || (echo "Missing: example/cgo_utils.c" && exit 1)
	@test -f argo.go || (echo "Missing: argo.go" && exit 1)
	@test -f example/main.go || (echo "Missing: example/main.go" && exit 1)
	@test -f example/argoFun.go || (echo "Missing: example/argoFun.go" && exit 1)
	@echo "All required files present"

# Información del sistema
info:
	@echo "System Information:"
	@echo "  Go version: $(shell go version)"
	@echo "  GCC version: $(shell gcc --version | head -n1)"
	@echo "  OS: $(shell uname -s)"
	@echo "  Architecture: $(shell uname -m)"
	@echo "  CGO enabled: $(shell go env CGO_ENABLED)"

# Test de la librería como módulo independiente
test-module:
	@echo "Testing argo as independent module..."
	@go build -o test_argo argo.go
	@echo "✅ Argo module compiles independently"
	@rm -f test_argo

# Ayuda
help:
	@echo "Available targets:"
	@echo "  all           - Build the project (default)"
	@echo "  build         - Compile the project"
	@echo "  run           - Build and run the examples"
	@echo "  test          - Run tests for main library"
	@echo "  test-example  - Run tests for example"
	@echo "  test-module   - Test that argo.go compiles independently"
	@echo "  benchmark     - Run benchmarks"
	@echo "  clean         - Remove generated files"
	@echo "  install       - Install Go module dependencies"
	@echo "  debug         - Build with debug flags"
	@echo "  release       - Build with release flags"
	@echo "  check         - Verify presence of key files"
	@echo "  info          - Show system and tool versions"
	@echo "  help          - Display this help message"