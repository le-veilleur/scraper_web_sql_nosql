package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/maxime-louis14/api-golang/database"
	"github.com/maxime-louis14/api-golang/logger"
	"github.com/maxime-louis14/api-golang/middleware"
	"github.com/maxime-louis14/api-golang/routes"
)

// Variables de versioning injectées lors du build
var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

// BuildInfo contient les informations de build
type BuildInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// HealthResponse structure pour le health check
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Build     BuildInfo `json:"build"`
	Database  string    `json:"database"`
}

// Route d'exposition des métriques
func metricsHandler(c *fiber.Ctx) error {
	metricsJSON, err := logger.GetMetricsJSON()
	if err != nil {
		logger.LogError("Erreur lors de la récupération des métriques", err, nil)
		return c.Status(500).JSON(fiber.Map{
			"error":   true,
			"message": "Erreur lors de la récupération des métriques",
		})
	}

	c.Set("Content-Type", "application/json")
	return c.Send(metricsJSON)
}

func main() {
	// Affichage des informations de version
	fmt.Printf("Go API MongoDB Scrapper\n")
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Git Commit: %s\n", gitCommit)
	fmt.Printf("Build Time: %s\n", buildTime)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	// Initialisation des logs
	logger.LogInfo("Démarrage du serveur", map[string]interface{}{
		"version":    version,
		"git_commit": gitCommit,
		"build_time": buildTime,
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	})

	// Initialisation de l'application Fiber avec configuration
	app := fiber.New(fiber.Config{
		AppName:      fmt.Sprintf("Go API MongoDB Scrapper v%s", version),
		ServerHeader: "Go API MongoDB Scrapper",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
				"version": version,
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(cors.New())

	// Middleware de logging personnalisé
	app.Use(middleware.LoggingMiddleware())

	logger.LogInfo("Application Fiber initialisée avec les middlewares", nil)

	// Connexion à MongoDB
	client := database.DBinstance()
	defer func() {
		logger.LogInfo("Fermeture de la connexion MongoDB", nil)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			logger.LogError("Erreur lors de la déconnexion MongoDB", err, nil)
			log.Fatalf("Error disconnecting MongoDB client: %v", err)
		}
		logger.LogInfo("Connexion MongoDB fermée", nil)
	}()
	logger.LogInfo("Connecté à MongoDB", nil)

	// Route de health check
	app.Get("/health", func(c *fiber.Ctx) error {
		// Test de la connexion MongoDB
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		dbStatus := "connected"
		if err := client.Ping(ctx, nil); err != nil {
			dbStatus = "disconnected"
			logger.LogError("Ping MongoDB échoué", err, nil)
		} else {
			logger.LogDatabase(logger.INFO, "Ping MongoDB réussi", "ping", "mongodb", time.Since(time.Now()), nil)
		}

		return c.JSON(HealthResponse{
			Status:    "ok",
			Timestamp: time.Now(),
			Build: BuildInfo{
				Version:   version,
				GitCommit: gitCommit,
				BuildTime: buildTime,
				GoVersion: runtime.Version(),
				OS:        runtime.GOOS,
				Arch:      runtime.GOARCH,
			},
			Database: dbStatus,
		})
	})

	// Route d'informations de version
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(BuildInfo{
			Version:   version,
			GitCommit: gitCommit,
			BuildTime: buildTime,
			GoVersion: runtime.Version(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
		})
	})

	// Route pour les métriques
	app.Get("/metrics", metricsHandler)

	// Configuration des routes API
	routes.RecetteRoute(app)
	logger.LogInfo("Routes configurées", nil)

	// Démarrage du logger de métriques périodique (toutes les 30 secondes)
	logger.StartMetricsLogger(30 * time.Second)

	// Démarrage du serveur
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.LogInfo("Serveur démarré", map[string]interface{}{
		"port":        port,
		"health_url":  "http://localhost:" + port + "/health",
		"version_url": "http://localhost:" + port + "/version",
		"metrics_url": "http://localhost:" + port + "/metrics",
	})

	if err := app.Listen(":" + port); err != nil {
		logger.LogError("Erreur lors du démarrage du serveur", err, nil)
		log.Fatalf("Error starting server: %v", err)
	}
}
