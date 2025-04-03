package escritura

import (
    "encoding/json"
    "os"
)

// WriteJSON escribe los recortes en un archivo JSON.
func EscribirJSON(filePath string, recortes interface{}) error {
    // Convertir los recortes a JSON formateado
    jsonData, err := json.MarshalIndent(recortes, "", "  ")
    if err != nil {
        return err
    }

    // Crear el archivo
    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Escribir el JSON en el archivo
    _, err = file.Write(jsonData)
    if err != nil {
        return err
    }

    return nil
}