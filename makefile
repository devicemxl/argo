# Makefile para Arena + CGO proyecto

# Variables
CC = gcc
CFLAGS = -std=c99 -O2 -Wall -Wextra -fPIC
LDFLAGS = -shared -lm

# Archivos
# C_SOURCES y C_HEADERS se usan indirectamente por Go a través de cgo
# NO necesitas listar argo.go aquí como un archivo fuente de Go para el BINARY
# El 'go build' lo resolverá automáticamente porque main.go lo importa.
# Especificamos la ruta del main.go
GO_MAIN_SOURCE = ./example/main.go
BINARY = arena_cgo_example # El nombre del binario que se generará

# Objetivos principales
.PHONY: all clean build run test benchmark install debug release check info help

all: build

# Compilar el proyecto
build: # NO necesitas listar GO_SOURCES directamente aquí
	@echo "Building Arena + CGO project..."
	# go build compilará el main.go en examples/ y automáticamente incluirá argo.go
	# porque main.go lo importa desde el módulo raíz.
	@go build -o $(BINARY) $(GO_MAIN_SOURCE)
	@echo "Build completed: $(BINARY)"

# Ejecutar el programa
run: build
	@echo "Running Arena + CGO examples..."
	@./$(BINARY)

# Ejecutar tests
test:
	@echo "Running tests..."
	# Esto correrá los tests en todos los paquetes del módulo raíz, incluyendo argo.go
	@go test -v ./...

# Ejecutar benchmarks
benchmark:
	@echo "Running benchmarks..."
	# Esto correrá los benchmarks en todos los paquetes del módulo raíz
	@go test -bench=. -benchmem ./...

# Compilar biblioteca C por separado (opcional, Go lo hace por ti para CGO)
# Si necesitas compilarla fuera de Go por alguna razón, esto está bien
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
	@rm -f libcgo_utils.so # Si generas la lib C por separado
	@rm -f go.sum
	@rm -rf vendor # Si usas vendor, esto lo limpia
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
	@test -f cgo_utils.h || (echo "Missing: cgo_utils.h" && exit 1)
	@test -f cgo_utils.c || (echo "Missing: cgo_utils.c" && exit 1)
	@test -f argo.go || (echo "Missing: argo.go" && exit 1)
	@test -f examples/main.go || (echo "Missing: examples/main.go" && exit 1) # Ruta corregida
	@echo "All required files present"

# Información del sistema
info:
	@echo "System Information:"
	@echo "  Go version: $(shell go version)"
	@echo "  GCC version: $(shell gcc --version | head -n1)"
	@echo "  OS: $(shell uname -s)"
	@echo "  Architecture: $(shell uname -m)"
	@echo "  CGO enabled: $(shell go env CGO_ENABLED)"

# Ayuda
help:
	@echo "Available targets:"
	@echo "  all        - Build the project (default)"
	@echo "  build      - Compile the project"
	@echo "  run        - Build and run the examples"
	@echo "  test       - Run tests"
	@echo "  benchmark  - Run benchmarks"
	@echo "  clean      - Remove generated files"
	@echo "  install    - Install Go module dependencies"
	@echo "  debug      - Build with debug flags"
	@echo "  release    - Build with release flags"
	@echo "  check      - Verify presence of key files"
	@echo "  info       - Show system and tool versions"
	@echo "  help       - Display this help message"