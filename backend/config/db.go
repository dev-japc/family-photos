package config

import (
	"log"
	"os"

	"backend/models" // Asegúrate de que coincida con el nombre de tu módulo en go.mod

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB es la variable global que usaremos para hacer consultas a la base de datos
var DB *gorm.DB

func ConnectDatabase() {
	// Leemos la URL de la base de datos desde el entorno de Docker
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Attempt to open the connection with PostgreSQL
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	log.Println("Successfully connected to the database.")

	// AutoMigrate se encarga de revisar tus 'structs' de Go y crear o actualizar 
	// las tablas en PostgreSQL de forma automática si no existen.
	err = database.AutoMigrate(&models.AllowedNationalIdNumber{}, &models.User{})
	if err != nil {
		log.Fatal("Error executing migrations: ", err)
	}

	log.Println("Migrations completed.")

	// Asignamos la conexión a nuestra variable global
	DB = database
}