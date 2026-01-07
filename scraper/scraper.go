package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

// Variables de versioning injectées lors du build
// Ces valeurs sont remplacées par les flags de compilation lors du build Docker
var (
	version   = "dev"     // Version de l'application
	gitCommit = "unknown" // Hash du commit Git
	buildTime = "unknown" // Timestamp de compilation
)

// BuildInfo supprimé - non utilisé après réduction des logs

// Recipe représente une recette complète avec tous ses détails
type Recipe struct {
	Name         string        `json:"name"`         // Nom de la recette
	Page         string        `json:"page"`         // URL de la page de la recette
	Image        string        `json:"image"`        // URL de l'image de la recette
	Ingredients  []Ingredient  `json:"ingredients"`  // Liste des ingrédients
	Instructions []Instruction `json:"instructions"` // Liste des instructions
}

// Ingredient représente un ingrédient avec sa quantité et son unité
type Ingredient struct {
	Quantity string `json:"quantity"` // Quantité (ex: "2", "1/2")
	Unit     string `json:"unit"`     // Unité (ex: "cups", "tablespoons")
}

// Instruction représente une étape de la recette
type Instruction struct {
	Number      string `json:"number"`      // Numéro de l'étape (ex: "1", "2")
	Description string `json:"description"` // Description de l'étape
}

// RecipeData contient les informations de base d'une recette avant le scraping détaillé
// Utilisé pour passer les données entre les goroutines
type RecipeData struct {
	URL   string // URL de la page de la recette
	Title string // Titre de la recette
	Image string // URL de l'image de la recette
}

// ScrapingStats contient toutes les statistiques de performance du scraper
// Thread-safe grâce au Mutex pour les accès concurrents
type ScrapingStats struct {
	// Compteurs de requêtes HTTP
	TotalRequests    int64 `json:"total_requests"`     // Total des requêtes HTTP
	MainPageRequests int64 `json:"main_page_requests"` // Requêtes vers les pages de catégories
	RecipeRequests   int64 `json:"recipe_requests"`    // Requêtes vers les pages de recettes

	// Compteurs de recettes
	RecipesFound     int64 `json:"recipes_found"`     // Nombre de recettes découvertes
	RecipesCompleted int64 `json:"recipes_completed"` // Nombre de recettes traitées avec succès
	RecipesFailed    int64 `json:"recipes_failed"`    // Nombre de recettes en échec

	// Métriques de performance temporelles
	StartTime         time.Time     `json:"start_time"`          // Heure de début du scraping
	EndTime           time.Time     `json:"end_time"`            // Heure de fin du scraping
	TotalDuration     time.Duration `json:"total_duration"`      // Durée totale du scraping
	RequestsPerSecond float64       `json:"requests_per_second"` // Requêtes par seconde
	RecipesPerSecond  float64       `json:"recipes_per_second"`  // Recettes par seconde

	// Configuration des workers
	MaxWorkers    int   `json:"max_workers"`    // Nombre maximum de workers
	ActiveWorkers int64 `json:"active_workers"` // Nombre de workers actifs

	// Statistiques détaillées par worker
	WorkerStats map[int]WorkerStats `json:"worker_stats"` // Map des stats par worker

	Mutex sync.RWMutex // Mutex pour la sécurité des accès concurrents
}

// WorkerStats contient les statistiques d'un worker individuel
type WorkerStats struct {
	WorkerID         int           `json:"worker_id"`         // ID unique du worker
	RequestsHandled  int64         `json:"requests_handled"`  // Nombre de requêtes traitées
	RecipesProcessed int64         `json:"recipes_processed"` // Nombre de recettes traitées
	StartTime        time.Time     `json:"start_time"`        // Heure de démarrage du worker
	EndTime          time.Time     `json:"end_time"`          // Heure de fin du worker
	Duration         time.Duration `json:"duration"`          // Durée totale d'activité
}

// NewScrapingStats crée une nouvelle instance de ScrapingStats
// maxWorkers: nombre maximum de workers qui seront utilisés
func NewScrapingStats(maxWorkers int) *ScrapingStats {
	return &ScrapingStats{
		StartTime:   time.Now(),                // Initialiser avec l'heure actuelle
		MaxWorkers:  maxWorkers,                // Stocker le nombre max de workers
		WorkerStats: make(map[int]WorkerStats), // Initialiser la map des stats par worker
	}
}

// IncrementMainPageRequest incrémente le compteur de requêtes vers les pages principales
// Thread-safe grâce au mutex
func (s *ScrapingStats) IncrementMainPageRequest() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.TotalRequests++    // Incrémenter le total des requêtes
	s.MainPageRequests++ // Incrémenter les requêtes vers les pages principales
}

// IncrementRecipeRequest incrémente le compteur de requêtes vers les pages de recettes
// Thread-safe grâce au mutex
func (s *ScrapingStats) IncrementRecipeRequest() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.TotalRequests++  // Incrémenter le total des requêtes
	s.RecipeRequests++ // Incrémenter les requêtes vers les recettes
}

