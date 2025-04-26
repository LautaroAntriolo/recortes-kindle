package main
// go:build windows
// +build windows
//Subir estas dos lineas al inicio del archivo para evitar errores de compilación en Windows
import (
	"fmt"
	"log"
	"net/http"
	"recortesKindle/paquetes/escritura"
	"recortesKindle/paquetes/lectura"
	"recortesKindle/paquetes/proceso"
	"recortesKindle/rutas"

	"github.com/gorilla/mux"
)

func main() {
	// Determinar qué archivo de entrada usar
	archivoEntrada := "datos/entrada/recortes.txt" // Valor por defecto
	archivoSalida := "datos/salida/notas.json"     // Valor por defecto

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
	escritura.EscribirJSON(archivoSalida, recortes)

    r := mux.NewRouter()

    r.HandleFunc("/", rutas.Inicio)
    r.HandleFunc("/data", rutas.DataHandler)
    r.HandleFunc("/similitudes/{palabra}", rutas.Similitudes)

    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    fmt.Println("🚀 Servidor iniciado en http://localhost:8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatal(err)
    }
}
