package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Variables globales pour le logging dans un fichier
var (
	logFile   *os.File
	logMutex  sync.Mutex
	logInited bool
)

// initLogger initialise le système de logging vers un fichier unique
func initLogger() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logInited {
		return nil
	}

	// Nom du fichier de log fixe
	logFilename := "scraper.log"

	var err error
	// Ouvrir en mode append pour ne pas écraser les logs précédents
	logFile, err = os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ouverture du fichier de log: %v", err)
	}

	// Écrire à la fois dans le fichier ET dans stdout (pour Docker)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	// Ajouter un séparateur pour indiquer le début d'une nouvelle exécution
	separator := strings.Repeat("=", 80)
	log.Printf("\n%s\n", separator)
	log.Printf("🚀 NOUVELLE EXÉCUTION - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("%s\n\n", separator)

	logInited = true
	return nil
}

// closeLogger ferme le fichier de log
func closeLogger() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		logFile.Close()
		logFile = nil
		logInited = false
	}
}

// Fonctions de logging avec variables dynamiques

// logInfo enregistre un message d'information
func logInfo(format string, args ...interface{}) {
	if !logInited {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()
	log.Printf(format, args...)
}

// logConfig enregistre un message de configuration
func logConfig(message string) {
	logInfo("⏳ %s\n", message)
}

// logRequest enregistre une requête HTTP
func logRequest(url string, total int64) {
	logInfo("🌐 Requête principale vers %s (Total: %d) - Délai de 100ms appliqué...\n", url, total)
}

// logResponse enregistre une réponse HTTP
func logResponse(url string, duration time.Duration, size int) {
	logInfo("✅ Réponse reçue en %v pour %s (Taille: %d bytes)\n", duration, url, size)
}

// logRecipeFound enregistre une recette trouvée
func logRecipeFound(recipeNum int64, title string) {
	logInfo("📝 Recette #%d ajoutée à la queue: '%s'\n", recipeNum, title)
}

// logRecipeQueueFull enregistre un avertissement de queue pleine
func logRecipeQueueFull(title string) {
	logInfo("⚠️  Channel plein, recette ignorée: '%s'\n", title)
}

// logPagination enregistre une page de pagination
func logPagination(category string, pageNum, maxPages int, url string) {
	logInfo("📄 Page suivante trouvée pour %s (page %d/%d): %s\n", category, pageNum, maxPages, url)
}

// logPaginationDelay enregistre le délai de pagination
func logPaginationDelay() {
	logInfo("⏳ Pause de 500ms avant la page suivante (respect du serveur et évite le rate limiting)...")
}

// logPaginationLimit enregistre la limite de pagination atteinte
func logPaginationLimit(category string, maxPages int) {
	logInfo("✅ Limite de pages atteinte pour %s (%d pages)\n", category, maxPages)
}

// logRecipeRequest enregistre une requête de recette
func logRecipeRequest(url string, total int64) {
	logInfo("🔍 Requête recette vers %s (Total: %d) - Délai de 50ms appliqué...\n", url, total)
}

// logIngredientsFound enregistre les ingrédients trouvés
func logIngredientsFound(count int, recipeName string) {
	logInfo("🔍 Ingrédients trouvés: %d pour '%s'\n", count, recipeName)
}

// logInstructionsFound enregistre les instructions trouvées
func logInstructionsFound(count int, recipeName string) {
	logInfo("🔍 Instructions trouvées: %d pour '%s'\n", count, recipeName)
}

// logRecipeCompleted enregistre une recette complétée
func logRecipeCompleted(recipeNum int64, recipeName string) {
	logInfo("✅ Recette #%d complétée: '%s'\n", recipeNum, recipeName)
}

// logWorkerStart enregistre le démarrage d'un worker
func logWorkerStart(workerID int, recipeTitle string) {
	logInfo("🚀 Worker #%d démarre le traitement de: %s\n", workerID, recipeTitle)
}

// logWorkerSteps enregistre les étapes du worker
func logWorkerSteps() {
	logInfo("   ⏳ Étapes: 1) Requête HTTP (50ms délai) → 2) Parsing HTML → 3) Extraction données")
}

// logWorkerHTTPComplete enregistre la fin de la requête HTTP
func logWorkerHTTPComplete(duration time.Duration) {
	logInfo("   ✅ Requête HTTP terminée en %v (délai inclus)\n", duration)
}

// logWorkerComplete enregistre la fin du traitement d'un worker
func logWorkerComplete(workerID int, totalDuration, httpDuration time.Duration, recipeTitle string) {
	logInfo("⏱️  Worker #%d terminé en %v (HTTP: %v, Parsing: %v): %s\n",
		workerID, totalDuration, httpDuration, totalDuration-httpDuration, recipeTitle)
}

// logWorkerError enregistre une erreur de worker
func logWorkerError(workerID int, recipeTitle string, err error) {
	logInfo("❌ Worker #%d - Erreur lors de la visite de la page de recette '%s': %v\n", workerID, recipeTitle, err)
}

// logWorkerQueue enregistre la taille de la queue
func logWorkerQueue(workerID int, queueLength int) {
	if queueLength > 0 {
		logInfo("📊 Worker #%d - Queue: %d recettes en attente\n", workerID, queueLength)
	}
}

// logWorkerInit enregistre l'initialisation des workers
func logWorkerInit(count int) {
	logInfo("🏭 Initialisation de %d workers pour le traitement des recettes\n", count)
}

// logWorkerStarted enregistre le démarrage d'un worker
func logWorkerStarted(workerID int) {
	logInfo("🚀 Worker #%d démarré\n", workerID)
}

// logWorkersReady enregistre que les workers sont prêts
func logWorkersReady(count int) {
	logInfo("📊 %d workers réutilisables démarrés et prêts à traiter les recettes\n", count)
}

// logWorkerFinished enregistre la fin d'un worker
func logWorkerFinished(workerID int, requests, recipes int64, duration time.Duration) {
	logInfo("🏁 Worker #%d terminé: %d requêtes, %d recettes, %v\n",
		workerID, requests, recipes, duration)
}

// logAllWorkersFinished enregistre que tous les workers ont terminé
func logAllWorkersFinished(count int) {
	logInfo("🏁 Tous les %d workers ont terminé\n", count)
}

// logCategoryStart enregistre le début du scraping d'une catégorie
func logCategoryStart(categoryNum, totalCategories int, url string) {
	logInfo("🌐 Scraping catégorie %d/%d: %s\n", categoryNum, totalCategories, url)
}

// logCategoryInfo enregistre les informations sur une catégorie
func logCategoryInfo(maxPages, maxRecipesPerPage int) {
	logInfo("   ⏳ Cette catégorie va prendre du temps car:\n")
	logInfo("      - %d pages à visiter (100ms délai entre chaque)\n", maxPages)
	logInfo("      - ~%d recettes par page à traiter (50ms délai par recette)\n", maxRecipesPerPage)
	logInfo("      - Parsing HTML pour chaque page et recette")
}

// logCategoryComplete enregistre la fin d'une catégorie
func logCategoryComplete(categoryNum, totalCategories int, duration time.Duration) {
	logInfo("   ✅ Catégorie %d/%d terminée en %v\n", categoryNum, totalCategories, duration)
}

// logCategoryPause enregistre la pause entre catégories
func logCategoryPause() {
	logInfo("⏳ Pause de 1 seconde entre les catégories (respect du serveur)...")
}

// logCategoryError enregistre une erreur de catégorie
func logCategoryError(url string, err error) {
	logInfo("⚠️  Erreur lors de la visite de la catégorie %s: %v\n", url, err)
}

// logCategoryPhaseComplete enregistre la fin de la phase de collecte
func logCategoryPhaseComplete(duration time.Duration) {
	logInfo("✅ Phase de collecte des catégories terminée en %v\n", duration)
}

// logScrapingStart enregistre le début du scraping
func logScrapingStart(categoryCount int) {
	logInfo("Début du scraping de %d catégories...\n", categoryCount)
}

// logScrapingEstimate enregistre l'estimation du temps
func logScrapingEstimate(pages, recipes int, minSeconds int) {
	logInfo("⏳ Estimation: ~%d pages × 100ms délai + ~%d recettes × 50ms délai = ~%d secondes minimum\n",
		pages, recipes, minSeconds)
}

// logProcessingPhase enregistre le début de la phase de traitement
func logProcessingPhase(found, completed, inProgress int64) {
	logInfo("📊 Phase de traitement des recettes:\n")
	logInfo("   - %d recettes trouvées, %d complétées, %d en cours de traitement\n",
		found, completed, inProgress)
}

// logProcessingEstimate enregistre l'estimation du temps restant
func logProcessingEstimate(remaining int64, estimatedTime time.Duration) {
	if remaining > 0 {
		logInfo("   ⏳ Temps estimé restant: ~%v (basé sur %d recettes × ~110ms)\n",
			estimatedTime, remaining)
	}
}

// logProcessingClose enregistre la fermeture de la queue
func logProcessingClose() {
	logInfo("⏳ Fermeture de la queue et attente de la fin du traitement des workers...")
}

// logProcessingComplete enregistre la fin du traitement
func logProcessingComplete() {
	logInfo("✅ Tous les workers ont terminé le traitement des recettes")
}

// logSaveStart enregistre le début de la sauvegarde
func logSaveStart(count int, filename string) {
	logInfo("💾 Sauvegarde de %d recettes dans %s...\n", count, filename)
}

// logSaveComplete enregistre la fin de la sauvegarde
func logSaveComplete(duration time.Duration) {
	logInfo("✅ Sauvegarde terminée en %v\n", duration)
}

// logSaveError enregistre une erreur de sauvegarde
func logSaveError(err error) {
	logInfo("Erreur lors de l'enregistrement des recettes: %v\n", err)
}

// logVersionPrint enregistre les informations de version (pour printVersionInfo)
func logVersionPrint(version, gitCommit, buildTime, goVersion, os, arch string) {
	logInfo("Go MongoDB Scrapper\n")
	logInfo("Version: %s\n", version)
	logInfo("Git Commit: %s\n", gitCommit)
	logInfo("Build Time: %s\n", buildTime)
	logInfo("Go Version: %s\n", goVersion)
	logInfo("OS/Arch: %s/%s\n\n", os, arch)
}

// logDetailedStatsPerformance enregistre les performances générales
func logDetailedStatsPerformance(totalDuration time.Duration, requestsPerSec, recipesPerSec float64) {
	logInfo("⏱️  Durée totale: %v\n", totalDuration)
	logInfo("🚀 Requêtes par seconde: %.2f\n", requestsPerSec)
	logInfo("📝 Recettes par seconde: %.2f\n", recipesPerSec)
}

// logDetailedStatsRequests enregistre les statistiques de requêtes
func logDetailedStatsRequests(total, mainPage, recipe int64) {
	logInfo("\n🌐 REQUÊTES:\n")
	logInfo("   Total: %d\n", total)
	logInfo("   Page principale: %d\n", mainPage)
	logInfo("   Pages recettes: %d\n", recipe)
}

// logDetailedStatsRecipes enregistre les statistiques de recettes
func logDetailedStatsRecipes(found, completed, failed int64, successRate float64) {
	logInfo("\n📝 RECETTES:\n")
	logInfo("   Trouvées: %d\n", found)
	logInfo("   Complétées: %d\n", completed)
	logInfo("   Échouées: %d\n", failed)
	logInfo("   Taux de succès: %.1f%%\n", successRate)
}

// logDetailedStatsConfig enregistre la configuration automatique
func logDetailedStatsConfig(logicalCPU, physicalCores, adaptiveRatio, calculatedWorkers, finalWorkers int) {
	logInfo("\n💻 CONFIGURATION AUTOMATIQUE:\n")
	logInfo("   Processeurs logiques: %d\n", logicalCPU)
	logInfo("   Cœurs physiques détectés: %d\n", physicalCores)
	logInfo("   Ratio adaptatif: %d (calculé automatiquement)\n", adaptiveRatio)
	logInfo("   Calcul: %d cœurs × %d = %d workers\n", physicalCores, adaptiveRatio, calculatedWorkers)
	logInfo("   Configuration finale: %d workers\n", finalWorkers)
}

// logDetailedStatsWorker enregistre les stats d'un worker
func logDetailedStatsWorker(workerID int, requests, recipes int64, duration time.Duration) {
	logInfo("   Worker #%d: %d requêtes, %d recettes, %v\n", workerID, requests, recipes, duration)
}

// logDetailedStatsWorkersHeader enregistre l'en-tête des stats par worker
func logDetailedStatsWorkersHeader() {
	logInfo("\n📈 PERFORMANCE PAR WORKER:\n")
}

// logDetailedStatsAnalysis enregistre l'analyse de performance
func logDetailedStatsAnalysis(avgRequestsPerRecipe, requestsPerSec float64, avgTimePerRecipe float64) {
	logInfo("\n💡 ANALYSE DE PERFORMANCE:\n")
	logInfo("   Requêtes moyennes par recette: %.1f\n", avgRequestsPerRecipe)
	logInfo("   Débit estimé: %.0f requêtes/seconde\n", requestsPerSec)
	if avgTimePerRecipe > 0 {
		logInfo("   Temps moyen par recette: %.2f secondes\n", avgTimePerRecipe)
	}
}

// logDetailedStatsFooter enregistre le pied de page des statistiques
func logDetailedStatsFooter(filename string) {
	logInfo("\n💾 Fichier de sortie: %s\n", filename)
	logInfo("%s\n", strings.Repeat("=", 80))
}
