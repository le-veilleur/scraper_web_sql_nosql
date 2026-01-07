#!/bin/bash

# Script de test des métriques détaillées du scraper
# Ce script démontre les nouvelles fonctionnalités de métriques avancées

echo "🚀 Test des métriques détaillées du scraper Go MongoDB"
echo "====================================================="

# Configuration
SCRAPER_DIR="/home/maka/GitHub/go_api_mongo_scrapper/scraper"
OUTPUT_FILE="scraper_metrics_test.log"

echo ""
echo "📋 Configuration du test :"
echo "   - Répertoire scraper: $SCRAPER_DIR"
echo "   - Fichier de sortie: $OUTPUT_FILE"
echo "   - Métriques en temps réel toutes les 5 secondes"
echo "   - Statistiques détaillées à la fin"
echo ""

# Vérifier que le scraper existe
if [ ! -f "$SCRAPER_DIR/scraper" ]; then
    echo "🔨 Compilation du scraper..."
    cd "$SCRAPER_DIR"
    go build -o scraper scraper.go
    if [ $? -ne 0 ]; then
        echo "❌ Erreur lors de la compilation du scraper"
        exit 1
    fi
    echo "✅ Scraper compilé avec succès"
else
    echo "✅ Binaire scraper trouvé"
fi

echo ""
echo "🧪 Lancement des tests unitaires..."
cd "$SCRAPER_DIR"
go test -v -run "TestScrapingStats|TestWorkerStats|TestCalculateFinalStats|TestGetDetailedStats"
if [ $? -ne 0 ]; then
    echo "❌ Certains tests ont échoué"
    exit 1
fi
echo "✅ Tous les tests sont passés"

echo ""
echo "🏃 Lancement du scraper avec métriques détaillées..."
echo "   (Appuyez sur Ctrl+C pour arrêter prématurément)"
echo ""

# Lancer le scraper et capturer la sortie
cd "$SCRAPER_DIR"
timeout 60s ./scraper 2>&1 | tee "$OUTPUT_FILE"

echo ""
echo "📊 Analyse des métriques collectées..."

if [ -f "$OUTPUT_FILE" ]; then
    echo ""
    echo "📈 RÉSUMÉ DES MÉTRIQUES :"
    echo "========================"
    
    # Extraire les métriques principales
    echo "🌐 Requêtes :"
    grep -o "Total: [0-9]*" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée de requêtes trouvée"
    grep -o "Page principale: [0-9]*" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée de page principale trouvée"
    grep -o "Pages recettes: [0-9]*" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée de pages recettes trouvée"
    
    echo ""
    echo "📝 Recettes :"
    grep -o "Trouvées: [0-9]*" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée de recettes trouvées"
    grep -o "Complétées: [0-9]*" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée de recettes complétées trouvée"
    grep -o "Échouées: [0-9]*" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée de recettes échouées trouvée"
    
    echo ""
    echo "🏭 Workers :"
    grep -o "Workers utilisés: [0-9]*" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée de workers trouvée"
    
    echo ""
    echo "⚡ Performance :"
    grep -o "Requêtes par seconde: [0-9.]*" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée de performance trouvée"
    grep -o "Recettes par seconde: [0-9.]*" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée de performance trouvée"
    
    echo ""
    echo "💡 Analyse de performance :"
    grep -o "Requêtes moyennes par recette: [0-9.]*" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée d'analyse trouvée"
    grep -o "Débit estimé: [0-9]* requêtes/seconde" "$OUTPUT_FILE" | tail -1 || echo "   Aucune donnée de débit trouvée"
    
    echo ""
    echo "📈 PERFORMANCE PAR WORKER :"
    echo "==========================="
    grep "Worker #[0-9]*:" "$OUTPUT_FILE" | tail -10 || echo "   Aucune donnée de workers trouvée"
    
    echo ""
    echo "📊 MÉTRIQUES EN TEMPS RÉEL :"
    echo "============================"
    grep "📊 \[.*\] Req:" "$OUTPUT_FILE" | tail -5 || echo "   Aucune donnée de temps réel trouvée"
    
    echo ""
    echo "🔍 LOGS DÉTAILLÉS :"
    echo "==================="
    echo "Dernières 10 lignes de logs :"
    tail -10 "$OUTPUT_FILE"
    
else
    echo "❌ Aucun fichier de sortie trouvé"
fi

echo ""
echo "✅ Test terminé !"
echo ""
echo "📁 Fichiers générés :"
echo "   - Logs détaillés: $OUTPUT_FILE"
echo "   - Données JSON: $SCRAPER_DIR/data.json"
echo ""
echo "🔗 Pour relancer le test :"
echo "   ./scripts/test_scraper_detailed_metrics.sh"
echo ""
echo "📖 Pour voir les métriques en continu :"
echo "   cd $SCRAPER_DIR && ./scraper"
