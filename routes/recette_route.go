package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maxime-louis14/api-golang/controllers"
)

// GetRecetteByName récupère une recette par son nom
// @Summary Récupérer une recette par son nom
// @Description Récupère une recette en utilisant son nom
// @Tags Recettes
// @Param name path string true "Nom de la recette"
// @Produce json
// @Success 200 {object} models.Recette
// @Failure 404 {string} string "Recette introuvable"
// @Router /recettes/{name} [get]

func RecetteRoute(app *fiber.App) {
	app.Post("/scraper/run", controllers.LaunchScraper)
	app.Get("/scraper/data", controllers.GetScraperData)
	app.Post("/recettes", controllers.PostRecette)
	app.Get("/recettes", controllers.GetAllRecettes)
	app.Get("/recette/:id", controllers.GetRecetteByID)
	app.Get("/recette/name/:name", controllers.GetRecetteByName)
	app.Get("/recette/ingredient/:ingredient", controllers.GetRecettesByIngredient)

}