// IncrementRecipesFound incrémente le compteur de recettes découvertes
// Thread-safe grâce au mutex
func (s *ScrapingStats) IncrementRecipesFound() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.RecipesFound++ // Incrémenter le nombre de recettes trouvées
}

// IncrementRecipesCompleted incrémente le compteur de recettes traitées avec succès
// Thread-safe grâce au mutex
func (s *ScrapingStats) IncrementRecipesCompleted() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.RecipesCompleted++ // Incrémenter le nombre de recettes complétées
}

// IncrementRecipesFailed incrémente le compteur de recettes en échec
// Thread-safe grâce au mutex
func (s *ScrapingStats) IncrementRecipesFailed() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.RecipesFailed++ // Incrémenter le nombre de recettes échouées
}

func (s *ScrapingStats) UpdateWorkerStats(workerID int, requests, recipes int64) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if worker, exists := s.WorkerStats[workerID]; exists {
		worker.RequestsHandled += requests
		worker.RecipesProcessed += recipes
		worker.EndTime = time.Now()
		worker.Duration = worker.EndTime.Sub(worker.StartTime)
		s.WorkerStats[workerID] = worker
	} else {
		s.WorkerStats[workerID] = WorkerStats{
			WorkerID:         workerID,
			RequestsHandled:  requests,
			RecipesProcessed: recipes,
			StartTime:        time.Now(),
			EndTime:          time.Now(),
			Duration:         0,
		}
	}
}

func (s *ScrapingStats) GetTotalRequests() int64 {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.TotalRequests
}

func (s *ScrapingStats) CalculateFinalStats() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.EndTime = time.Now()
	s.TotalDuration = s.EndTime.Sub(s.StartTime)

	if s.TotalDuration.Seconds() > 0 {
		s.RequestsPerSecond = float64(s.TotalRequests) / s.TotalDuration.Seconds()
		s.RecipesPerSecond = float64(s.RecipesCompleted) / s.TotalDuration.Seconds()
	}
}

func (s *ScrapingStats) GetDetailedStats() ScrapingStats {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	// Créer une copie sans le mutex
	return ScrapingStats{
		TotalRequests:     s.TotalRequests,
		MainPageRequests:  s.MainPageRequests,
		RecipeRequests:    s.RecipeRequests,
		RecipesFound:      s.RecipesFound,
		RecipesCompleted:  s.RecipesCompleted,
		RecipesFailed:     s.RecipesFailed,
		StartTime:         s.StartTime,
		EndTime:           s.EndTime,
		TotalDuration:     s.TotalDuration,
		RequestsPerSecond: s.RequestsPerSecond,
		RecipesPerSecond:  s.RecipesPerSecond,
		MaxWorkers:        s.MaxWorkers,
		ActiveWorkers:     s.ActiveWorkers,
		WorkerStats:       s.WorkerStats,
	}
}

// getPhysicalCores détecte le vrai nombre de cœurs physiques
func getPhysicalCores() int {
	// Méthode 1: Lire /proc/cpuinfo sur Linux
	if runtime.GOOS == "linux" {
		if cores := detectPhysicalCoresFromProc(); cores > 0 {
			return cores
		}
	}

	// Méthode 2: Estimation intelligente basée sur les patterns courants
	numLogicalCPU := runtime.NumCPU()

	// Patterns courants d'hyperthreading
	switch {
	case numLogicalCPU == 1:
		return 1
	case numLogicalCPU == 2:
		return 2 // Probablement 2 cœurs sans HT
	case numLogicalCPU == 4:
		return 2 // Probablement 2 cœurs avec HT
	case numLogicalCPU == 6:
		return 6 // Probablement 6 cœurs sans HT
	case numLogicalCPU == 8:
		return 4 // Probablement 4 cœurs avec HT
	case numLogicalCPU == 12:
		return 6 // Probablement 6 cœurs avec HT
	case numLogicalCPU == 16:
		return 8 // Probablement 8 cœurs avec HT
	case numLogicalCPU == 24:
		return 12 // Probablement 12 cœurs avec HT
	case numLogicalCPU == 32:
		return 16 // Probablement 16 cœurs avec HT
	case numLogicalCPU%2 == 0:
		// Si pair, essayer de diviser par 2 (hyperthreading probable)
		estimated := numLogicalCPU / 2
		if estimated >= 1 {
			return estimated
		}
	}

	// Fallback: utiliser le nombre logique
	return numLogicalCPU
}

// detectPhysicalCoresFromProc lit /proc/cpuinfo pour détecter les vrais cœurs physiques
func detectPhysicalCoresFromProc() int {
	// Cette fonction serait implémentée pour lire /proc/cpuinfo
	// et compter les vrais cœurs physiques
	// Pour l'instant, on retourne 0 pour utiliser la méthode de fallback
	return 0
}

