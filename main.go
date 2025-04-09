//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"recortesKindle/paquetes/escritura"
	"recortesKindle/paquetes/lectura"
	"recortesKindle/paquetes/proceso"
)

// Estructura para el JSON
type Registro struct {
	ID          int    `json:"id"`
	Autor       string `json:"autor"`
	Nombre      string `json:"nombre"`
	Pagina      int    `json:"pagina"`
	Contenido   string `json:"contenido"`
	Visibilidad bool   `json:"visibilidad"`
	Fecha       string `json:"fecha"`
	Hora        string `json:"hora"`
}

func main() {
	// Determinar qué archivo de entrada usar
	archivoEntrada := "My Clippings.txt" // Valor por defecto
	archivoSalida := "notas.json"       // Valor por defecto

	// Si se proporciona un argumento, usarlo como ruta del archivo
	if len(os.Args) > 1 {
		archivoEntrada = os.Args[1]
	}

	// Si se proporciona un segundo argumento, usarlo como archivo de salida
	if len(os.Args) > 2 {
		archivoSalida = os.Args[2]
	}

	// Leer el archivo de texto
	lines, err := lectura.LeerArchivo(archivoEntrada)
	if err != nil {
		log.Fatalf("Error al leer el archivo: %v", err)
	}

	// Procesar los recortes
	recortes, err := proceso.ProcesoDeLineas(lines)
	if err != nil {
		log.Fatalf("Error al procesar los recortes: %v", err)
	}

	// Escribir los recortes en un archivo JSON
	err = escritura.EscribirJSON(archivoSalida, recortes)
	if err != nil {
		log.Fatalf("Error al escribir el archivo JSON: %v", err)
	}

	// Mostrar el JSON en la salida estándar para que Python lo pueda capturar
	contenidoJSON, err := os.ReadFile(archivoSalida)
	if err != nil {
		log.Fatalf("Error al leer el archivo JSON generado: %v", err)
	}
	fmt.Print(string(contenidoJSON))

	jsonData, err := json.MarshalIndent(recortes, "", "  ")
	if err != nil {
		log.Fatalf("Error al convertir a JSON: %v", err)
	}
	fmt.Print(string(jsonData))
}
