package controllers

import (
	"backend/config"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateAlbum crea un nuevo contenedor de fotos (Solo Admin)
func CreateAlbum(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre del álbum es obligatorio"})
		return
	}

	album := models.Album{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := config.DB.Create(&album).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el álbum"})
		return
	}

	c.JSON(http.StatusCreated, album)
}

// GetAlbums lista todos los álbumes disponibles
func GetAlbums(c *gin.Context) {
	var albums []models.Album
	
	// Buscamos los álbumes ordenados por el más reciente
	if err := config.DB.Order("created_at desc").Find(&albums).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los álbumes"})
		return
	}

	c.JSON(http.StatusOK, albums)
}