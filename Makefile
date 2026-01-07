# Makefile pour le projet Go API Mongo Scrapper

.PHONY: test test-verbose test-coverage benchmark clean build run help ci cd release docker scripts

# Variables
BINARY_NAME=scraper
SCRAPER_DIR=./scraper
VERSION?=production
REGISTRY?=ghcr.io
IMAGE_NAME?=go-api-mongo-scrapper
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Commandes de test
test:
	@echo "Exécution des tests unitaires..."
	cd $(SCRAPER_DIR) && go test -v

test-verbose:
	@echo "Exécution des tests unitaires (mode verbose)..."
	cd $(SCRAPER_DIR) && go test -v -race

test-coverage:
	@echo "Génération du rapport de couverture..."
	cd $(SCRAPER_DIR) && go test -v -race -coverprofile=coverage.out
	cd $(SCRAPER_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "Rapport de couverture généré: scraper/coverage.html"

benchmark:
	@echo "Exécution des benchmarks..."
	cd $(SCRAPER_DIR) && go test -bench=. -benchmem

# Commandes de build
build:
	@echo "Compilation du scraper..."
	cd $(SCRAPER_DIR) && go build -o $(BINARY_NAME) .

build-server:
	@echo "Compilation du serveur API..."
	go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o server ./main.go

build-scraper:
	@echo "Compilation du scraper avec versioning..."
	cd $(SCRAPER_DIR) && go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o $(BINARY_NAME) ./scraper.go

build-all: build-server build-scraper
	@echo "Compilation complète terminée"

run:
	@echo "Exécution du scraper..."
	cd $(SCRAPER_DIR) && go run .

run-server:
	@echo "Exécution du serveur API..."
	go run ./main.go

# Commandes de nettoyage
clean:
	@echo "Nettoyage des fichiers temporaires..."
	cd $(SCRAPER_DIR) && rm -f $(BINARY_NAME) coverage.out coverage.html test_recipes.json
	rm -f server
	rm -rf dist/
	docker system prune -f || true

# Commandes de développement
deps:
	@echo "Installation des dépendances..."
	go mod download
	go mod tidy

lint:
	@echo "Vérification du code avec golint..."
	cd $(SCRAPER_DIR) && golint ./...

fmt:
	@echo "Formatage du code..."
	cd $(SCRAPER_DIR) && go fmt ./...
	go fmt ./...

vet:
	@echo "Analyse statique du code..."
	cd $(SCRAPER_DIR) && go vet ./...
	go vet ./...

# Commandes CI/CD
ci: deps fmt vet test build-all
	@echo "Pipeline CI terminé avec succès"

ci-full: deps fmt vet test-coverage benchmark build-all docker-build
	@echo "Pipeline CI complet terminé avec succès"

# Commandes Docker améliorées
docker-build:
	@echo "Construction des images Docker avec versioning..."
	./scripts/build.sh $(VERSION)

docker-build-api:
	@echo "Construction de l'image API..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest \
		-f dockerfile .

docker-build-scraper:
	@echo "Construction de l'image Scraper..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t go-scraper:$(VERSION) \
		-t go-scraper:latest \
		-f scraper/dockerfile .

docker-run:
	@echo "Démarrage de l'application avec Docker Compose..."
	VERSION=$(VERSION) GIT_COMMIT=$(GIT_COMMIT) BUILD_TIME=$(BUILD_TIME) \
	docker-compose up api-server mongodb

docker-run-scraper:
	@echo "Exécution du scraper avec Docker..."
	VERSION=$(VERSION) GIT_COMMIT=$(GIT_COMMIT) BUILD_TIME=$(BUILD_TIME) \
	docker-compose --profile scraper up scraper

docker-run-tools:
	@echo "Démarrage des outils (MongoDB Express)..."
	VERSION=$(VERSION) GIT_COMMIT=$(GIT_COMMIT) BUILD_TIME=$(BUILD_TIME) \
	docker-compose --profile tools up mongo-express

docker-stop:
	@echo "Arrêt des containers Docker..."
	docker-compose down

docker-logs:
	@echo "Affichage des logs Docker..."
	docker-compose logs -f

docker-push:
	@echo "Push des images vers le registry..."
	PUSH_TO_REGISTRY=true ./scripts/build.sh $(VERSION)

# Commandes de release
release-prepare:
	@echo "Préparation de la release $(VERSION)..."
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "Erreur: VERSION doit être définie (ex: make release-prepare VERSION=v1.0.0)"; \
		exit 1; \
	fi
	git tag -a $(VERSION) -m "Release $(VERSION)"

release-build:
	@echo "Build de release pour $(VERSION)..."
	mkdir -p dist
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o dist/server-linux-amd64 ./main.go
	GOOS=linux GOARCH=amd64 cd $(SCRAPER_DIR) && go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o ../dist/scraper-linux-amd64 ./scraper.go
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o dist/server-linux-arm64 ./main.go
	GOOS=linux GOARCH=arm64 cd $(SCRAPER_DIR) && go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o ../dist/scraper-linux-arm64 ./scraper.go
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o dist/server-windows-amd64.exe ./main.go
	GOOS=windows GOARCH=amd64 cd $(SCRAPER_DIR) && go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o ../dist/scraper-windows-amd64.exe ./scraper.go
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o dist/server-darwin-amd64 ./main.go
	GOOS=darwin GOARCH=amd64 cd $(SCRAPER_DIR) && go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o ../dist/scraper-darwin-amd64 ./scraper.go
	# macOS arm64
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o dist/server-darwin-arm64 ./main.go
	GOOS=darwin GOARCH=arm64 cd $(SCRAPER_DIR) && go build -ldflags="-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)" -o ../dist/scraper-darwin-arm64 ./scraper.go

release: release-build docker-build
	@echo "Release $(VERSION) prête"

# Commandes de déploiement
deploy-staging: ci docker-build
	@echo "Déploiement en staging..."
	# Ajoutez ici vos commandes de déploiement staging

deploy-production: ci docker-build
	@echo "Déploiement en production..."
	# Ajoutez ici vos commandes de déploiement production

# Commandes de monitoring
health-check:
	@echo "Vérification de l'état de l'application..."
	curl -f http://localhost:8080/health || echo "Service non disponible"

version-check:
	@echo "Vérification des informations de version..."
	curl -f http://localhost:8080/version || echo "Service non disponible"

logs:
	@echo "Affichage des logs..."
	docker logs go-api-server || echo "Container non trouvé"

# Scripts utilitaires
scripts-setup:
	@echo "Configuration des scripts..."
	chmod +x scripts/*.sh

# Commande d'aide
help:
	@echo "Commandes disponibles:"
	@echo "  test              - Exécuter les tests unitaires"
	@echo "  test-verbose      - Exécuter les tests avec race detection"
	@echo "  test-coverage     - Générer un rapport de couverture HTML"
	@echo "  benchmark         - Exécuter les benchmarks"
	@echo "  build             - Compiler le scraper"
	@echo "  build-server      - Compiler le serveur API"
	@echo "  build-scraper     - Compiler le scraper avec versioning"
	@echo "  build-all         - Compiler scraper et serveur"
	@echo "  run               - Exécuter le scraper"
	@echo "  run-server        - Exécuter le serveur API"
	@echo "  clean             - Nettoyer les fichiers temporaires"
	@echo "  deps              - Installer les dépendances"
	@echo "  lint              - Vérifier le code avec golint"
	@echo "  fmt               - Formater le code"
	@echo "  vet               - Analyse statique du code"
	@echo "  ci                - Pipeline CI complet"
	@echo "  ci-full           - Pipeline CI avec couverture et benchmarks"
	@echo "  docker-build      - Construire toutes les images Docker"
	@echo "  docker-build-api  - Construire l'image API"
	@echo "  docker-build-scraper - Construire l'image Scraper"
	@echo "  docker-run        - Démarrer l'application"
	@echo "  docker-run-scraper - Exécuter le scraper"
	@echo "  docker-run-tools  - Démarrer MongoDB Express"
	@echo "  docker-stop       - Arrêter les containers"
	@echo "  docker-logs       - Afficher les logs"
	@echo "  docker-push       - Push les images vers le registry"
	@echo "  release-build     - Build de release multi-plateforme"
	@echo "  release           - Créer une release complète"
	@echo "  deploy-staging    - Déployer en staging"
	@echo "  deploy-production - Déployer en production"
	@echo "  health-check      - Vérifier l'état de l'application"
	@echo "  version-check     - Vérifier les informations de version"
	@echo "  logs              - Afficher les logs"
	@echo "  scripts-setup     - Configurer les scripts"
	@echo "  help              - Afficher cette aide"

# Commande par défaut
all: deps fmt vet test build-all 