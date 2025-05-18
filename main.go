//go:build windows
// +build windows

package main

// go:build windows
//Subir estas dos lineas al inicio del archivo para evitar errores de compilación en Windows
import (
	"fmt"
	"log"
	"net/http"
	"recortesKindle/rutas"

	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	//chequedas
	r.HandleFunc("/", rutas.Inicio)
	r.HandleFunc("/cargar-archivo", rutas.CargarArchivoHandler).Methods("POST")
	r.HandleFunc("/api/bases-de-datos", rutas.ListarBasesDeDatosHandler).Methods("GET")
	r.HandleFunc("/api/datos", rutas.ObtenerDatosHandler).Methods("GET")

	fmt.Println("🚀 Servidor iniciado en http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
