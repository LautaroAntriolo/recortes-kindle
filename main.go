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
	"recortesKindle/paquetes/modelos"
	"recortesKindle/paquetes/proceso"
)

func main() {
	// Determinar qué archivo de entrada usar
	archivoEntrada := "datos/entrada/recortes.txt" // Valor por defecto
	archivoSalida := "datos/salida/notas.json"     // Valor por defecto
	terminoBusqueda := ""

	// //Si se proporciona un argumento, usarlo como ruta del archivo
	// if len(os.Args) > 1 {
	// 	archivoEntrada = os.Args[1]
	// }

	// //Si se proporciona un segundo argumento, usarlo como resultado de salida
	// if len(os.Args) > 2 {
	// 	// archivoSalida = os.Args[2]
	// 	terminoBusqueda = os.Args[2]
	// }

	// //Si se proporciona un tercer argumento, usarlo como termino de busqueda
	// if len(os.Args) > 3 {
	// 	terminoBusqueda = os.Args[3]
	// }

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
	docs := []modelos.Documento{}
	if terminoBusqueda != "" {
		// A partir de acá comienzo a realizar las funcionalidades desde el html.
		resultados, resultadoAnalisis, err := analisis.Similitudes(jsonInfo, terminoBusqueda)
		if err != nil {
			log.Fatal("Error buscando coincidencias:", err)
		}

		// Convertir el resultado completo a JSON
		jsonCompleto, err := json.MarshalIndent(resultadoAnalisis, "", "    ")
		if err != nil {
			log.Fatal("Error al generar JSON:", err)
		}

		if err := os.WriteFile("similitudes/por_palabra/"+terminoBusqueda+"_resultado.json", jsonCompleto, 0644); err != nil {
			log.Fatal("Error guardando resultados:", err)
		}

		fmt.Printf("✔ Se encontraron %d resultados\n✔ Guardados en: similitudes\n", len(resultados))

		var docs []modelos.Documento
		err = json.Unmarshal(jsonCompleto, &docs)
		if err != nil {
			// Verifica si es un solo documento
			var singleDoc modelos.Documento
			if err := json.Unmarshal(jsonCompleto, &singleDoc); err != nil {
				log.Fatalf("Error al deserializar JSON: %v", err)
			}
			docs = []modelos.Documento{singleDoc} // Convierte a slice con un elemento
		}
	}

	// Cargar y ejecutar la plantilla
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Error al cargar plantilla: %v", err)
	}

	// Configurar el manejador HTTP
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := t.Execute(w, docs)
		if err != nil {
			http.Error(w, "Error al renderizar plantilla", http.StatusInternalServerError)
			log.Printf("Error al ejecutar plantilla: %v", err)
		}
	})

	// Configurar archivos estáticos (opcional, si tienes CSS/JS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("🚀 Servidor iniciado en http://localhost:8080")
	fmt.Println("Presiona Ctrl+C para detenerlo")

	// Iniciar el servidor
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}

}
