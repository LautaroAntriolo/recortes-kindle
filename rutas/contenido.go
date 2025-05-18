package rutas

import (
	"log"
	"net/http"
	"text/template"
)

func Inicio(w http.ResponseWriter, r *http.Request) {
	// 1. Cargar el template con manejo de errores
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("Error al parsear template: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// 2. Ejecutar el template con manejo de errores
	err = tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Error al ejecutar template: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
}