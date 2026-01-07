package controllers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
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
	scraperPath := "/app/scraper"

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

	// S'assurer que le répertoire de sauvegarde existe
	dataDir := "/go_api_mongo_scrapper/scraper"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.LogError("Erreur lors de la création du répertoire de sauvegarde", err, map[string]interface{}{
			"data_dir": dataDir,
		})
		// Continuer quand même, le volume peut déjà exister
	}

	// Commande pour exécuter le scraper
	cmd := exec.Command(scraperPath)

	// Définir le répertoire de travail pour que le fichier data.json soit sauvegardé dans un emplacement connu
	cmd.Dir = dataDir

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

// LogMessage représente un message de log pour le streaming
type LogMessage struct {
	Type      string `json:"type"`      // "stdout", "stderr", "info", "error", "done"
	Message   string `json:"message"`   // Contenu du message
	Timestamp string `json:"timestamp"` // Timestamp ISO 8601
}

// LaunchScraperStream lance le scraper et stream les logs en temps réel via SSE
func LaunchScraperStream(c *fiber.Ctx) error {
	requestID := c.Locals("requestID").(string)
	start := time.Now()

	// Configuration des headers pour Server-Sent Events (SSE)
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // Désactive le buffering de nginx

	logger.LogInfo("Démarrage du scraper (mode streaming)", map[string]interface{}{
		"request_id": requestID,
	})

	// Chemin vers le binaire du scraper
	scraperPath := "/app/scraper"

	// Vérifie que le fichier existe
	if _, err := os.Stat(scraperPath); os.IsNotExist(err) {
		errorMsg := fmt.Sprintf("❌ Binaire scraper introuvable: %s", scraperPath)
		logger.LogError("Binaire scraper introuvable", err, map[string]interface{}{
			"scraper_path": scraperPath,
			"request_id":   requestID,
		})
		return c.Status(500).SendString(errorMsg)
	}

	// Utiliser directement BodyWriter pour le streaming
	w := c.Context().Response.BodyWriter()

	// Message de démarrage
	startMsg := LogMessage{
		Type:      "info",
		Message:   "🚀 Démarrage du scraper...",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	jsonData, _ := json.Marshal(startMsg)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)

	// S'assurer que le répertoire de sauvegarde existe
	dataDir := "/go_api_mongo_scrapper/scraper"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.LogError("Erreur lors de la création du répertoire de sauvegarde", err, map[string]interface{}{
			"data_dir":   dataDir,
			"request_id": requestID,
		})
		// Continuer quand même, le volume peut déjà exister
	}

	// Commande pour exécuter le scraper
	cmd := exec.Command(scraperPath)

	// Définir le répertoire de travail pour que le fichier data.json soit sauvegardé dans un emplacement connu
	cmd.Dir = dataDir

	// Créer des pipes pour capturer stdout et stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erreur lors de la création du pipe stdout: %v", err)
		msg := LogMessage{
			Type:      "error",
			Message:   errorMsg,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		jsonData, _ := json.Marshal(msg)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erreur lors de la création du pipe stderr: %v", err)
		msg := LogMessage{
			Type:      "error",
			Message:   errorMsg,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		jsonData, _ := json.Marshal(msg)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		return err
	}

	// Démarrer la commande
	if err := cmd.Start(); err != nil {
		errorMsg := fmt.Sprintf("❌ Erreur lors du démarrage du scraper: %v", err)
		msg := LogMessage{
			Type:      "error",
			Message:   errorMsg,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		jsonData, _ := json.Marshal(msg)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		logger.LogError("Erreur lors du démarrage du scraper", err, map[string]interface{}{
			"request_id": requestID,
		})
		return err
	}

	// WaitGroup pour synchroniser les goroutines
	var wg sync.WaitGroup

	// Goroutine pour lire stdout ligne par ligne
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			msg := LogMessage{
				Type:      "stdout",
				Message:   line,
				Timestamp: time.Now().Format(time.RFC3339),
			}
			jsonData, _ := json.Marshal(msg)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
		}
	}()

	// Goroutine pour lire stderr ligne par ligne
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			msg := LogMessage{
				Type:      "stderr",
				Message:   line,
				Timestamp: time.Now().Format(time.RFC3339),
			}
			jsonData, _ := json.Marshal(msg)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
		}
	}()

	// Attendre la fin de l'exécution
	err = cmd.Wait()
	wg.Wait() // Attendre que toutes les goroutines de lecture soient terminées

	if err != nil {
		errorMsg := fmt.Sprintf("❌ Le scraper s'est terminé avec une erreur: %v", err)
		msg := LogMessage{
			Type:      "error",
			Message:   errorMsg,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		jsonData, _ := json.Marshal(msg)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		logger.LogError("Échec de l'exécution du scraper", err, map[string]interface{}{
			"scraper_path": scraperPath,
			"request_id":   requestID,
		})
		return err
	}

	// Message de fin
	duration := time.Since(start)
	successMsg := fmt.Sprintf("✅ Scraper exécuté avec succès en %s", duration.String())
	msg := LogMessage{
		Type:      "done",
		Message:   successMsg,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	jsonData, _ = json.Marshal(msg)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)

	logger.LogInfo("Scraper exécuté avec succès (mode streaming)", map[string]interface{}{
		"request_id": requestID,
		"duration":   duration.String(),
	})

	return nil
}

// GetScraperData récupère le fichier JSON généré par le scraper
func GetScraperData(c *fiber.Ctx) error {
	requestID := "unknown"
	if id, ok := c.Locals("requestID").(string); ok {
		requestID = id
	}

	// Emplacements possibles du fichier data.json
	possiblePaths := []string{
		"/app/data.json", // Répertoire de travail de l'API
		"/go_api_mongo_scrapper/scraper/data.json", // Volume partagé scraper_data
		"./data.json", // Répertoire courant
		"data.json",   // Répertoire courant (relatif)
	}

	var filePath string
	var found bool

	// Chercher le fichier dans les emplacements possibles
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			filePath = path
			found = true
			break
		}
	}

	if !found {
		logger.LogError("Fichier data.json introuvable", nil, map[string]interface{}{
			"request_id":     requestID,
			"searched_paths": possiblePaths,
		})
		return c.Status(404).JSON(fiber.Map{
			"error":   true,
			"message": "Fichier data.json introuvable. Le scraper n'a peut-être pas encore été exécuté.",
		})
	}

	// Lire le fichier
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		logger.LogError("Erreur lors de la lecture du fichier data.json", err, map[string]interface{}{
			"request_id": requestID,
			"file_path":  filePath,
		})
		return c.Status(500).JSON(fiber.Map{
			"error":   true,
			"message": "Erreur lors de la lecture du fichier",
		})
	}

	// Obtenir les informations du fichier
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		logger.LogError("Erreur lors de la récupération des informations du fichier", err, map[string]interface{}{
			"request_id": requestID,
			"file_path":  filePath,
		})
	}

	// Définir les headers pour le téléchargement
	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"scraper-data-%s.json\"", time.Now().Format("20060102-150405")))
	if fileInfo != nil {
		c.Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	}

	logger.LogInfo("Fichier data.json téléchargé avec succès", map[string]interface{}{
		"request_id": requestID,
		"file_path":  filePath,
		"file_size":  len(fileContent),
	})

	// Envoyer le fichier
	return c.Send(fileContent)
}
