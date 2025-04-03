package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"recortesKindle/paquetes/escritura"
	"recortesKindle/paquetes/lectura"
	"recortesKindle/paquetes/proceso"
	"strconv"
	"strings"
)

// Estructura para el JSON
type Registro struct {
	Autor       string `json:"autor"`
	Nombre      string `json:"nombre"`
	Pagina      int    `json:"pagina"`
	Contenido   string `json:"contenido"`
	Visibilidad bool   `json:"visibilidad"`
	Fecha       string `json:"fecha"`
	Hora        string `json:"hora"`
}

func main() {
	// Leer el archivo de texto
	lines, err := lectura.LeerArchivo("misRecortes.txt")
	if err != nil {
		log.Fatalf("Error al leer el archivo: %v", err)
	}

	// Procesar los recortes
	recortes, err := proceso.ProcesoDeLineas(lines)
	if err != nil {
		log.Fatalf("Error al procesar los recortes: %v", err)
	}

	// Escribir los recortes en un archivo JSON
	err = escritura.EscribirJSON("notas.json", recortes)
	if err != nil {
		log.Fatalf("Error al escribir el archivo JSON: %v", err)
	}

	// fmt.Println("El archivo notas.json ha sido creado exitosamente.")

	// Preguntar al usuario si quiere visualizar los registros
	fmt.Println("¿Desea visualizar los registros? (s/n):")
	reader := bufio.NewReader(os.Stdin)
	respuesta, _ := reader.ReadString('\n')
	respuesta = strings.TrimSpace(strings.ToLower(respuesta))

	if respuesta != "s" && respuesta != "si" && respuesta != "sí" {
		fmt.Println("Programa finalizado.")
		return
	}

	// Leer el archivo JSON generado para visualización
	visualizarRegistros("notas.json")
}

// Función para visualizar los registros de forma interactiva
func visualizarRegistros(archivoJSON string) {
	contenido, err := ioutil.ReadFile(archivoJSON)
	if err != nil {
		log.Fatalf("Error al leer el archivo JSON: %v", err)
	}

	// Deserializar JSON
	var registros []Registro
	err = json.Unmarshal(contenido, &registros)
	if err != nil {
		log.Fatalf("Error al deserializar JSON: %v", err)
	}

	if len(registros) == 0 {
		fmt.Println("No se encontraron registros.")
		return
	}

	// Interfaz por línea de comandos
	indiceActual := 0
	salir := false
	reader := bufio.NewReader(os.Stdin)

	for !salir {
		// Limpiar pantalla (aprox)
		fmt.Print("\033[H\033[2J")

		if indiceActual < 0 || indiceActual >= len(registros) {
			indiceActual = 0
		}

		registro := registros[indiceActual]

		// Mostrar datos del registro actual
		fmt.Println("==== REGISTRO", indiceActual+1, "DE", len(registros), "====")
		fmt.Println("Autor:", registro.Autor)
		fmt.Println("Nombre:", registro.Nombre)
		fmt.Println("Página:", registro.Pagina)
		fmt.Println("Contenido:", registro.Contenido)
		fmt.Println("Fecha:", registro.Fecha)
		fmt.Println("Hora:", registro.Hora)
		fmt.Printf("Visibilidad: %v\n", registro.Visibilidad)
		fmt.Println("===================================")
		fmt.Println("\nOpciones:")
		fmt.Println("A - Anterior registro")
		fmt.Println("S - Siguiente registro")
		fmt.Println("V - Cambiar visibilidad")
		fmt.Println("I - Ir a un registro específico")
		fmt.Println("G - Guardar cambios")
		fmt.Println("X - Salir")
		fmt.Print("\nElija una opción: ")

		opcion, _ := reader.ReadString('\n')
		opcion = strings.TrimSpace(strings.ToUpper(opcion))

		switch opcion {
		case "A":
			if indiceActual > 0 {
				indiceActual--
			} else {
				fmt.Println("Ya está en el primer registro.")
				esperarTecla()
			}
		case "S":
			if indiceActual < len(registros)-1 {
				indiceActual++
			} else {
				fmt.Println("Ya está en el último registro.")
				esperarTecla()
			}
		case "V":
			registros[indiceActual].Visibilidad = !registros[indiceActual].Visibilidad
			fmt.Printf("Visibilidad cambiada a: %v\n", registros[indiceActual].Visibilidad)
			esperarTecla()
		case "I":
			fmt.Print("Ingrese el número de registro (1-", len(registros), "): ")
			numStr, _ := reader.ReadString('\n')
			numStr = strings.TrimSpace(numStr)
			num, err := strconv.Atoi(numStr)
			if err != nil || num < 1 || num > len(registros) {
				fmt.Println("Número inválido.")
			} else {
				indiceActual = num - 1
			}
		case "G":
			// Serializar y guardar cambios
			jsonBytes, err := json.MarshalIndent(registros, "", "  ")
			if err != nil {
				log.Println("Error al serializar JSON:", err)
				esperarTecla()
				continue
			}
			err = ioutil.WriteFile(archivoJSON, jsonBytes, 0644)
			if err != nil {
				log.Println("Error al guardar el archivo:", err)
				esperarTecla()
				continue
			}
			fmt.Println("Cambios guardados exitosamente.")
			esperarTecla()
		case "X":
			salir = true
		default:
			fmt.Println("Opción no válida.")
			esperarTecla()
		}
	}
}

// Función auxiliar para esperar a que el usuario presione Enter
func esperarTecla() {
	fmt.Print("\nPresione Enter para continuar...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}