// calculateAdaptiveRatio calcule le ratio optimal basé sur le nombre de cœurs
func calculateAdaptiveRatio(numCores int) int {
	switch {
	case numCores <= 2:
		return 3 // Plus de workers sur machines faibles pour compenser
	case numCores <= 4:
		return 2 // Ratio standard pour machines moyennes
	case numCores <= 8:
		return 2 // Ratio standard pour machines puissantes
	case numCores <= 16:
		return 1 // Moins de workers sur très grosses machines (éviter la surcharge)
	default:
		return 1 // Ratio conservateur pour machines extrêmes
	}
}

// calculateOptimalWorkers calcule le nombre optimal de workers basé sur les ressources CPU
// minWorkers: nombre minimum de workers (par défaut 1)
// maxWorkers: nombre maximum de workers (par défaut 50)
func calculateOptimalWorkers(minWorkers, maxWorkers int) int {
	// Détecter le vrai nombre de cœurs physiques
	numPhysicalCores := getPhysicalCores()

	// Calculer le ratio adaptatif basé sur le nombre de cœurs
	adaptiveRatio := calculateAdaptiveRatio(numPhysicalCores)

	optimalWorkers := numPhysicalCores * adaptiveRatio

	// Appliquer les limites
	if optimalWorkers < minWorkers {
		optimalWorkers = minWorkers
	}
	if optimalWorkers > maxWorkers {
		optimalWorkers = maxWorkers
	}

	return optimalWorkers
}

// printVersionInfo affiche les informations de version (simplifié)
func printVersionInfo() {
	// Logs de version supprimés pour réduire la verbosité
}

// getBuildInfo supprimé - non utilisé après réduction des logs

// userAgents contient une liste de User-Agents réalistes pour simuler différents navigateurs
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
}

var userAgentMutex sync.Mutex
var userAgentIndex = 0

// getRandomUserAgent retourne un User-Agent aléatoire de la liste
func getRandomUserAgent() string {
	userAgentMutex.Lock()
	defer userAgentMutex.Unlock()

	// Utiliser un index rotatif pour distribuer les User-Agents
	userAgentIndex = (userAgentIndex + 1) % len(userAgents)
	return userAgents[userAgentIndex]
}

// configureRealisticHeaders configure les headers HTTP pour simuler un navigateur réel
func configureRealisticHeaders(r *colly.Request) {
	// User-Agent réaliste
	r.Headers.Set("User-Agent", getRandomUserAgent())

	// Headers standards d'un navigateur moderne
	r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	r.Headers.Set("Accept-Language", "en-US,en;q=0.9,fr;q=0.8")
	r.Headers.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	r.Headers.Set("DNT", "1")
	r.Headers.Set("Connection", "keep-alive")
	r.Headers.Set("Upgrade-Insecure-Requests", "1")
	r.Headers.Set("Sec-Fetch-Dest", "document")
	r.Headers.Set("Sec-Fetch-Mode", "navigate")
	r.Headers.Set("Sec-Fetch-Site", "none")
	r.Headers.Set("Sec-Fetch-User", "?1")
	r.Headers.Set("Cache-Control", "max-age=0")

	// Headers sec-ch-ua pour simuler un navigateur moderne (Chrome/Edge)
	r.Headers.Set("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	r.Headers.Set("sec-ch-ua-mobile", "?0")
	r.Headers.Set("sec-ch-ua-platform", `"Windows"`)

	// Ajouter un Referer réaliste
	if r.URL != nil && r.URL.Host != "" {
		// Pour la première visite, utiliser Google comme referer
		if !strings.Contains(r.URL.String(), "allrecipes.com") || r.URL.Path == "/" {
			r.Headers.Set("Referer", "https://www.google.com/")
		} else {
			// Pour les pages internes, utiliser le domaine comme referer
			r.Headers.Set("Referer", "https://www.allrecipes.com/")
		}
	} else {
		// Referer par défaut pour la première visite
		r.Headers.Set("Referer", "https://www.google.com/")
	}
}

// getRandomDelay retourne un délai aléatoire entre min et max millisecondes
func getRandomDelay(minMs, maxMs int) time.Duration {
	if maxMs <= minMs {
		return time.Duration(minMs) * time.Millisecond
	}
	delay := minMs + rand.Intn(maxMs-minMs+1)
	return time.Duration(delay) * time.Millisecond
}

// createMainCollector crée et configure le collecteur principal pour les pages de catégories
// Ce collecteur visite les pages de listes de recettes et extrait les URLs des recettes individuelles
func createMainCollector(stats *ScrapingStats, recipeURLs chan<- RecipeData) *colly.Collector {
	collector := colly.NewCollector()

	// Configuration des limites pour être respectueux du serveur
	// Délais augmentés et parallélisme réduit pour éviter la détection
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",                    // Appliquer à tous les domaines
		Parallelism: 3,                      // Réduit à 3 requêtes simultanées
		Delay:       500 * time.Millisecond, // Délai de base de 500ms entre les requêtes
		RandomDelay: 1 * time.Second,        // Délai aléatoire jusqu'à 1 seconde (fonctionnalité native Colly)
	})

	// Handler appelé avant chaque requête HTTP
	collector.OnRequest(func(r *colly.Request) {
		// Configurer les headers réalistes pour éviter la détection
		configureRealisticHeaders(r)

		// Les délais aléatoires sont gérés automatiquement par Colly via RandomDelay dans LimitRule
		stats.IncrementMainPageRequest() // Incrémenter le compteur de requêtes
		// Log de requête supprimé pour réduire la verbosité
	})

	// Gérer les erreurs HTTP (403, 429, etc.)
	collector.OnError(func(r *colly.Response, err error) {
		statusCode := r.StatusCode
		if statusCode == 403 || statusCode == 429 {
			log.Printf("⚠️  Erreur %d détectée pour %s: %v\n", statusCode, r.Request.URL, err)
			log.Printf("🔄 Attente prolongée avant retry (10-20s)...\n")
			// Attendre beaucoup plus longtemps en cas d'erreur (10-20 secondes)
			time.Sleep(getRandomDelay(10000, 20000))
		} else {
			log.Printf("❌ Erreur HTTP %d pour %s: %v\n", statusCode, r.Request.URL, err)
		}
	})

	// Handler appelé pour chaque élément HTML correspondant au sélecteur CSS
	// Ce sélecteur cible les cartes de recettes sur AllRecipes
	collector.OnHTML("div.mntl-taxonomysc-article-list-group .mntl-card", func(e *colly.HTMLElement) {
		// Extraire l'URL, le titre et l'image de la recette
		page := e.Request.AbsoluteURL(e.Attr("href")) // URL de la page de la recette
		title := e.ChildText("span.card__title-text") // Titre de la recette
		image := e.ChildAttr("img", "data-src")       // URL de l'image

		// Vérifier que nous avons les données essentielles
		if page != "" && title != "" {
			stats.IncrementRecipesFound() // Incrémenter le compteur de recettes trouvées

			// Créer l'objet RecipeData avec les informations extraites
			recipeData := RecipeData{
				URL:   page,
				Title: title,
				Image: image,
			}

			// Envoyer la recette dans le channel (non-bloquant)
			select {
			case recipeURLs <- recipeData:
				// Log supprimé pour réduire la verbosité (trop de logs)
			default:
				log.Printf("⚠️  Channel plein, recette ignorée: '%s'\n", title)
			}
		}
	})

	return collector
}

