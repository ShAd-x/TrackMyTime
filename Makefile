.PHONY: build run clean test install help

# Variables
BINARY_NAME=trackmytime
CMD_PATH=./cmd/agent
BUILD_DIR=build

# Build the application
build:
	@echo "🔨 Compilation de l'agent..."
	go build -o $(BINARY_NAME) $(CMD_PATH)
	@echo "✅ Binaire créé: $(BINARY_NAME)"

# Run the application
run:
	@echo "🚀 Démarrage de l'agent..."
	go run $(CMD_PATH)/main.go

# Build and run
build-run: build
	@echo "🚀 Lancement de l'agent..."
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "🧹 Nettoyage..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	@echo "✅ Nettoyé"

# Run tests
test:
	@echo "🧪 Lancement des tests..."
	go test -v ./...

# Install dependencies
install:
	@echo "📦 Installation des dépendances..."
	go mod download
	go mod tidy
	@echo "✅ Dépendances installées"

# Build for all platforms
build-all:
	@echo "🔨 Compilation multi-plateformes..."
	@mkdir -p $(BUILD_DIR)
	@echo "  → Linux amd64..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	@echo "  → Windows amd64..."
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "  → macOS amd64..."
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	@echo "  → macOS arm64 (M1/M2)..."
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "✅ Tous les binaires créés dans $(BUILD_DIR)/"

# Display help
help:
	@echo "TrackMyTime - Agent de Tracking d'Activité"
	@echo ""
	@echo "Commandes disponibles:"
	@echo "  make build       - Compiler l'agent"
	@echo "  make run         - Exécuter l'agent (sans compiler)"
	@echo "  make build-run   - Compiler puis exécuter"
	@echo "  make clean       - Nettoyer les fichiers compilés"
	@echo "  make test        - Lancer les tests"
	@echo "  make install     - Installer les dépendances"
	@echo "  make build-all   - Compiler pour toutes les plateformes"
	@echo "  make help        - Afficher cette aide"

# Build export command
build-export:
	@echo "🔨 Compilation de l'outil d'export..."
	go build -o trackmytime-export ./cmd/export
	@echo "✅ Outil d'export créé: trackmytime-export"

# Build both
build-all-cmds: build build-export

# Export aggregated stats (today)
export-today:
	@./trackmytime-export -aggregated

# Export aggregated stats (week)
export-week:
	@./trackmytime-export -aggregated -period week
