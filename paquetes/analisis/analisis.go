package analisis

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Estructura que representa cada elemento del JSON
type Documento struct {
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

func Similitudes(jsonInfo []byte, terminoBusqueda string) ([]Documento, []byte, error) {
	// Deserializar el JSON de entrada
	var documentos []Documento
	if err := json.Unmarshal(jsonInfo, &documentos); err != nil {
		return nil, nil, fmt.Errorf("error al deserializar JSON: %v", err)
	}

	// Si no hay término de búsqueda, devolver todo con formato
	if strings.TrimSpace(terminoBusqueda) == "" {
		jsonFormateado, err := json.MarshalIndent(documentos, "", "    ")
		if err != nil {
			return nil, nil, fmt.Errorf("error al formatear JSON: %v", err)
		}
		return documentos, jsonFormateado, nil
	}

	// Preparar la expresión regular para búsqueda exacta
	termino := strings.ToLower(strings.TrimSpace(terminoBusqueda))
	regex, err := regexp.Compile(fmt.Sprintf(`(?i)\b%s\b`, regexp.QuoteMeta(termino)))
	if err != nil {
		return nil, nil, fmt.Errorf("error en término de búsqueda: %v", err)
	}

	// Filtrar documentos
	var resultados []Documento
	for _, doc := range documentos {
		if regex.MatchString(strings.ToLower(doc.Autor)) ||
			regex.MatchString(strings.ToLower(doc.Nombre)) ||
			regex.MatchString(strings.ToLower(doc.Contenido)) {
			resultados = append(resultados, doc)
		}
	}

	// Generar JSON con indentación
	jsonResultados, err := json.MarshalIndent(resultados, "", " ")
	if err != nil {
		return nil, nil, fmt.Errorf("error al generar JSON: %v", err)
	}

	return resultados, jsonResultados, nil
}