// createMainCollectorWithPagination crée un collecteur avec support de la pagination
func createMainCollectorWithPagination(stats *ScrapingStats, recipeURLs chan<- RecipeData, maxPages int) *colly.Collector {
	collector := colly.NewCollector()

	// Configuration des limites avec délais plus longs pour éviter la détection
	// Parallélisme réduit à 1 pour éviter la détection anti-bot
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,               // Réduit à 1 requête à la fois pour éviter la détection
		Delay:       2 * time.Second, // Délai de base augmenté à 2 secondes
		RandomDelay: 2 * time.Second, // Délai aléatoire jusqu'à 2 secondes (fonctionnalité native Colly)
	})

	// Map pour suivre les pages visitées par catégorie
	visitedPages := make(map[string]int)
	var mutex sync.Mutex

	collector.OnRequest(func(r *colly.Request) {
		// Configurer les headers réalistes pour éviter la détection
		configureRealisticHeaders(r)

		// Les délais aléatoires sont gérés automatiquement par Colly via RandomDelay dans LimitRule
		stats.IncrementMainPageRequest()
		// Log de requête supprimé pour réduire la verbosité
	})

	// Gérer les erreurs HTTP (403, 429, etc.)
	collector.OnError(func(r *colly.Response, err error) {
		statusCode := r.StatusCode
		if statusCode == 403 || statusCode == 429 {
			log.Printf("⚠️  Erreur %d détectée pour %s: %v\n", statusCode, r.Request.URL, err)
			log.Printf("🔄 Attente prolongée avant retry (10-20s)...\n")
			// Attendre beaucoup plus longtemps en cas d'erreur (10-20 secondes)
			time.Sleep(getRandomDelay(10000, 20000))
		} else {
			log.Printf("❌ Erreur HTTP %d pour %s: %v\n", statusCode, r.Request.URL, err)
		}
	})

	// Gérer les recettes sur la page actuelle
	collector.OnHTML("div.mntl-taxonomysc-article-list-group .mntl-card", func(e *colly.HTMLElement) {
		page := e.Request.AbsoluteURL(e.Attr("href"))
		title := e.ChildText("span.card__title-text")
		image := e.ChildAttr("img", "data-src")

		if page != "" && title != "" {
			stats.IncrementRecipesFound()
			recipeData := RecipeData{
				URL:   page,
				Title: title,
				Image: image,
			}

			select {
			case recipeURLs <- recipeData:
				// Log supprimé pour réduire la verbosité (trop de logs)
			default:
				log.Printf("⚠️  Channel plein, recette ignorée: '%s'\n", title)
			}
		}
	})

	// Gérer la pagination
	collector.OnHTML("a[data-testid='pagination-next']", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		if nextPageURL == "" {
			return
		}

		// Extraire la catégorie de base de l'URL actuelle
		baseCategory := e.Request.URL.Path
		if strings.Contains(baseCategory, "?") {
			baseCategory = strings.Split(baseCategory, "?")[0]
		}

		mutex.Lock()
		pagesVisited := visitedPages[baseCategory]
		mutex.Unlock()

		if pagesVisited < maxPages {
			mutex.Lock()
			visitedPages[baseCategory] = pagesVisited + 1
			mutex.Unlock()

			// Log de pagination supprimé pour réduire la verbosité

			// Visiter la page suivante avec un délai aléatoire plus long
			randomDelay := getRandomDelay(2000, 5000) // Délai aléatoire entre 2s et 5s
			time.Sleep(randomDelay)
			collector.Visit(nextPageURL)
		} else {
			log.Printf("✅ Limite de pages atteinte pour %s (%d pages)\n", baseCategory, maxPages)
		}
	})

	return collector
}

