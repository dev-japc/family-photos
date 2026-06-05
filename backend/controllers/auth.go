package controllers

import (
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

func Register(c *gin.Context) {
	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos o incompletos"})
		return
	}

	// 1. Validar lista blanca (Si da error, es porque NO existe en la lista blanca)
	var allowed models.AllowedNationalIdNumber
	if err := config.DB.Where("national_id_number = ?", input.NationalIdNumber).First(&allowed).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "This ID is not allowed to register. Family members only"})
		return
	}

	// 2. CORREGIDO: Validar si ya existe el usuario
	var existingUser models.User
	err := config.DB.Where("national_id_number = ?", input.NationalIdNumber).First(&existingUser).Error
	
	// Si NO da error (err == nil), significa que SÍ encontró un usuario con ese ID
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists!"})
		return
	}
	
	// Si da un error diferente a "no encontrado", es un problema real del servidor
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

	// 4. Crear el usuario
	newUser := models.User{
		FullName:         input.FullName,
		Email:            input.Email,
		Password:         string(hashedPassword),
		NationalIdNumber: input.NationalIdNumber,
		Role:             "family_member",
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el usuario en el sistema"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado con éxito. ¡Bienvenido a la familia!"})
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