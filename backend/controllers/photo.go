package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"
	"os"

	"backend/config"
	"backend/models"

	"github.com/gin-gonic/gin"
)

func UploadPhoto(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no image was received in the field'image'"})
		return
	}

	// generate unique filename in milliseconds
	extension := filepath.Ext(file.Filename)
	uniqueFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)

	// Define the path where the file will be saved physically on the server
	uploadPath := filepath.Join("uploads", uniqueFilename)

	// save the file on the 'uploads' folder
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving file"})
		return
	}

	// create the record in the database
	// the user should be admin, we need to connect JWT
	// 5. Extraer el UserID real inyectado por el Middleware de autenticación
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user unidentified, unauthenticated"})
		return
	}
	userID := userIDValue.(uint)

	// create the record in the database with the real owner of the photo
	newPhoto := models.Photo{
		Title:       title,
		Description: description,
		URL:         "/uploads/" + uniqueFilename,
		UserID:      userID,                          
	}

	if err := config.DB.Create(&newPhoto).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving photo on DB"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Photo uploaded successfully",
		"photo":   newPhoto,
	})
}

// GetPhotos gets all photos from the database
func GetPhotos(c *gin.Context) {
	var photos []models.Photo

	// Search all photos from the database ordered by the most recent
	if err := config.DB.Order("created_at desc").Find(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the list of photos"})
		return
	}

	// Respond with the list of photos (if there are none, it will return an empty array [])
	c.JSON(http.StatusOK, photos)
}

func UpdatePhoto(c *gin.Context) {
	photoID := c.Param("id")
	userID, _ := c.Get("userID")

	var photo models.Photo
	// 1. Buscar la foto
	if err := config.DB.First(&photo, photoID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Foto no encontrada"})
		return
	}

	// 2. Validar que el usuario sea el dueño de la foto
	if photo.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para modificar esta foto"})
		return
	}

	// 3. Leer los nuevos campos del formulario
	title := c.PostForm("title")
	description := c.PostForm("description")

	if title != "" {
		photo.Title = title
	}
	photo.Description = description // Permite vaciar la descripción si se desea

	// 4. Guardar cambios
	if err := config.DB.Save(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron actualizar los datos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Foto actualizada con éxito", "photo": photo})
}

// DeletePhoto deletes the database record and the physical file from the server
func DeletePhoto(c *gin.Context) {
	photoID := c.Param("id")
	userID, _ := c.Get("userID")

	var photo models.Photo
	if err := config.DB.First(&photo, photoID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	if photo.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this image"})
		return
	}

	// 1. Delete the physical file from the disk (ej: "uploads/17807065459...png")
	// Remove the initial slash from the saved URL to get the correct relative path
	filePath := photo.URL[1:] 
	if err := os.Remove(filePath); err != nil {
		// Log the error but continue to avoid leaving the DB inconsistent
		fmt.Println("Error deleting the physical file:", err)
	}

	// 2. Delete the record from the database
	if err := config.DB.Delete(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting the image from the database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image and file deleted successfully"})
}