// createRecipeCollector crée un collecteur pour collecter une recette individuelle
func createRecipeCollector(stats *ScrapingStats) *colly.Collector {
	collector := colly.NewCollector()

	// Configuration avec délais plus longs pour éviter la détection
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       2 * time.Second, // Délai de base augmenté à 2 secondes
	})

	collector.OnRequest(func(r *colly.Request) {
		// Configurer les headers réalistes pour éviter la détection
		configureRealisticHeaders(r)

		// Les délais aléatoires sont gérés automatiquement par Colly via RandomDelay dans LimitRule
		stats.IncrementRecipeRequest()
		// Log de requête supprimé pour réduire la verbosité
	})

	// Gérer les erreurs HTTP (403, 429, etc.)
	collector.OnError(func(r *colly.Response, err error) {
		statusCode := r.StatusCode
		if statusCode == 403 || statusCode == 429 {
			log.Printf("⚠️  Erreur %d détectée pour la recette %s: %v\n", statusCode, r.Request.URL, err)
			log.Printf("🔄 Attente prolongée avant retry (10-20s)...\n")
			// Attendre beaucoup plus longtemps en cas d'erreur (10-20 secondes)
			time.Sleep(getRandomDelay(10000, 20000))
		} else {
			log.Printf("❌ Erreur HTTP %d pour la recette %s: %v\n", statusCode, r.Request.URL, err)
		}
	})

	return collector
}

// scrapeRecipeDetails configure les handlers pour collecter les détails d'une recette
func scrapeRecipeDetails(collector *colly.Collector, recipe *Recipe, completedRecipes chan<- Recipe, stats *ScrapingStats) {
	// Collecter les ingrédients - Nouveaux sélecteurs CSS pour AllRecipes 2024
	collector.OnHTML("ul.mm-recipes-structured-ingredients__list", func(e *colly.HTMLElement) {
		var ingredients []Ingredient

		e.ForEach("li.mm-recipes-structured-ingredients__list-item", func(_ int, ingr *colly.HTMLElement) {
			// Extraire la quantité et l'unité séparément
			quantity := strings.TrimSpace(ingr.ChildText("span[data-ingredient-quantity=true]"))
			unit := strings.TrimSpace(ingr.ChildText("span[data-ingredient-unit=true]"))
			name := strings.TrimSpace(ingr.ChildText("span[data-ingredient-name=true]"))

			// Si on a des données structurées, les utiliser
			if quantity != "" || unit != "" || name != "" {
				// Construire le texte complet de l'ingrédient
				fullText := strings.TrimSpace(ingr.Text)
				ingredients = append(ingredients, Ingredient{
					Quantity: fullText, // Texte complet pour l'instant
					Unit:     "",       // Pas de séparation pour l'instant
				})
			}
		})

		recipe.Ingredients = ingredients
		// Log d'ingrédients supprimé pour réduire la verbosité
	})

	// Collecter les instructions - Nouveaux sélecteurs CSS pour AllRecipes 2024
	collector.OnHTML("div.mm-recipes-steps__content", func(e *colly.HTMLElement) {
		var instructions []Instruction

		// Chercher dans les listes ordonnées avec la structure correcte
		e.ForEach("ol.mntl-sc-block li", func(i int, inst *colly.HTMLElement) {
			number := strconv.Itoa(i + 1)
			// Extraire le texte de la balise <p> à l'intérieur du <li>
			description := strings.TrimSpace(inst.ChildText("p.mntl-sc-block-html"))
			if description == "" {
				// Fallback sur le texte complet si pas de balise p
				description = strings.TrimSpace(inst.Text)
			}
			if description != "" {
				instructions = append(instructions, Instruction{
					Number:      number,
					Description: description,
				})
			}
		})

		recipe.Instructions = instructions
		// Log d'instructions supprimé pour réduire la verbosité
	})

	// Quand la collecte de la recette est terminée
	collector.OnScraped(func(r *colly.Response) {
		stats.IncrementRecipesCompleted()
		completedRecipes <- *recipe
		// Log de recette complétée supprimé pour réduire la verbosité (trop de logs)
	})
}

