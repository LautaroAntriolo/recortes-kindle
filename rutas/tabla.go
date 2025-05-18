package rutas

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"recortesKindle/paquetes/analisis"
	"text/template"

	"github.com/gorilla/mux"
)

func Inicio(w http.ResponseWriter, r *http.Request) {
	// Verifica que la plantilla existe
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Error al cargar plantilla: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Ejecuta la plantilla sin datos (o con los datos que necesites)
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error al renderizar plantilla: "+err.Error(), http.StatusInternalServerError)
	}
}

// Handler para cargar un archivo específico
func DataHandler(w http.ResponseWriter, r *http.Request) {
	nombre := r.URL.Query().Get("archivo")
	if nombre == "" {
		http.Error(w, "Falta el parámetro 'archivo'", http.StatusBadRequest)
		return
	}

	path := filepath.Join("datos/salida", nombre)
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "No se pudo abrir el archivo JSON", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var data interface{}
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		http.Error(w, "Error al parsear el JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func Similitudes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	palabra, ok := vars["palabra"]
	if !ok || palabra == "" {
		http.Error(w, "palabra no proporcionada", http.StatusBadRequest)
		return
	}

	jsonInfo, err := os.ReadFile("datos/salida/recortes.json")
	if err != nil {
		http.Error(w, "error leyendo datos", http.StatusInternalServerError)
		return
	}

	_, resultadoAnalisis, err := analisis.Similitudes(jsonInfo, palabra)
	if err != nil {
		http.Error(w, "error buscando similitudes", http.StatusInternalServerError)
		return
	}

	// Ahora resultadoAnalisis es *modelos.ResultadoAnalisis, accedemos directo
	resultadoPalabra, ok := resultadoAnalisis.Busquedas[palabra]
	if !ok {
		// No hay resultados para esa palabra
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	// respuesta: solo enviamos resultadoPalabra.Resultados
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resultadoPalabra.Resultados); err != nil {
		http.Error(w, "error enviando resultados", http.StatusInternalServerError)
	}
}

