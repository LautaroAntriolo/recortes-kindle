//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"recortesKindle/paquetes/analisis"
	"recortesKindle/paquetes/escritura"
	"recortesKindle/paquetes/lectura"
	"recortesKindle/paquetes/proceso"
)

type Documento struct {
	ID          int         `json:"id"`
	Autor       string      `json:"autor"`
	Nombre      string      `json:"nombre"`
	Pagina      int         `json:"pagina"`
	Contenido   string      `json:"contenido"`
	Visibilidad bool        `json:"visibilidad"`
	Fecha       string      `json:"fecha"`
	Hora        string      `json:"hora"`
	Etiquetas   interface{} `json:"etiquetas"`
}

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
	archivoEntrada := "recortes.txt" // Valor por defecto
	archivoSalida := "notas.json"    // Valor por defecto
	terminoBusqueda := ""

	// Si se proporciona un argumento, usarlo como ruta del archivo
	if len(os.Args) > 1 {
		archivoEntrada = os.Args[1]
	}

	// Si se proporciona un tercer argumento, usarlo como resultado de salida
	if len(os.Args) > 2 {
		archivoSalida = os.Args[2]
	}

	// Si se proporciona un segundo argumento, usarlo como termino de busqueda
	if len(os.Args) > 3 {
		terminoBusqueda = os.Args[3]
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

	resultados, jsonFiltrado, err := analisis.Similitudes(jsonInfo, terminoBusqueda)
	if err != nil {
		log.Fatal("Error buscando coincidencias:", err)
	}

	if err := os.WriteFile("similitudes.json", jsonFiltrado, 0644); err != nil {
		log.Fatal("Error guardando resultados:", err)
	}

	fmt.Printf("✔ Se encontraron %d resultados\n✔ Guardados en: similitudes\n", len(resultados))

	var docs []Documento
	err = json.Unmarshal(jsonFiltrado, &docs)
	if err != nil {
		// Verifica si es un solo documento
		var singleDoc Documento
		if err := json.Unmarshal(jsonFiltrado, &singleDoc); err != nil {
			log.Fatalf("Error al deserializar JSON: %v", err)
		}
		docs = []Documento{singleDoc} // Convierte a slice con un elemento
	}

	// Crear el archivo HTML
	file, err := os.Create("diagrama.html")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Plantilla corregida
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Mapa de Documentos</title>
	</head>
	<main>
		<center>
			<h1>Acá irán todos los nodos</h1>
		</center>
	</main>
	<html>
	`

	t := template.Must(template.New("diagrama.html").Parse(tmpl))
	if err := t.Execute(file, docs); err != nil {
		log.Fatalf("Error al ejecutar plantilla: %v", err)
	}

	fmt.Println("🚀 Servidor iniciado en http://localhost:8080")
	fmt.Println("Presiona Ctrl+C para detenerlo")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "diagrama.html")
	})

	// Bloquea aquí y sirve el archivo HTML hasta que lo detengas
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}

}