// processRecipeReusable traite une recette dans un worker réutilisable
func processRecipeReusable(recipeData RecipeData, stats *ScrapingStats, completedRecipes chan<- Recipe, workerStats *WorkerStats) {
	startTime := time.Now()
	// Log de traitement supprimé pour réduire la verbosité

	// Créer un collecteur dédié pour cette recette
	recipeCollector := createRecipeCollector(stats)

	recipe := Recipe{
		Name:  recipeData.Title,
		Page:  recipeData.URL,
		Image: recipeData.Image,
	}

	// Configurer la collecte des détails
	scrapeRecipeDetails(recipeCollector, &recipe, completedRecipes, stats)

	// Visiter la page de la recette
	err := recipeCollector.Visit(recipeData.URL)
	if err != nil {
		stats.IncrementRecipesFailed()
		log.Printf("❌ Worker #%d - Erreur lors de la visite de la page de recette '%s': %v\n", workerStats.WorkerID, recipeData.Title, err)
	} else {
		// Mettre à jour les stats du worker
		workerStats.RequestsHandled++
		workerStats.RecipesProcessed++
	}

	duration := time.Since(startTime)
	_ = duration // Utilisé pour les stats mais pas loggé pour réduire la verbosité
}

// startRecipeProcessor démarre la goroutine qui traite les URLs de recettes
func startRecipeProcessor(recipeURLs <-chan RecipeData, completedRecipes chan<- Recipe, stats *ScrapingStats, wg *sync.WaitGroup) {
	go func() {
		maxWorkers := stats.MaxWorkers // Utiliser le nombre optimal calculé automatiquement
		semaphore := make(chan struct{}, maxWorkers)

		// Log d'initialisation supprimé pour réduire la verbosité

		// Créer des workers réutilisables
		for i := 0; i < maxWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				workerStats := WorkerStats{
					WorkerID:         workerID,
					RequestsHandled:  0,
					RecipesProcessed: 0,
					StartTime:        time.Now(),
				}

				// Log de démarrage worker supprimé pour réduire la verbosité

				// Le worker traite les recettes en continu
				for recipeData := range recipeURLs {
					// Acquérir un slot dans le semaphore
					semaphore <- struct{}{}

					// Traiter la recette
					processRecipeReusable(recipeData, stats, completedRecipes, &workerStats)

					// Libérer le slot
					<-semaphore
				}

				// Mettre à jour les stats finales du worker
				workerStats.EndTime = time.Now()
				workerStats.Duration = workerStats.EndTime.Sub(workerStats.StartTime)
				stats.Mutex.Lock()
				stats.WorkerStats[workerID] = workerStats
				stats.Mutex.Unlock()

				// Log de fin worker supprimé pour réduire la verbosité
			}(i)
		}

		// Log de workers démarrés supprimé pour réduire la verbosité

		// Attendre que toutes les goroutines se terminent
		wg.Wait()
		close(completedRecipes)
		// Log de fin workers supprimé pour réduire la verbosité
	}()
}

// startRecipeCollector démarre la goroutine qui collecte les recettes terminées
func startRecipeCollector(completedRecipes <-chan Recipe, recipes *[]Recipe, recipesMutex *sync.RWMutex, done chan<- bool) {
	go func() {
		for recipe := range completedRecipes {
			recipesMutex.Lock()
			*recipes = append(*recipes, recipe)
			recipesMutex.Unlock()
		}
		done <- true
	}()
}

// saveRecipesToFile sauvegarde les recettes dans un fichier JSON
func saveRecipesToFile(recipes []Recipe, filename string) error {
	content, err := json.MarshalIndent(recipes, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, content, 0644)
}

