package controllers

import (
	"os"
	"os/exec"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maxime-louis14/api-golang/logger"
)

// LaunchScraper lance le scraper via une route API
func LaunchScraper(c *fiber.Ctx) error {
	start := time.Now()
	requestID := c.Locals("requestID").(string)

	logger.LogInfo("Démarrage du scraper", map[string]interface{}{
		"request_id": requestID,
	})

	// Ajoute un délai de 4 secondes
	time.Sleep(4 * time.Second)

	// Exécute le scraper
	if err := RunScraper(); err != nil {
		logger.LogError("Erreur lors de l'exécution du scraper", err, map[string]interface{}{
			"request_id": requestID,
		})
		return c.Status(500).SendString("Erreur lors de l'exécution du scraper")
	}

	duration := time.Since(start)
	logger.LogInfo("Scraper exécuté avec succès", map[string]interface{}{
		"request_id": requestID,
		"duration":   duration.String(),
	})

	return c.Status(200).SendString("Scraper exécuté avec succès")
}

// RunScraper exécute le binaire du scraper
func RunScraper() error {
	start := time.Now()
	// Chemin vers le binaire du scraper
	scraperPath := "/go_api_mongo_scrapper/scraper/scraper"

	logger.LogInfo("Vérification de l'existence du binaire scraper", map[string]interface{}{
		"scraper_path": scraperPath,
	})

	// Vérifie que le fichier existe
	if _, err := os.Stat(scraperPath); os.IsNotExist(err) {
		logger.LogError("Binaire scraper introuvable", err, map[string]interface{}{
			"scraper_path": scraperPath,
		})
		return err
	}

	logger.LogInfo("Lancement du binaire scraper", map[string]interface{}{
		"scraper_path": scraperPath,
	})

	// Commande pour exécuter le scraper
	cmd := exec.Command(scraperPath)

	// Associe les sorties standard et erreur du scraper aux sorties du serveur
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Exécute la commande
	if err := cmd.Run(); err != nil {
		logger.LogError("Échec de l'exécution du scraper", err, map[string]interface{}{
			"scraper_path": scraperPath,
		})
		return err
	}

	duration := time.Since(start)
	logger.LogInfo("Scraper exécuté avec succès", map[string]interface{}{
		"scraper_path": scraperPath,
		"duration":     duration.String(),
	})
	return nil
}

// GetScraperData récupère le fichier JSON généré par le scraper
func GetScraperData(c *fiber.Ctx) error {
	// Emplacements possibles du fichier data.json
	possiblePaths := []string{
		"/go_api_mongo_scrapper/scraper/data.json", // Volume partagé (emplacement principal)
		"/app/data.json", // Répertoire de travail de l'API
		"./data.json",    // Répertoire courant
		"data.json",      // Répertoire courant (relatif)
	}

	var filePath string
	var found bool

	// Chercher le fichier dans les emplacements possibles
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			filePath = path
			found = true
			logger.LogInfo("Fichier data.json trouvé", map[string]interface{}{
				"path": filePath,
			})
			break
		}
	}

	if !found {
		logger.LogError("Fichier data.json introuvable", nil, map[string]interface{}{
			"paths_checked": possiblePaths,
		})
		return c.Status(404).JSON(fiber.Map{
			"error":   true,
			"message": "Fichier data.json introuvable. Le scraper doit être exécuté au moins une fois.",
		})
	}

	// Lire le fichier
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		logger.LogError("Erreur lors de la lecture du fichier data.json", err, map[string]interface{}{
			"path": filePath,
		})
		return c.Status(500).JSON(fiber.Map{
			"error":   true,
			"message": "Erreur lors de la lecture du fichier",
		})
	}

	// Déterminer le nom du fichier avec timestamp
	timestamp := time.Now().Format("20060102-150405")
	filename := "scraper-data-" + timestamp + ".json"

	// Retourner le fichier avec les headers appropriés
	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.Send(fileContent)
}
