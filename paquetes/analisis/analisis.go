package analisis

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Estructura que representa cada elemento del JSON
type Nota struct {
	ID          int     `json:"id"`
	Autor       string  `json:"autor"`
	Nombre      string  `json:"nombre"`
	Pagina      int     `json:"pagina"`
	Contenido   string  `json:"contenido"`
	Visibilidad bool    `json:"visibilidad"`
	Fecha       string  `json:"fecha"`
	Hora        string  `json:"hora"`
	Etiquetas   *string `json:"etiquetas"`
}

func Similitudes(jsonData []byte, busqueda string) ([]Nota, []byte, error) {
	// Decodificar JSON original
	var notas []Nota
	if err := json.Unmarshal(jsonData, &notas); err != nil {
		return nil, nil, fmt.Errorf("error decodificando JSON: %v", err)
	}

	// Filtrar resultados (búsqueda case-insensitive)
	var resultados []Nota
	busqueda = strings.ToLower(busqueda)

	for _, nota := range notas {
		if strings.Contains(strings.ToLower(nota.Autor), busqueda) ||
			strings.Contains(strings.ToLower(nota.Nombre), busqueda) ||
			strings.Contains(strings.ToLower(nota.Contenido), busqueda) {
			resultados = append(resultados, nota)
		}
	}

	// Convertir resultados a JSON
	jsonResultados, err := json.MarshalIndent(resultados, "", "  ")
	if err != nil {
		return resultados, nil, fmt.Errorf("error generando JSON: %v", err)
	}

	return resultados, jsonResultados, nil
}
