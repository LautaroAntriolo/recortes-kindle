package rutas

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"recortesKindle/paquetes/escritura"
	"recortesKindle/paquetes/lectura"
	"recortesKindle/paquetes/proceso"

	"github.com/gorilla/mux"
)

const (
	uploadDir = "./datos/entrada"
	outputDir = "./datos/salida"
)

var outputFile = "notas.json"

func init() {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
		os.MkdirAll(outputDir, 0755)
	}
}

func CargarArchivoHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("archivo")
	if err != nil {
		http.Error(w, "Error al obtener el archivo: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if filepath.Ext(handler.Filename) != ".txt" {
		http.Error(w, "Solo se permiten archivos .txt", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadDir, handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error al guardar el archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error al guardar el contenido: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lines, err := lectura.LeerArchivo(filePath)
	if err != nil {
		http.Error(w, "Error al leer el archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	recortes, err := proceso.ProcesoDeLineas(lines)
	if err != nil {
		http.Error(w, "Error al procesar los recortes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	outputPath := filepath.Join(outputDir, outputFile)
	jsonData, err := escritura.EscribirJSON(outputPath, recortes)
	if err != nil {
		http.Error(w, "Error al generar JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("JSON generado (%d bytes)", len(jsonData))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func ObtenerDatosHandler(w http.ResponseWriter, r *http.Request) {
	outputPath := filepath.Join(outputDir, outputFile)

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		http.Error(w, "Error al leer el archivo JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func MostrarArchivos(w http.ResponseWriter, r *http.Request) {
	dir := "datos/entrada"

	files, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, "No se pudo leer el directorio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var nombres []string
	for _, file := range files {
		if !file.IsDir() {
			nombres = append(nombres, file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nombres); err != nil {
		http.Error(w, "Error al codificar respuesta JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

func ObtenerArchivoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nombreArchivo := vars["nombre"]

	rutaArchivo := fmt.Sprintf("datos/entrada/%s", nombreArchivo)

	contenido, err := os.ReadFile(rutaArchivo)
	if err != nil {
		http.Error(w, "No se pudo leer el archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(contenido)
}