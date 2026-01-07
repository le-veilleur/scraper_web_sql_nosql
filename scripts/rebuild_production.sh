#!/bin/bash

# Script pour reconstruire les images Docker en production sur le VPS

set -e

echo "🔄 Reconstruction des images Docker en production..."

# Aller dans le dossier du projet
cd ~/GitHub/scraper_web_sql_nosql || exit 1

# Récupérer les dernières modifications
echo "📥 Récupération des dernières modifications..."
git pull origin main

# Reconstruire les images avec la version production
echo "🔨 Reconstruction des images Docker (version: production)..."
docker-compose build --no-cache

# Redémarrer les services
echo "🚀 Redémarrage des services..."
docker-compose down
docker-compose up -d

# Nettoyer les anciennes images
echo "🧹 Nettoyage des anciennes images..."
docker image prune -f

echo "✅ Reconstruction terminée !"
echo ""
echo "📊 Vérification des services :"
docker-compose ps

echo ""
echo "📝 Logs du scraper (dernières lignes) :"
docker-compose logs --tail=20 scraper

