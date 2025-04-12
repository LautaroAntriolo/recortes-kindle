package escritura

import (
	"encoding/json"
	"fmt"
	"os"
)

// EscribirJSON escribe los recortes en un archivo JSON.
func EscribirJSON(filePath string, recortes interface{}) ([]byte, error) {
	// Convertir los recortes a JSON formateado
	jsonData, err := json.MarshalIndent(recortes, "", "  ")
	if err != nil {
		return nil, err
	}

	// Crear el archivo
	file, err := os.Create(filePath)
	if err != nil {
		return jsonData, err // Devolvemos el jsonData aunque haya error por si quiere usarse
	}
	defer file.Close()

	// Escribir el JSON en el archivo
	_, err = file.Write(jsonData)
	if err != nil {
        fmt.Println("Error al escribir el JSON")
		return jsonData, err // Devolvemos el jsonData aunque haya error
	}

	return jsonData, nil
}
