package analisis

import (
	"encoding/json"
	"fmt"
	"recortesKindle/paquetes/modelos"
	"regexp"
	"strings"
	"time"
)

// Función auxiliar para crear la estructura de resultados con metadata e índices

func crearEstructuraResultado(documentos []modelos.Documento) modelos.ResultadoAnalisis {
	// Creamos la estructura manualmente
	resultado := modelos.ResultadoAnalisis{
		Metadata: modelos.Metadata{
			Version:                "1.0",
			FechaAnalisis:          time.Now(),
			TotalDocumentos:        len(documentos),
			DocumentosConEtiquetas: contarDocumentosConEtiquetas(documentos),
		},
		Indices: modelos.Indices{
			PorAutor:    make(map[string][]int),
			PorLibro:    make(map[string][]int),
			PorEtiqueta: make(map[string][]int),
		},
		Busquedas: make(map[string]modelos.Busqueda),
	}

	// Generar índices
	for _, doc := range documentos {
		// Índice por autor
		resultado.Indices.PorAutor[doc.Autor] = append(resultado.Indices.PorAutor[doc.Autor], doc.ID)

		// Índice por libro
		resultado.Indices.PorLibro[doc.Nombre] = append(resultado.Indices.PorLibro[doc.Nombre], doc.ID)

		// Índice por etiquetas
		if doc.Etiquetas != nil {
			for _, etiqueta := range doc.Etiquetas {
				resultado.Indices.PorEtiqueta[etiqueta] = append(resultado.Indices.PorEtiqueta[etiqueta], doc.ID)
			}
		}
	}

	return resultado
}

func contarDocumentosConEtiquetas(documentos []modelos.Documento) int {
	count := 0
	for _, doc := range documentos {
		if doc.Etiquetas != nil && len(doc.Etiquetas) > 0 {
			count++
		}
	}
	return count
}
func matchInSlice(regex *regexp.Regexp, items []string) bool {
    for _, item := range items {
        if regex.MatchString(strings.ToLower(item)) {
            return true
        }
    }
    return false
}
func Similitudes(jsonInfo []byte, terminoBusqueda string) ([]modelos.Documento, *modelos.ResultadoAnalisis, error) {
	// Deserializar el JSON de entrada
	var documentos []modelos.Documento
	if err := json.Unmarshal(jsonInfo, &documentos); err != nil {
		return nil, nil, fmt.Errorf("error al deserializar JSON: %v", err)
	}

	// Crear estructura base del resultado
	resultado := modelos.ResultadoAnalisis{
		Metadata: modelos.Metadata{
			Version:       "1.0",
			FechaAnalisis: time.Now(),
		},
		Indices: modelos.Indices{
			PorAutor:    make(map[string][]int),
			PorLibro:    make(map[string][]int),
			PorEtiqueta: make(map[string][]int),
		},
	}

	// Si no hay término de búsqueda, devolver todo
	if strings.TrimSpace(terminoBusqueda) == "" {
		resultado.Documentos = documentos
		resultado.Metadata.TotalDocumentos = len(documentos)
		resultado.Metadata.DocumentosConEtiquetas = contarDocumentosConEtiquetas(documentos)
		generarIndicesCompletos(&resultado, documentos)
		return documentos, &resultado, nil
	}

	// Preparar la expresión regular para búsqueda exacta
	termino := strings.ToLower(strings.TrimSpace(terminoBusqueda))
	regex, err := regexp.Compile(fmt.Sprintf(`(?i)\b%s\b`, regexp.QuoteMeta(termino)))
	if err != nil {
		return nil, nil, fmt.Errorf("error en término de búsqueda: %v", err)
	}

	// Filtrar documentos y generar índices solo para coincidencias
	var resultados []modelos.Documento
	for _, doc := range documentos {
		if regex.MatchString(strings.ToLower(doc.Autor)) ||
			regex.MatchString(strings.ToLower(doc.Nombre)) ||
			regex.MatchString(strings.ToLower(doc.Contenido)) || 
			matchInSlice(regex, doc.Etiquetas) {
			resultados = append(resultados, doc)

			// Agregar a índices solo si coincide
			resultado.Indices.PorAutor[doc.Autor] = append(resultado.Indices.PorAutor[doc.Autor], doc.ID)
			resultado.Indices.PorLibro[doc.Nombre] = append(resultado.Indices.PorLibro[doc.Nombre], doc.ID)

			if doc.Etiquetas != nil {
				for _, etiqueta := range doc.Etiquetas {
					resultado.Indices.PorEtiqueta[etiqueta] = append(resultado.Indices.PorEtiqueta[etiqueta], doc.ID)
				}
			}
		}
	}

	// Actualizar metadata
	resultado.Metadata.TotalDocumentos = len(resultados)
	resultado.Metadata.DocumentosConEtiquetas = contarDocumentosConEtiquetas(resultados)

	// Configurar resultados de búsqueda
	if len(resultados) > 0 {
		resultado.Busquedas = map[string]modelos.Busqueda{
			terminoBusqueda: {
				Resultados:         resultados,
				TotalCoincidencias: len(resultados),
			},
		}
	} else {
		resultado.Busquedas = map[string]modelos.Busqueda{
			"no_results": {
				Resultados:         []modelos.Documento{},
				TotalCoincidencias: 0,
			},
		}
	}

	return resultados, &resultado, nil
}

