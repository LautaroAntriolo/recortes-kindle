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
	r.HandleFunc("/", rutas.Inicio).Methods("GET")
	r.HandleFunc("/data", rutas.DataHandler).Methods("GET")
	r.HandleFunc("/mostrar-archivos", rutas.MostrarArchivos).Methods("GET")
	r.HandleFunc("/cargar-archivo", rutas.CargarArchivoHandler).Methods("POST")

	//No chequedas
	r.HandleFunc("/similitudes/{palabra}", rutas.Similitudes).Methods("GET")
	r.HandleFunc("/archivo/{nombre}", rutas.ObtenerArchivoHandler).Methods("GET")
	

	fmt.Println("🚀 Servidor iniciado en http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
