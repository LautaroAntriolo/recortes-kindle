//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"recortesKindle/paquetes/analisis"
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
	archivoSalida := "notas.json"        // Valor por defecto

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
	jsonInfo, err := escritura.EscribirJSON(archivoSalida, recortes)
	if err != nil {
		log.Fatalf("Error al escribir el archivo JSON: %v", err)
	}

	terminoBusqueda := "amor"
	resultados, jsonFiltrado, err := analisis.Similitudes(jsonInfo, terminoBusqueda)
	if err != nil {
		log.Fatal("Error buscando coincidencias:", err)
	}

	// 3. Escribir el archivo con resultados (controlado desde main)
	if err := os.WriteFile("similitudes.json", jsonFiltrado, 0644); err != nil {
		log.Fatal("Error guardando resultados:", err)
	}

	fmt.Printf("✔ Se encontraron %d resultados\n✔ Guardados en: similitudes\n", len(resultados))

}
