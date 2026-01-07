#!/bin/bash

# Script de test des métriques de l'API Go MongoDB Scrapper
# Ce script démontre les nouvelles fonctionnalités de logging et métriques

echo "🚀 Test des métriques de l'API Go MongoDB Scrapper"
echo "=================================================="

# Configuration
API_URL="http://localhost:8080"
LOG_FILE="api_metrics_test.log"

# Fonction pour faire des requêtes et afficher les résultats
make_request() {
    local method=$1
    local endpoint=$2
    local description=$3
    
    echo ""
    echo "📡 $description"
    echo "   $method $endpoint"
    
    if [ "$method" = "GET" ]; then
        curl -s -w "\n   Status: %{http_code} | Time: %{time_total}s\n" \
             "$API_URL$endpoint" | head -5
    else
        curl -s -w "\n   Status: %{http_code} | Time: %{time_total}s\n" \
             -X "$method" "$API_URL$endpoint" | head -5
    fi
    
    echo "   ---"
}

# Fonction pour attendre un peu entre les requêtes
wait_between_requests() {
    echo "⏳ Attente de 2 secondes..."
    sleep 2
}

echo ""
echo "🔍 1. Test du health check"
make_request "GET" "/health" "Vérification de l'état de l'API"

wait_between_requests

echo ""
echo "📊 2. Test des métriques (avant les requêtes)"
make_request "GET" "/metrics" "Récupération des métriques initiales"

wait_between_requests

echo ""
echo "📋 3. Test de récupération de toutes les recettes"
make_request "GET" "/recettes" "Liste de toutes les recettes"

wait_between_requests

echo ""
echo "🔍 4. Test de recherche par ingrédient"
make_request "GET" "/recette/ingredient/cup" "Recherche par ingrédient 'cup'"

wait_between_requests

echo ""
echo "📊 5. Test des métriques (après quelques requêtes)"
make_request "GET" "/metrics" "Métriques après les requêtes"

wait_between_requests

echo ""
echo "🔄 6. Test du scraper"
make_request "POST" "/scraper/run" "Lancement du scraper"

wait_between_requests

echo ""
echo "📊 7. Test des métriques finales"
make_request "GET" "/metrics" "Métriques finales"

echo ""
echo "✅ Test terminé !"
echo ""
echo "📝 Consultez les logs de l'API pour voir :"
echo "   - Les logs structurés JSON de chaque requête"
echo "   - Les métriques de performance en temps réel"
echo "   - Les logs d'opérations de base de données"
echo "   - Les métriques système (mémoire, goroutines, etc.)"
echo ""
echo "🔗 Endpoints disponibles :"
echo "   - Health: $API_URL/health"
echo "   - Version: $API_URL/version"
echo "   - Métriques: $API_URL/metrics"
echo "   - Recettes: $API_URL/recettes"
echo "   - Scraper: $API_URL/scraper/run"
