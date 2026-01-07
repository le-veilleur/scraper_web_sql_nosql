# Étape 1 : Construire le binaire
FROM golang:1.22-alpine AS builder

# Arguments de build pour le versioning
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

# Installer les outils nécessaires
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le reste du code source
COPY . .

# Construire le binaire API avec versioning
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o api-server ./main.go

# Construire le binaire scraper avec versioning
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -o scraper-binary ./scraper/scraper.go

# Étape 2 : Image finale minimale
FROM scratch

# Redéclarer les arguments pour cette étape
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

# Copier les certificats CA pour HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copier les données de timezone
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

# Copier les binaires
COPY --from=builder /app/api-server /app/api-server
COPY --from=builder /app/scraper-binary /app/scraper

# Labels pour traçabilité
LABEL version="${VERSION}" \
      git.commit="${GIT_COMMIT}" \
      build.time="${BUILD_TIME}" \
      maintainer="Maxime Louis <maxime.louis14@example.com>" \
      description="Go API MongoDB Scrapper - API Server" \
      org.opencontainers.image.source="https://github.com/maxime-louis14/go_api_mongo_scrapper"

# Variables d'environnement par défaut
ENV PORT=8080 \
    ENV=production \
    LOG_LEVEL=info

# Exposer les ports
EXPOSE 8080

# Démarrer l'application
ENTRYPOINT ["/app/api-server"]