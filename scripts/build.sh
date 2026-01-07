#!/bin/bash

# Script de build avec versioning automatique
# Usage: ./scripts/build.sh [version]

set -e

# Couleurs pour les logs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Fonction de log
log() {
    echo -e "${BLUE}[BUILD]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Vérifier que nous sommes dans le bon répertoire
if [ ! -f "go.mod" ]; then
    error "Ce script doit être exécuté depuis la racine du projet"
fi

# Récupérer les informations de versioning
VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

log "Informations de build:"
log "  Version: $VERSION"
log "  Git Commit: $GIT_COMMIT"
log "  Build Time: $BUILD_TIME"

# Exporter les variables pour docker-compose
export VERSION
export GIT_COMMIT
export BUILD_TIME

# Nettoyer les anciens builds
log "Nettoyage des anciens builds..."
make clean || warn "Erreur lors du nettoyage"

# Build local avec Make
log "Build local avec Make..."
make ci || error "Échec du build local"

# Build des images Docker
log "Construction des images Docker..."

# Build de l'image API
log "Construction de l'image API..."
docker build \
    --build-arg VERSION="$VERSION" \
    --build-arg GIT_COMMIT="$GIT_COMMIT" \
    --build-arg BUILD_TIME="$BUILD_TIME" \
    -t go-api-mongo-scrapper:$VERSION \
    -t go-api-mongo-scrapper:latest \
    -f dockerfile .

# Build de l'image Scraper
log "Construction de l'image Scraper..."
docker build \
    --build-arg VERSION="$VERSION" \
    --build-arg GIT_COMMIT="$GIT_COMMIT" \
    --build-arg BUILD_TIME="$BUILD_TIME" \
    -t go-scraper:$VERSION \
    -t go-scraper:latest \
    -f scraper/dockerfile .

# Vérifier les images
log "Vérification des images construites..."
docker images | grep -E "(go-api-mongo-scrapper|go-scraper)" || error "Images non trouvées"

# Test des images
log "Test des images..."

# Test de l'image API (health check)
log "Test de l'image API..."
CONTAINER_ID=$(docker run -d -p 8082:8080 go-api-mongo-scrapper:$VERSION)
sleep 5

if docker exec $CONTAINER_ID wget --no-verbose --tries=1 --spider http://localhost:8080/version; then
    success "Image API fonctionne correctement"
else
    warn "L'image API ne répond pas correctement"
fi

docker stop $CONTAINER_ID >/dev/null 2>&1
docker rm $CONTAINER_ID >/dev/null 2>&1

# Test de l'image Scraper (version info)
log "Test de l'image Scraper..."
if docker run --rm go-scraper:$VERSION --help >/dev/null 2>&1; then
    success "Image Scraper fonctionne correctement"
else
    warn "L'image Scraper ne répond pas correctement"
fi

# Afficher les tailles des images
log "Tailles des images:"
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | grep -E "(go-api-mongo-scrapper|go-scraper)"

# Optionnel: Push vers le registry
if [ "$PUSH_TO_REGISTRY" = "true" ]; then
    log "Push vers le registry..."
    
    REGISTRY=${REGISTRY:-ghcr.io/maxime-louis14}
    
    # Tag pour le registry
    docker tag go-api-mongo-scrapper:$VERSION $REGISTRY/go-api-mongo-scrapper:$VERSION
    docker tag go-api-mongo-scrapper:latest $REGISTRY/go-api-mongo-scrapper:latest
    docker tag go-scraper:$VERSION $REGISTRY/go-scraper:$VERSION
    docker tag go-scraper:latest $REGISTRY/go-scraper:latest
    
    # Push
    docker push $REGISTRY/go-api-mongo-scrapper:$VERSION
    docker push $REGISTRY/go-api-mongo-scrapper:latest
    docker push $REGISTRY/go-scraper:$VERSION
    docker push $REGISTRY/go-scraper:latest
    
    success "Images pushées vers $REGISTRY"
fi

success "Build terminé avec succès!"
log "Pour démarrer l'application:"
log "  docker-compose up api-server mongodb"
log "Pour lancer le scraper:"
log "  docker-compose --profile scraper up scraper"
log "Pour lancer mongo-express:"
log "  docker-compose --profile tools up mongo-express" 