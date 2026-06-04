package main

import (
	"fmt"
	"net/http"

	"backend/config"
)

func main() {
	//DB init
	config.ConnectDatabase()
	
	// Ruta de prueba
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "¡Let's Go developer.")
	})

	fmt.Println("Server  listening on http://localhost:8080")
	
	// La llave { DEBE estar en la misma línea que el if
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}