// Función auxiliar para generar índices completos (solo cuando no hay búsqueda)
func generarIndicesCompletos(resultado *modelos.ResultadoAnalisis, documentos []modelos.Documento) {
	for _, doc := range documentos {
		resultado.Indices.PorAutor[doc.Autor] = append(resultado.Indices.PorAutor[doc.Autor], doc.ID)
		resultado.Indices.PorLibro[doc.Nombre] = append(resultado.Indices.PorLibro[doc.Nombre], doc.ID)
		if doc.Etiquetas != nil {
			for _, etiqueta := range doc.Etiquetas {
				resultado.Indices.PorEtiqueta[etiqueta] = append(resultado.Indices.PorEtiqueta[etiqueta], doc.ID)
			}
		}
	}
}

func SimilitudEnLaPalabra(jsonInfo []byte, terminosBusqueda ...string) (map[string][]modelos.Documento, *modelos.ResultadoAnalisis, error) {
	// Deserializar el JSON de entrada
	var documentos []modelos.Documento
	if err := json.Unmarshal(jsonInfo, &documentos); err != nil {
		return nil, nil, fmt.Errorf("error al deserializar JSON: %v", err)
	}

	// Crear estructura base del resultado
	resultado := crearEstructuraResultado(documentos)
	resultadoFinal := make(map[string][]modelos.Documento)

	// Filtrar términos válidos
	var terminosValidos []string
	for _, termino := range terminosBusqueda {
		if t := strings.TrimSpace(termino); t != "" {
			terminosValidos = append(terminosValidos, t)
		}
	}

	// Si no hay términos válidos, devolver todo
	if len(terminosValidos) == 0 {
		resultado.Documentos = documentos
		resultadoFinal["todos"] = documentos
		return resultadoFinal, &resultado, nil
	}

	// Preparar expresiones regulares para cada término
	regexes := make(map[string]*regexp.Regexp)
	for _, termino := range terminosValidos {
		regex, err := regexp.Compile(fmt.Sprintf(`(?i)\b%s\b`, regexp.QuoteMeta(termino)))
		if err != nil {
			return nil, nil, fmt.Errorf("error en término de búsqueda '%s': %v", termino, err)
		}
		regexes[termino] = regex
	}

	// Buscar documentos que coincidan con cada término
	busquedas := make(map[string]modelos.Busqueda)
	for termino, regex := range regexes {
		var resultados []modelos.Documento
		for _, doc := range documentos {
			if regex.MatchString(strings.ToLower(doc.Autor)) ||
				regex.MatchString(strings.ToLower(doc.Nombre)) ||
				regex.MatchString(strings.ToLower(doc.Contenido)) {
				resultados = append(resultados, doc)
			}
		}
		if len(resultados) > 0 {
			resultadoFinal[termino] = resultados
			busquedas[termino] = modelos.Busqueda{
				Resultados:         resultados,
				TotalCoincidencias: len(resultados),
			}
		}
	}

	// Configurar resultados de búsqueda
	if len(busquedas) > 0 {
		resultado.Busquedas = busquedas
	} else {
		resultado.Busquedas = map[string]modelos.Busqueda{
			"no_results": {
				Resultados:         []modelos.Documento{},
				TotalCoincidencias: 0,
			},
		}
		resultadoFinal["no_results"] = []modelos.Documento{}
	}

	return resultadoFinal, &resultado, nil
}
