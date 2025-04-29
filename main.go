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
	// Determinar qué archivo de entrada usar
	// archivoEntrada := "datos/entrada/recortes.txt" // Valor por defecto
	// archivoSalida := "datos/salida/notas.json" // Valor por defecto

	// Leer el archivo de texto
	// lines, err := lectura.LeerArchivo(archivoEntrada)
	// if err != nil {
	// 	log.Fatalf("Error al leer el archivo: %v", err)
	// }

	// // Procesar los recortes
	// recortes, err := proceso.ProcesoDeLineas(lines)
	// if err != nil {
	// 	log.Fatalf("Error al procesar los recortes: %v", err)
	// }

	// Escribir los recortes en un archivo JSON
	// escritura.EscribirJSON(archivoSalida, recortes)

	r := mux.NewRouter()

	r.HandleFunc("/", rutas.Inicio).Methods("GET")
	r.HandleFunc("/data", rutas.DataHandler).Methods("GET")
	r.HandleFunc("/cargar-archivo", rutas.CargarArchivoHandler).Methods("POST")
	r.HandleFunc("/similitudes/{palabra}", rutas.Similitudes).Methods("GET")
	r.HandleFunc("/archivos", rutas.MostrarArchivos).Methods("GET")
	r.HandleFunc("/archivo/{nombre}", rutas.ObtenerArchivoHandler).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("🚀 Servidor iniciado en http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