// printDetailedStats affiche les statistiques détaillées
func printDetailedStats(stats *ScrapingStats, filename string) {
	stats.CalculateFinalStats()
	detailedStats := stats.GetDetailedStats()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 STATISTIQUES DÉTAILLÉES DU COLLECTEUR")
	fmt.Println(strings.Repeat("=", 80))

	// Performance générale
	fmt.Printf("⏱️  Durée totale: %v\n", detailedStats.TotalDuration)
	fmt.Printf("🚀 Requêtes par seconde: %.2f\n", detailedStats.RequestsPerSecond)
	fmt.Printf("📝 Recettes par seconde: %.2f\n", detailedStats.RecipesPerSecond)

	// Requêtes
	fmt.Println("\n🌐 REQUÊTES:")
	fmt.Printf("   Total: %d\n", detailedStats.TotalRequests)
	fmt.Printf("   Page principale: %d\n", detailedStats.MainPageRequests)
	fmt.Printf("   Pages recettes: %d\n", detailedStats.RecipeRequests)

	// Recettes
	fmt.Println("\n📝 RECETTES:")
	fmt.Printf("   Trouvées: %d\n", detailedStats.RecipesFound)
	fmt.Printf("   Complétées: %d\n", detailedStats.RecipesCompleted)
	fmt.Printf("   Échouées: %d\n", detailedStats.RecipesFailed)
	fmt.Printf("   Taux de succès: %.1f%%\n", float64(detailedStats.RecipesCompleted)/float64(detailedStats.RecipesFound)*100)

	// Configuration automatique
	numLogicalCPU := runtime.NumCPU()
	numPhysicalCores := getPhysicalCores()
	adaptiveRatio := calculateAdaptiveRatio(numPhysicalCores)
	fmt.Println("\n💻 CONFIGURATION AUTOMATIQUE:")
	fmt.Printf("   Processeurs logiques: %d\n", numLogicalCPU)
	fmt.Printf("   Cœurs physiques détectés: %d\n", numPhysicalCores)
	fmt.Printf("   Ratio adaptatif: %d (calculé automatiquement)\n", adaptiveRatio)
	fmt.Printf("   Calcul: %d cœurs × %d = %d workers\n", numPhysicalCores, adaptiveRatio, numPhysicalCores*adaptiveRatio)
	fmt.Printf("   Configuration finale: %d workers\n", detailedStats.MaxWorkers)

	// Détails par worker
	if len(detailedStats.WorkerStats) > 0 {
		fmt.Println("\n📈 PERFORMANCE PAR WORKER:")
		for workerID, workerStats := range detailedStats.WorkerStats {
			fmt.Printf("   Worker #%d: %d requêtes, %d recettes, %v\n",
				workerID, workerStats.RequestsHandled, workerStats.RecipesProcessed, workerStats.Duration)
		}
	}

	// Calculs de performance
	avgRequestsPerRecipe := float64(detailedStats.RecipeRequests) / float64(detailedStats.RecipesCompleted)
	fmt.Println("\n💡 ANALYSE DE PERFORMANCE:")
	fmt.Printf("   Requêtes moyennes par recette: %.1f\n", avgRequestsPerRecipe)
	fmt.Printf("   Débit estimé: %.0f requêtes/seconde\n", detailedStats.RequestsPerSecond)

	if detailedStats.RecipesPerSecond > 0 {
		fmt.Printf("   Temps moyen par recette: %.2f secondes\n", 1/detailedStats.RecipesPerSecond)
	}

	fmt.Printf("\n💾 Fichier de sortie: %s\n", filename)
	fmt.Println(strings.Repeat("=", 80))
}

// printRealTimeStats affiche les statistiques en temps réel (désactivé pour réduire la verbosité)
func printRealTimeStats(stats *ScrapingStats) {
	// Logs de temps réel désactivés pour réduire la verbosité
	// Les statistiques finales sont toujours affichées à la fin
}

