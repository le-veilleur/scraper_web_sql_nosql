# 🍽️ Go API MongoDB Scrapper

[![CI](https://github.com/le-veilleur/go_api_mongo_scrapper/workflows/Continuous%20Integration/badge.svg)](https://github.com/le-veilleur/go_api_mongo_scrapper/actions/workflows/ci.yml)
[![CD](https://github.com/le-veilleur/go_api_mongo_scrapper/workflows/Continuous%20Deployment/badge.svg)](https://github.com/le-veilleur/go_api_mongo_scrapper/actions/workflows/cd.yml)
[![Release](https://github.com/le-veilleur/go_api_mongo_scrapper/workflows/Release/badge.svg)](https://github.com/le-veilleur/go_api_mongo_scrapper/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/le-veilleur/go_api_mongo_scrapper)](https://goreportcard.com/report/github.com/le-veilleur/go_api_mongo_scrapper)

> **API REST performante en Go avec MongoDB et scraper de recettes parallèle pour le restaurant Hótwings**

Une solution complète développée pour le restaurant Hótwings afin d'étendre son activité avec un service de livraison. L'API propose une carte étendue de recettes scrapées depuis AllRecipes.com avec un système de scraping parallèle optimisé utilisant des goroutines.

## 📋 Table des matières

- [🎯 Aperçu du projet](#-aperçu-du-projet)
- [🛠️ Technologies utilisées](#️-technologies-utilisées)
- [✨ Fonctionnalités](#-fonctionnalités)
- [🏗️ Architecture](#️-architecture)
- [🚀 Démarrage rapide](#-démarrage-rapide)
- [📚 Documentation API](#-documentation-api)
- [🔧 Configuration](#-configuration)
- [🧪 Tests](#-tests)
- [🐳 Docker](#-docker)
- [⚡ Performance](#-performance)
- [📊 Monitoring](#-monitoring)
- [🔄 CI/CD](#-cicd)
- [🤝 Contribution](#-contribution)
- [📄 Licence](#-licence)

## 🎯 Aperçu du projet

### Contexte métier

Le restaurant **Hótwings** souhaite développer son activité avec un service de livraison en proposant une carte très étendue de plats et recettes. Pour plaire à tous les goûts, l'API permet de proposer une large variété de recettes scrapées depuis [AllRecipes.com](https://www.allrecipes.com/).

### Objectifs techniques

- ✅ **API REST** complète avec Fiber framework
- ✅ **Base de données** MongoDB avec Docker
- ✅ **Scraper performant** avec goroutines parallèles
- ✅ **Tests complets** avec couverture de code
- ✅ **CI/CD automatisé** avec GitHub Actions
- ✅ **Containerisation** Docker complète
- ✅ **Cross-platform** binaires pour Linux, Windows, macOS

## 🛠️ Technologies utilisées

### Langages et Runtime
- **Go (Golang)** 1.22+ - Langage de programmation principal
- **Bash** - Scripts d'automatisation et de déploiement

### Frameworks et Bibliothèques Go
- **Fiber v2** (`github.com/gofiber/fiber/v2`) - Framework web HTTP rapide et Express-like
- **Colly** (`github.com/gocolly/colly`) - Framework de web scraping/scraping
- **MongoDB Driver** (`go.mongodb.org/mongo-driver`) - Driver officiel MongoDB pour Go
- **godotenv** (`github.com/joho/godotenv`) - Gestion des variables d'environnement depuis `.env`
- **testify** (`github.com/stretchr/testify`) - Framework de tests avec assertions

### Bases de données
- **MongoDB 7.0** - Base de données NoSQL principale
- **Mongo Express** - Interface web pour la gestion de MongoDB

### Containerisation et Orchestration
- **Docker** - Containerisation des services
- **Docker Compose** - Orchestration multi-conteneurs
- **Dockerfile** - Images personnalisées pour API et scraper

### CI/CD et Déploiement
- **GitHub Actions** - Pipeline CI/CD automatisé
- **SSH** - Déploiement automatisé sur VPS
- **Git** - Contrôle de version

### Outils de développement
- **Make** - Automatisation des tâches de build et de test
- **Go Modules** - Gestion des dépendances
- **Bash Scripts** - Scripts d'automatisation (`build.sh`, `test_metrics.sh`, etc.)

### Formats et Protocoles
- **JSON** - Format d'échange de données
- **REST API** - Architecture API RESTful
- **HTTP/HTTPS** - Protocoles de communication

### Infrastructure et Services
- **VPS (Virtual Private Server)** - Serveur de production
- **Linux** - Système d'exploitation serveur
- **Port 8082** - Port par défaut de l'API
- **Port 27017/27018** - Port MongoDB
- **Port 8081** - Port Mongo Express

### Techniques et Concepts
- **Goroutines** - Concurrence et parallélisme en Go
- **Channels** - Communication entre goroutines
- **Sync (WaitGroups, Mutexes)** - Synchronisation des goroutines
- **Web Scraping** - Collecte automatisée de données web
- **Anti-bot Measures** - Techniques anti-détection (User-Agent rotation, headers réalistes, délais aléatoires)
- **Structured Logging** - Système de logs structurés
- **Health Checks** - Monitoring de l'état des services
- **Metrics** - Collecte de métriques de performance

### Middleware et Sécurité
- **CORS** - Cross-Origin Resource Sharing
- **Recovery** - Gestion des panics
- **Logger** - Middleware de logging HTTP
- **Environment Variables** - Configuration via variables d'environnement

### Monitoring et Observabilité
- **Health Endpoints** (`/health`, `/version`, `/metrics`)
- **Structured Logs** - Logs JSON structurés
- **Performance Metrics** - Métriques de performance en temps réel

## ✨ Fonctionnalités

### 🔍 Fonctionnalités de lecture
- **Lister les recettes** - Récupération de toutes les recettes avec pagination
- **Détail d'une recette** - Informations complètes : ingrédients, instructions, image
- **Recherche avancée** - Par nom de recette ou ingrédient

### 🔄 Importation de données
- **Import JSON** - Importation de recettes depuis fichier JSON
- **Scraper automatique** - Récupération automatique depuis AllRecipes.com
- **Gestion des erreurs** - Système robuste de gestion d'erreurs

### 🛠️ Outils et monitoring
- **Health checks** - Endpoints de santé de l'application
- **Métriques** - Monitoring en temps réel
- **Logs structurés** - Système de logging avancé
- **Swagger** - Documentation API interactive

## 🏗️ Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        WEB[Web Client]
        API_CLIENT[API Client]
    end
    
    subgraph "API Layer"
        FIBER[Fiber Server<br/>Port 8080]
        MIDDLEWARE[CORS, Logging, Recovery]
        ROUTES[Recipe Routes]
    end
    
    subgraph "Business Layer"
        CONTROLLERS[Controllers]
        MODELS[Data Models]
        RESPONSES[API Responses]
    end
    
    subgraph "Data Layer"
        MONGODB[(MongoDB<br/>Port 27017)]
        SCRAPER_DATA[JSON Files]
    end
    
    subgraph "Scraper Layer"
        SCRAPER[Go Scraper<br/>Colly Framework]
        WORKERS[Goroutines<br/>Parallel Processing]
        ALLRECIPES[AllRecipes.com]
    end
    
    subgraph "Infrastructure"
        DOCKER[Docker Compose]
        MONGO_EXPRESS[Mongo Express<br/>Port 8081]
        LOGS[Structured Logs]
    end
    
    WEB --> FIBER
    API_CLIENT --> FIBER
    FIBER --> MIDDLEWARE
    MIDDLEWARE --> ROUTES
    ROUTES --> CONTROLLERS
    CONTROLLERS --> MODELS
    MODELS --> MONGODB
    CONTROLLERS --> RESPONSES
    SCRAPER --> WORKERS
    WORKERS --> ALLRECIPES
    SCRAPER --> SCRAPER_DATA
    SCRAPER_DATA --> CONTROLLERS
    MONGODB --> MONGO_EXPRESS
    FIBER --> LOGS
    DOCKER --> FIBER
    DOCKER --> MONGODB
    DOCKER --> SCRAPER
```

### Structure du projet

```
go_api_mongo_scrapper/
├── 📁 api-server/          # Serveur API principal
├── 📁 controllers/         # Contrôleurs API
├── 📁 database/           # Configuration MongoDB
├── 📁 docs/              # Documentation complète
├── 📁 logger/            # Système de logging
├── 📁 middleware/        # Middlewares Fiber
├── 📁 models/            # Modèles de données
├── 📁 responses/         # Réponses API standardisées
├── 📁 routes/            # Définition des routes
├── 📁 scraper/           # Module de scraping
│   ├── scraper.go        # Code principal du scraper
│   ├── scraper_test.go   # Tests unitaires
│   └── README_TESTS.md   # Documentation des tests
├── 📁 scripts/           # Scripts de build et déploiement
├── 📄 docker-compose.yml # Configuration Docker
├── 📄 dockerfile         # Image Docker API
├── 📄 Makefile          # Commandes de développement
└── 📄 main.go           # Point d'entrée de l'API
```

## 🚀 Démarrage rapide

### Prérequis

- **Go** 1.22+
- **Docker** & Docker Compose
- **Make** (optionnel mais recommandé)
- **Git**

### Installation en 3 étapes

```bash
# 1. Cloner le repository
git clone https://github.com/le-veilleur/go_api_mongo_scrapper.git
cd go_api_mongo_scrapper

# 2. Démarrer l'infrastructure
docker-compose up -d

# 3. Lancer l'API
go run main.go
```

### Vérification de l'installation

```bash
# Vérifier que l'API fonctionne
curl http://localhost:8080/health

# Vérifier les informations de version
curl http://localhost:8080/version

# Accéder à l'interface MongoDB
# http://localhost:8081 (admin/admin123)
```

## 📚 Documentation API

### Endpoints principaux

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| `GET` | `/health` | État de santé de l'API |
| `GET` | `/version` | Informations de version |
| `GET` | `/metrics` | Métriques de l'application |
| `GET` | `/recipes` | Liste des recettes |
| `POST` | `/recipes` | Créer une recette |
| `GET` | `/recipes/:id` | Récupérer une recette |
| `PUT` | `/recipes/:id` | Modifier une recette |
| `DELETE` | `/recipes/:id` | Supprimer une recette |

### Exemples d'utilisation

#### Récupérer toutes les recettes

```bash
curl -X GET "http://localhost:8080/recipes" \
  -H "Content-Type: application/json"
```

**Réponse :**
```json
{
  "success": true,
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "name": "Chocolate Chip Cookies",
      "image": "https://example.com/cookies.jpg",
      "ingredients": [
        {
          "quantity": "2",
          "unit": "cups",
          "name": "flour"
        }
      ],
      "instructions": [
        {
          "step": 1,
          "description": "Preheat oven to 375°F"
        }
      ],
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 150
  }
}
```

#### Créer une nouvelle recette

```bash
curl -X POST "http://localhost:8080/recipes" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pasta Carbonara",
    "image": "https://example.com/carbonara.jpg",
    "ingredients": [
      {
        "quantity": "500",
        "unit": "g",
        "name": "pasta"
      }
    ],
    "instructions": [
      {
        "step": 1,
        "description": "Boil water and cook pasta"
      }
    ]
  }'
```

#### Rechercher des recettes

```bash
# Recherche par nom
curl -X GET "http://localhost:8080/recipes?search=pasta"

# Recherche par ingrédient
curl -X GET "http://localhost:8080/recipes?ingredient=tomato"
```

### Health Check

```bash
curl http://localhost:8080/health
```

**Réponse :**
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z",
  "build": {
    "version": "1.0.0",
    "git_commit": "abc1234",
    "build_time": "2024-01-15T10:00:00Z",
    "go_version": "go1.22.0",
    "os": "linux",
    "arch": "amd64"
  },
  "database": "connected"
}
```

## 🔧 Configuration

### Variables d'environnement

| Variable | Description | Valeur par défaut |
|----------|-------------|-------------------|
| `PORT` | Port du serveur API | `8080` |
| `MONGODB_URI` | URI de connexion MongoDB | `mongodb://admin:password123@localhost:27017/recipes?authSource=admin` |
| `DB_NAME` | Nom de la base de données | `recipes` |
| `LOG_LEVEL` | Niveau de logging | `info` |
| `ENV` | Environnement | `development` |

### Configuration Docker

Le fichier `docker-compose.yml` configure :

- **MongoDB** : Base de données principale
- **API Server** : Serveur Go avec Fiber
- **Mongo Express** : Interface web MongoDB (optionnel)
- **Scraper** : Service de scraping (optionnel)

### Configuration du scraper

| Variable | Description | Valeur par défaut |
|----------|-------------|-------------------|
| `SCRAPER_MAX_WORKERS` | Nombre de workers parallèles | `12` (adaptatif) |
| `SCRAPER_TIMEOUT` | Timeout des requêtes | `30s` |
| `SCRAPER_BASE_URL` | URL de base à scraper | `https://www.allrecipes.com` |
| `SCRAPER_MAX_PAGES` | Nombre maximum de pages | `5` |
| `SCRAPER_MAX_RECIPES_PER_PAGE` | Recettes par page | `20` |

## 🧪 Tests

### Exécution des tests

```bash
# Tests unitaires
make test

# Tests avec race detection
make test-verbose

# Rapport de couverture HTML
make test-coverage

# Benchmarks de performance
make benchmark
```

### Couverture de code

Le projet maintient une couverture de tests de **22.6%** avec :

- ✅ **12 tests unitaires** complets
- ✅ **2 benchmarks** de performance
- ✅ **Tests de concurrence** avec race detection
- ✅ **Tests de validation** des modèles
- ✅ **Tests d'intégration** API

### Rapport de couverture

```bash
# Générer le rapport HTML
make test-coverage

# Ouvrir le rapport
open scraper/coverage.html
```

## 🐳 Docker

### Images disponibles

```bash
# Dernière version
docker pull ghcr.io/maxime-louis14/go_api_mongo_scrapper:latest

# Version spécifique
docker pull ghcr.io/maxime-louis14/go_api_mongo_scrapper:v1.0.0
```

### Commandes Docker

```bash
# Démarrer l'application complète
docker-compose up -d

# Démarrer avec le scraper
docker-compose --profile scraper up -d

# Démarrer avec MongoDB Express
docker-compose --profile tools up -d

# Voir les logs
docker-compose logs -f

# Arrêter les services
docker-compose down
```

### Build des images

```bash
# Build de l'API
make docker-build-api

# Build du scraper
make docker-build-scraper

# Build complet
make docker-build
```

## ⚡ Performance

### Scraper

- **Parallélisme** : 12 workers adaptatifs (6 cœurs × 2)
- **Vitesse** : 650 recettes en 21.4 secondes (~30 recettes/seconde)
- **Mémoire** : Optimisé avec channels et sync.Pool
- **Robustesse** : Gestion d'erreurs et timeouts configurables

### API

- **Framework** : Fiber (Express-like pour Go)
- **Base de données** : MongoDB avec indexation optimisée
- **Middleware** : CORS, logging, compression, recovery
- **Performance** : ~10k req/s en conditions optimales
- **Latence** : < 50ms pour les requêtes simples

### Benchmarks

```bash
# Exécuter les benchmarks
make benchmark
```

**Résultats réels (dernière exécution) :**
```
📊 STATISTIQUES DÉTAILLÉES DU SCRAPER
⏱️  Durée totale: 21.46s
🚀 Requêtes par seconde: 30.29
📝 Recettes par seconde: 29.83
🌐 Total requêtes: 650
📝 Recettes trouvées: 640
✅ Taux de succès: 100.0%
💻 Workers: 12 (6 cœurs × 2 ratio adaptatif)
```

**Benchmarks de code :**
```
BenchmarkScraper-8          100    12345678 ns/op    4567890 B/op    12345 allocs/op
BenchmarkAPI-8             1000     1234567 ns/op     123456 B/op     1234 allocs/op
```

## 📊 Monitoring

### Métriques disponibles

- **Health check** : `/health` - État de l'application
- **Version** : `/version` - Informations de build
- **Métriques** : `/metrics` - Métriques détaillées JSON

### Logs structurés

Le système de logging inclut :

- **Niveaux** : DEBUG, INFO, WARN, ERROR
- **Format** : JSON structuré
- **Rotation** : Logs rotatifs automatiques
- **Métriques** : Temps de réponse, erreurs, throughput

### Monitoring en production

```bash
# Vérifier l'état de l'application
make health-check

# Vérifier les informations de version
make version-check

# Voir les logs en temps réel
make logs
```

## 🔄 CI/CD

### Pipeline GitHub Actions

Le projet utilise 3 workflows principaux :

#### 1. Continuous Integration (CI)
- **Déclencheurs** : Push/PR sur `main` et `develop`
- **Tests** : Tests unitaires avec race detection
- **Code Quality** : Formatage, linting, analyse statique
- **Security** : Scan de sécurité avec Gosec
- **Build** : Compilation cross-platform

#### 2. Continuous Deployment (CD)
- **Staging** : Déploiement automatique sur push vers `main`
- **Production** : Déploiement sur tags `v*`
- **Rollback** : Rollback automatique en cas d'échec

#### 3. Release
- **Binaires** : Multi-plateformes (Linux, Windows, macOS)
- **Docker** : Images multi-architecture
- **Assets** : Changelog automatique et assets GitHub

### Commandes de développement

```bash
# Pipeline CI local
make ci

# Pipeline CI complet avec couverture
make ci-full

# Créer une release
make release VERSION=v1.0.0
```

## 🤝 Contribution

### Workflow de contribution

1. **Fork** le projet
2. **Créer** une branche feature (`git checkout -b feature/AmazingFeature`)
3. **Développer** et tester (`make test && make ci`)
4. **Commit** les changements (`git commit -m 'Add some AmazingFeature'`)
5. **Push** vers la branche (`git push origin feature/AmazingFeature`)
6. **Ouvrir** une Pull Request

### Standards de code

- **Formatage** : `gofmt` obligatoire
- **Linting** : `golangci-lint` sans erreurs
- **Tests** : Couverture minimale de 80%
- **Documentation** : Commentaires Go standard
- **Commits** : Messages en français

### Commandes de développement

```bash
# Installation des dépendances
make deps

# Formatage du code
make fmt

# Analyse statique
make vet

# Linting
make lint

# Tests complets
make test-coverage

# Nettoyage
make clean
```

## 📄 Licence

Ce projet est sous licence **MIT**. Voir le fichier [LICENSE](LICENSE) pour plus de détails.

## 🆘 Support

- **Issues** : [GitHub Issues](https://github.com/le-veilleur/go_api_mongo_scrapper/issues)
- **Discussions** : [GitHub Discussions](https://github.com/le-veilleur/go_api_mongo_scrapper/discussions)
- **Documentation** : [docs/](docs/)

## 🗺️ Roadmap

### Version 1.1.0
- [ ] **Authentification JWT** - Système d'authentification sécurisé
- [ ] **Rate limiting** - Protection contre les abus
- [ ] **Cache Redis** - Amélioration des performances
- [ ] **Métriques Prometheus** - Monitoring avancé

### Version 1.2.0
- [ ] **Dashboard Grafana** - Interface de monitoring
- [ ] **Tests E2E** - Tests d'intégration complets
- [ ] **Déploiement Kubernetes** - Orchestration container
- [ ] **API GraphQL** - Alternative à REST

### Version 2.0.0
- [ ] **Microservices** - Architecture distribuée
- [ ] **Event Sourcing** - Historique des événements
- [ ] **Machine Learning** - Recommandations intelligentes
- [ ] **Multi-tenant** - Support multi-restaurants

---

## 🎓 Contexte académique

> **Projet développé dans le cadre de la formation NWS (Next Web School)**

Ce projet répond aux consignes spécifiques du restaurant **Hótwings** pour développer son activité de livraison avec une API permettant de proposer une carte étendue de recettes scrapées depuis AllRecipes.com.

### Consignes respectées

✅ **API REST** complète avec endpoints CRUD  
✅ **Base de données** MongoDB (NoSQL)  
✅ **Scraper performant** avec goroutines parallèles  
✅ **Swagger** intégré pour la documentation  
✅ **Import JSON** des données scrapées  
✅ **Tests complets** avec couverture de code  
✅ **Docker** pour la containerisation  
✅ **CI/CD** automatisé avec GitHub Actions  

### Liens du projet

[![Golang scrapper_go](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/le-veilleur/scrapper_go)
[![API_golang_Mysql](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/le-veilleur/go_api__scrapper_mysql_docker)

---

<div align="center">

**Développé avec ❤️ par [Maxime Louis](https://github.com/le-veilleur)**

[![GitHub](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/le-veilleur)
[![LinkedIn](https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white)](https://linkedin.com/in/le-veilleur)

</div>