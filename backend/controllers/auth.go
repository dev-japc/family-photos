package controllers

import (
	"fmt"
	"errors"
	"net/http"
	"time"

	"backend/config"
	"backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// key for JWT for development - in production, use ENV variables
var jwtSecret = []byte("my-super-secret-key-dev-secure-secret-key-for-development")

type RegisterInput struct {
	FullName         string `json:"full_name" binding:"required"`
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=6"`
	NationalIdNumber string `json:"national_id_number" binding:"required"`
}

//LOGIN structure
type LoginInput struct {
	NationalIdNumber string `json:"national_id_number" binding:"required"`
	Password 		 string `json:"password" binding:"required"`
}

// AddToWhitelist
func AddToWhitelist(c *gin.Context) {
    type WhitelistInput struct {
        NationalIDNumber string `json:"national_id_number" binding:"required"`
        Name             string `json:"name" binding:"required"`
    }

    var input WhitelistInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Id number and name are required"})
        return
    }

    type AllowedID struct {
        NationalIDNumber string `gorm:"column:national_id_number"`
        Name             string `gorm:"column:name"`
    }

    var existingID AllowedID
    // Check if ID already exists in whitelist
    err := config.DB.Table("allowed_national_id_numbers").Where("national_id_number = ?", input.NationalIDNumber).First(&existingID).Error
    if err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "This ID number is already in the whitelist"})
        return
    }

    newAllowed := AllowedID{
        NationalIDNumber: input.NationalIDNumber,
        Name:             input.Name,
    }
    if err := config.DB.Table("allowed_national_id_numbers").Create(&newAllowed).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add ID to whitelist"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": fmt.Sprintf("ID '%s' ('%s') added successfully to whitelist", input.NationalIDNumber, input.Name),
    })
}

// DeleteFromWhitelist deletes a national ID number from the whitelist
func DeleteFromWhitelist(c *gin.Context) {
	nationalID := c.Param("national_id_number")

	var allowed models.AllowedNationalIdNumber
	// Check if ID exists in whitelist
	if err := config.DB.Where("national_id_number = ?", nationalID).First(&allowed).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "This ID number is not in the whitelist"})
		return
	}

	// Delete the ID from whitelist
	if err := config.DB.Delete(&allowed).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete ID from whitelist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("ID '%s' ('%s') deleted successfully from whitelist", allowed.NationalIdNumber, allowed.Name),
	})
}

// GetWhitelist retrieves all allowed national ID numbers from the whitelist
func GetWhitelist(c *gin.Context) {
	var whitelist []models.AllowedNationalIdNumber

	if err := config.DB.Find(&whitelist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve whitelist"})
		return
	}

	c.JSON(http.StatusOK, whitelist)
}

func Register(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	var allowed models.AllowedNationalIdNumber
	if err := config.DB.Where("national_id_number = ?", input.NationalIdNumber).First(&allowed).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "This ID is not allowed to register. Family members only"})
		return
	}
	var existingUser models.User
	err := config.DB.Where("national_id_number = ?", input.NationalIdNumber).First(&existingUser).Error
	
	// If err == nil, it means a user with this ID was found
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists!"})
		return
	}
	
	// If it returns an error different from "not found", it's a real server problem
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno al verificar el usuario"})
		return
	}

	// 3. Encriptar la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar la seguridad de las credenciales"})
		return
	}

	// 4. Create users
	newUser := models.User{
		FullName:         input.FullName,
		Email:            input.Email,
		Password:         string(hashedPassword),
		NationalIdNumber: input.NationalIdNumber,
		Role:             "family_member",//default role for new users
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create the user in the system"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully. Welcome to the family"})
}

func Login(c *gin.Context) {
	var input LoginInput

	//validate json input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID and password are required"})
		return
	}

	//look for user in DB
	var user models.User
	if err := config.DB.Where("national_id_number = ?", input.NationalIdNumber).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// compare password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Credentials!!!"})
		return
	}

	// if everything goes fine - create "claims" - information utility inside token
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role": user.Role,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	// generate and sign a new token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "can't generate access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login OK",
		"token": tokenString,
		"full_name": user.FullName,
		"role": user.Role,		
	})
}

// UpdateUser 
func UpdateUser(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// temporal structure to receive JSON with changes
	type UpdateInput struct {
		FullName string `json:"full_name"`
		Password string `json:"password"`
	}

	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	// update name if it comes in the request
	if input.FullName != "" {
		user.FullName = input.FullName
	}

	// if the password changes, re-encrypt it with Bcrypt
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing the new password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// DeleteUser deletes the account of the authenticated user
func DeleteUser(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete user account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully. Sad to see you go."})
}

// AdminDeleteUser allow admin delete any user from the system
func AdminDeleteUser(c *gin.Context) {
	// 1. Get user ID to delete from URL
	targetUserID := c.Param("id")

	// Prevent admin from deleting itself
	adminID, _ := c.Get("userID")
	if fmt.Sprintf("%v", adminID) == targetUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can't delete your own account from here"})
		return
	}

	var user models.User
	// 2. Check if target user exists
	if err := config.DB.First(&user, targetUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "The user you are trying to delete does not exist"})
		return
	}

	// 3. Delete user from database
	if err := config.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete the user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("User '%s' (ID: %s) deleted successfully by the administrator", user.FullName, targetUserID),
	})
}