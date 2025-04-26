package rutas

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"recortesKindle/paquetes/analisis"
	"recortesKindle/paquetes/modelos"
	"text/template"

	"github.com/gorilla/mux"
)

func Inicio(w http.ResponseWriter, r *http.Request) {
	docs := []modelos.Documento{}
	// Cargar y ejecutar la plantilla
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Error al cargar plantilla: %v", err)
	}
	err = t.Execute(w, docs)
	if err != nil {
		http.Error(w, "Error al renderizar plantilla", http.StatusInternalServerError)
		log.Printf("Error al ejecutar plantilla: %v", err)
	}
}

func DataHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("datos/salida/notas.json")
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

	jsonInfo, err := os.ReadFile("datos/salida/notas.json")
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