// main est la fonction principale du collecteur
// Elle orchestre tout le processus de collecte : collecte des URLs, traitement des recettes, et sauvegarde
func main() {
	// ===== PHASE 1: INITIALISATION =====
	// Afficher les informations de version et de build
	printVersionInfo()

	// Configuration du collecteur - paramètres ajustables
	const minWorkers = 1          // Nombre minimum de workers
	const maxWorkers = 100        // Nombre maximum de workers
	const maxPagesPerCategory = 5 // Nombre maximum de pages à collecter par catégorie
	const maxRecipesPerPage = 20  // Estimation du nombre de recettes par page

	// Configuration automatique basée sur les ressources CPU
	optimalWorkers := calculateOptimalWorkers(minWorkers, maxWorkers)

	// Configuration automatique (logs supprimés pour réduire la verbosité)

	// Créer l'objet de statistiques thread-safe
	stats := NewScrapingStats(optimalWorkers)

	// Note: Dans Go 1.20+, le générateur global rand est automatiquement initialisé
	// Pas besoin d'appeler rand.Seed() qui est déprécié

	// Afficher les informations de démarrage (simplifié)
	log.Printf("🚀 Collecteur démarré avec %d workers\n", optimalWorkers)
	log.Printf("📊 Configuration: %d pages/catégorie, %d recettes/page max\n", maxPagesPerCategory, maxRecipesPerPage)

	// Démarrer l'affichage des statistiques en temps réel (désactivé pour réduire la verbosité)
	printRealTimeStats(stats)

	// ===== PHASE 2: CONFIGURATION DES CHANNELS =====
	// Channels pour la communication entre goroutines (pipeline de données)
	recipeURLs := make(chan RecipeData, 2000)   // Channel pour les URLs de recettes (buffer de 2000)
	completedRecipes := make(chan Recipe, 2000) // Channel pour les recettes complétées (buffer de 2000)
	done := make(chan bool)                     // Channel de signalisation de fin

	// Slice thread-safe pour stocker toutes les recettes finales
	var recipes []Recipe
	var recipesMutex sync.RWMutex // Mutex pour protéger l'accès concurrent au slice

	// WaitGroup pour synchroniser l'attente de la fin de toutes les goroutines
	var wg sync.WaitGroup

	// ===== PHASE 3: CONFIGURATION DES COLLECTEURS =====
	// Créer le collecteur principal avec support de la pagination
	mainCollector := createMainCollectorWithPagination(stats, recipeURLs, maxPagesPerCategory)

	// ===== PHASE 4: DÉMARRAGE DES GOROUTINES DE TRAITEMENT =====
	// Démarrer la goroutine qui collecte les recettes terminées
	startRecipeCollector(completedRecipes, &recipes, &recipesMutex, done)

	// Démarrer les workers qui traitent les URLs de recettes
	startRecipeProcessor(recipeURLs, completedRecipes, stats, &wg)

	// ===== PHASE 5: DÉFINITION DES CATÉGORIES À COLLECTER =====
	// Liste des catégories de recettes AllRecipes à collecter
	// Chaque catégorie sera visitée avec pagination automatique
	categories := []string{
		"https://www.allrecipes.com/recipes/16369/soups-stews-and-chili/soup/",               // Soupes
		"https://www.allrecipes.com/recipes/1246/soups-stews-and-chili/soup/chicken-soup/",   // Soupes de poulet
		"https://www.allrecipes.com/recipes/76/appetizers-and-snacks/",                       // Apéritifs et collations
		"https://www.allrecipes.com/recipes/113/appetizers-and-snacks/pastries/",             // Pâtisseries
		"https://www.allrecipes.com/recipes/1059/fruits-and-vegetables/vegetables/",          // Légumes
		"https://www.allrecipes.com/recipes/1083/fruits-and-vegetables/vegetables/cucumber/", // Concombres
		"https://www.allrecipes.com/recipes/77/drinks/",                                      // Boissons
		"https://www.allrecipes.com/recipes/79/desserts/",                                    // Desserts
		"https://www.allrecipes.com/recipes/81/side-dish/",                                   // Accompagnements
		"https://www.allrecipes.com/recipes/1569/everyday-cooking/on-the-go/tailgating/",     // Tailgating
	}

	// ===== PHASE 6: VISITE INITIALE DE LA PAGE D'ACCUEIL =====
	// Visiter la page d'accueil pour obtenir les cookies de session (important pour contourner Cloudflare)
	log.Printf("Visite de la page d'accueil pour obtenir les cookies de session...\n")
	homepageCollector := colly.NewCollector()
	homepageCollector.OnRequest(func(r *colly.Request) {
		configureRealisticHeaders(r)
		// Pour la première visite, utiliser Google comme referer
		r.Headers.Set("Referer", "https://www.google.com/")
	})
	err := homepageCollector.Visit("https://www.allrecipes.com/")
	if err != nil {
		log.Printf("⚠️  Erreur lors de la visite de la page d'accueil: %v\n", err)
	} else {
		log.Printf("✅ Page d'accueil visitée avec succès, cookies de session obtenus\n")
		// Attendre un peu après la visite de la page d'accueil
		time.Sleep(getRandomDelay(2000, 4000))
	}

	// ===== PHASE 7: EXÉCUTION DE LA COLLECTE =====
	// Démarrer la collecte de toutes les catégories définies
	log.Printf("Début de la collecte de %d catégories...\n", len(categories))
	for i, category := range categories {
		// Log de catégorie supprimé pour réduire la verbosité
		_ = i // Variable utilisée mais pas loggée

		// Visiter la catégorie (avec pagination automatique)
		err := mainCollector.Visit(category)
		if err != nil {
			log.Printf("⚠️  Erreur lors de la visite de la catégorie %s: %v\n", category, err)
			continue // Continuer avec la catégorie suivante en cas d'erreur
		}

		// Pause respectueuse entre les catégories avec délai aléatoire augmenté
		randomDelay := getRandomDelay(5000, 10000) // Délai aléatoire entre 5s et 10s
		time.Sleep(randomDelay)
	}

	// ===== PHASE 8: FINALISATION =====
	// Fermer le channel des URLs pour signaler qu'il n'y a plus de recettes à traiter
	close(recipeURLs)

	// Attendre que toutes les recettes soient collectées (signal du collector)
	<-done

	// ===== PHASE 9: SAUVEGARDE ET STATISTIQUES =====
	// Sauvegarder toutes les recettes dans un fichier JSON
	filename := "data.json"
	recipesMutex.RLock()
	err = saveRecipesToFile(recipes, filename)
	recipesMutex.RUnlock()

	if err != nil {
		log.Printf("Erreur lors de l'enregistrement des recettes: %v\n", err)
		return
	}

	// Afficher les statistiques détaillées de performance
	printDetailedStats(stats, filename)

	// Afficher les informations de build dans les logs finaux
	log.Printf("✅ Collecte terminée\n")
}
