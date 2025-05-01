package rutas

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"recortesKindle/paquetes/escritura"
	"recortesKindle/paquetes/lectura"
	"recortesKindle/paquetes/proceso"

	"github.com/gorilla/mux"
)

const (
	uploadDir = "./datos/entrada"
	outputDir = "./datos/salida"
)

func init() {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
		os.MkdirAll(outputDir, 0755)
	}
}

func CargarArchivoHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("archivo")
	nombreSinExtension := strings.TrimSuffix(handler.Filename, filepath.Ext(handler.Filename))
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

	outputPath := filepath.Join(outputDir, nombreSinExtension+".json")
	jsonData, err := escritura.EscribirJSON(outputPath, recortes)
	if err != nil {
		http.Error(w, "Error al generar JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("JSON generado (%d bytes)", len(jsonData))
	w.Header().Set("Content-Type", "application/json")
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

func ObtenerDatosHandler(w http.ResponseWriter, r *http.Request) {
	outputPath := filepath.Join(outputDir)

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

func ObtenerArchivoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nombreArchivo := vars["nombre"]

	// Convertir nombre.txt a nombre.json
	nombreJSON := strings.TrimSuffix(nombreArchivo, filepath.Ext(nombreArchivo)) + ".json"
	rutaArchivo := filepath.Join("datos", "salida", nombreJSON)

	// Verificar si el archivo existe
	if _, err := os.Stat(rutaArchivo); os.IsNotExist(err) {
		// Si no existe el .json, buscar .txt como alternativa
		rutaTXT := filepath.Join("datos", "salida", strings.TrimSuffix(nombreArchivo, filepath.Ext(nombreArchivo))+".txt")
		if _, err := os.Stat(rutaTXT); os.IsNotExist(err) {
			http.Error(w, "Archivo no encontrado", http.StatusNotFound)
			return
		}

		// Leer el archivo de texto
		contenidoTXT, err := os.ReadFile(rutaTXT)
		if err != nil {
			http.Error(w, "Error al leer archivo: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Convertir el texto plano a formato JSON
		lineas := strings.Split(string(contenidoTXT), "\n")
		recortes := []map[string]string{}

		// Procesar cada línea o bloque de texto
		var recorteActual map[string]string
		for _, linea := range lineas {
			linea = strings.TrimSpace(linea)
			if linea == "" {
				continue
			}

			// Si la línea parece un título nuevo, crear un nuevo recorte
			if strings.Contains(linea, "(") || strings.HasPrefix(linea, "Big_data") {
				if recorteActual != nil && len(recorteActual) > 0 {
					recortes = append(recortes, recorteActual)
				}
				recorteActual = map[string]string{
					"id":        fmt.Sprintf("recorte-%d", len(recortes)+1),
					"nombre":    linea,
					"contenido": "",
					"autor":     extraerAutor(linea),
				}
			} else if recorteActual != nil {
				// Procesar información como página, posición, etc.
				if strings.Contains(linea, "página") {
					recorteActual["pagina"] = extraerNumero(linea, "página")
				}
				if strings.Contains(linea, "posición") {
					recorteActual["posicion"] = extraerNumero(linea, "posición")
				}

				// Agregar al contenido si no es metadata
				if !strings.HasPrefix(linea, "-") && !strings.HasPrefix(linea, "página") && !strings.HasPrefix(linea, "posición") {
					if recorteActual["contenido"] != "" {
						recorteActual["contenido"] += " "
					}
					recorteActual["contenido"] += linea
				}
			}
		}

		// Agregar el último recorte si existe
		if recorteActual != nil && len(recorteActual) > 0 {
			recortes = append(recortes, recorteActual)
		}

		// Si no se pudo extraer nada, crear un elemento simple
		if len(recortes) == 0 {
			recortes = append(recortes, map[string]string{
				"id":        "recorte-1",
				"nombre":    nombreArchivo,
				"contenido": string(contenidoTXT),
				"fecha":     time.Now().Format("2006-01-02"),
			})
		}

		// Convertir a JSON
		jsonData, err := json.Marshal(recortes)
		if err != nil {
			http.Error(w, "Error al convertir a JSON: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	// Si llegamos aquí, el archivo JSON existe, intentar leerlo
	contenido, err := os.ReadFile(rutaArchivo)
	if err != nil {
		http.Error(w, "Error al leer archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Verificar si es JSON válido
	var js json.RawMessage
	if err := json.Unmarshal(contenido, &js); err != nil {
		// Si no es JSON válido pero existe el archivo, intentar arreglarlo
		contenidoStr := string(contenido)

		// Buscar dónde comienza el JSON real (si hay un prefijo de texto)
		inicioJSON := strings.IndexAny(contenidoStr, "{[")
		if inicioJSON > 0 {
			contenidoStr = contenidoStr[inicioJSON:]

			// Verificar si el JSON recortado es válido
			if err := json.Unmarshal([]byte(contenidoStr), &js); err != nil {
				// Si sigue sin ser válido, convertirlo a un formato JSON simple
				recortes := []map[string]string{
					{
						"id":        "recorte-1",
						"nombre":    nombreArchivo,
						"contenido": contenidoStr,
						"fecha":     time.Now().Format("2006-01-02"),
					},
				}

				jsonData, err := json.Marshal(recortes)
				if err != nil {
					http.Error(w, "Error al convertir a JSON: "+err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonData)
				return
			}

			// Si es válido después de recortar, enviarlo
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(contenidoStr))
			return
		}

		http.Error(w, "El archivo no contiene JSON válido", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(contenido)
}

// Función auxiliar para extraer el autor de una línea
func extraerAutor(linea string) string {
	// Buscar texto entre paréntesis
	inicio := strings.Index(linea, "(")
	fin := strings.Index(linea, ")")

	if inicio >= 0 && fin > inicio {
		return strings.TrimSpace(linea[inicio+1 : fin])
	}

	return ""
}

// Función auxiliar para extraer números de una línea
func extraerNumero(linea, palabra string) string {
	indice := strings.Index(linea, palabra)
	if indice < 0 {
		return ""
	}

	// Buscar el número después de la palabra
	resto := linea[indice+len(palabra):]
	var numero strings.Builder

	for _, r := range resto {
		if unicode.IsDigit(r) {
			numero.WriteRune(r)
		} else if numero.Len() > 0 && !unicode.IsDigit(r) {
			break
		}
	}

	return numero.String()
}
