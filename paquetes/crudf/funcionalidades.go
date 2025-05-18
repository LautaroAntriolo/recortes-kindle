package crudf

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"recortesKindle/paquetes/modelos"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func Favoritos(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	nombreArchivo := strings.TrimSuffix(params["nombre"], ".txt") + ".json"
	idStr := params["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	rutaArchivo := filepath.Join("dato", "salida", nombreArchivo)

	switch r.Method {
	case "PUT", "POST":
		fileData, err := os.ReadFile(rutaArchivo)
		if err != nil {
			http.Error(w, "Error al leer el archivo", http.StatusInternalServerError)
			return
		}

		var recortes []modelos.Recorte
		if err := json.Unmarshal(fileData, &recortes); err != nil {
			http.Error(w, "Error al decodificar JSON", http.StatusInternalServerError)
			return
		}

		var favoritoActualizado bool
		encontrado := false
		for i := range recortes {
			if recortes[i].ID == id {
				recortes[i].Favorito = !recortes[i].Favorito
				favoritoActualizado = recortes[i].Favorito
				encontrado = true
				break
			}
		}

		if !encontrado {
			http.Error(w, "Recorte no encontrado", http.StatusNotFound)
			return
		}

		updatedData, err := json.MarshalIndent(recortes, "", "  ")
		if err != nil {
			http.Error(w, "Error al codificar JSON", http.StatusInternalServerError)
			return
		}

		if err := os.WriteFile(rutaArchivo, updatedData, 0644); err != nil {
			http.Error(w, "Error al guardar el archivo", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			ID       int  `json:"id"`
			Favorito bool `json:"favorito"`
		}{
			ID:       id,
			Favorito: favoritoActualizado,
		})

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}
