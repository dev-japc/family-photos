package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backend/config"
	"backend/models"

	"github.com/gin-gonic/gin"
)

func UploadPhoto(c *gin.Context) {
	const MaxUploadSize = 5 * 1024 * 1024 
    c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadSize)

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

	extension := strings.ToLower(filepath.Ext(file.Filename))
    if extension != ".jpg" && extension != ".jpeg" && extension != ".png" && extension != ".webp" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format. Only JPG, JPEG, PNG, and WEBP are allowed"})
        return
    }

	// generate unique filename in milliseconds
	// extension := filepath.Ext(file.Filename)
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
	if description == "" {
		description = title
	}

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

// GetPhotos gets photos with pagination (Max 10 per page)
func GetPhotos(c *gin.Context) {
    var photos []models.Photo
    
    pageStr := c.DefaultQuery("page", "1")

    var page int

    fmt.Sscanf(pageStr, "%d", &page)
    if page < 1 {
        page = 1
    }

    // 2. limiting the number of photos per page to 10
    pageSize := 10
    offset := (page - 1) * pageSize

    // 3. counting the total number of photos in DB to calculate the number of pages needed for pagination
    var totalPhotos int64
    config.DB.Model(&models.Photo{}).Count(&totalPhotos)

    // 4. look for the photos in DB and sort them by creation
    if err := config.DB.Order("created_at desc").Limit(pageSize).Offset(offset).Find(&photos).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the list of photos"})
        return
    }

    // 5. JSON response with the photos, current page, total photos, and total pages
    c.JSON(http.StatusOK, gin.H{
        "photos":       photos,
        "current_page": page,
        "total_photos": totalPhotos,
        "total_pages":  (totalPhotos + int64(pageSize) - 1) / int64(pageSize), // Redondeo hacia arriba
    })
}

func UpdatePhoto(c *gin.Context) {
	photoID := c.Param("id")
	userID, _ := c.Get("userID")

	var photo models.Photo

	// 1. search the photo
	if err := config.DB.First(&photo, photoID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Foto no encontrada"})
		return
	}

	// 2. photo validation to check if user is the owner
	if photo.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para modificar esta foto"})
		return
	}

	// 3. reading the new data from the request
	title := c.PostForm("title")
	description := c.PostForm("description")

	if title != "" {
		photo.Title = title
	}
	photo.Description = description 

	// 4. keep changes if error occurs
	if err := config.DB.Save(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron actualizar los datos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Foto actualizada con éxito", "photo": photo})
}

// DeletePhoto deletes the database record and the physical file from the server
func DeletePhoto(c *gin.Context) {
	photoID := c.Param("id")
	// userID, _ := c.Get("userID")

	var photo models.Photo
	if err := config.DB.First(&photo, photoID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// if photo.UserID != userID.(uint) {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this image"})
	// 	return
	// }

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