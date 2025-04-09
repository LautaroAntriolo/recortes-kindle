package proceso

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Recorte representa un recorte procesado.
type Recorte struct {
	ID          int       `json:"id"`
	Autor       string    `json:"autor"`
	Nombre      string    `json:"nombre"`
	Pagina      int       `json:"pagina"`
	Contenido   string    `json:"contenido"`
	Visibilidad bool      `json:"visibilidad"`
	FechaStr    string    `json:"fecha"` // Fecha formateada como string (YYYY-MM-DD)
	HoraStr     string    `json:"hora"`  // Hora formateada como string (HH:MM:SS)
	DateTime    time.Time `json:"-"`     // Campo interno para cálculos (no se serializa a JSON)
}

// ProcesoDeLineas convierte las líneas en una lista de recortes ordenada.
func ProcesoDeLineas(lines []string) ([]Recorte, error) {
	var recortes []Recorte
	var currentRecorte Recorte
	currentID := 1 // Contador para los IDs

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "==========") {
			if currentRecorte.Autor != "" && currentRecorte.Contenido != "" {
				currentRecorte.ID = currentID
				recortes = append(recortes, currentRecorte)
				currentID++
				currentRecorte = Recorte{
					Visibilidad: true,
				}
			}
			continue
		}

		// Procesar la línea con el título y autor
		if strings.Contains(line, "(") && strings.Contains(line, ")") {
			parts := strings.Split(line, "(")
			if len(parts) > 1 {
				currentRecorte.Nombre = strings.TrimSpace(parts[0])
				autorPart := strings.Split(parts[1], ")")
				currentRecorte.Autor = strings.TrimSpace(autorPart[0])
			}
			continue
		}

		// Procesar la línea con página y fecha
		if strings.Contains(line, "Tu recorte en la página") || strings.Contains(line, "- Tu recorte en la página") {
			// Extraer número de página
			pageParts := strings.Split(line, "página ")
			if len(pageParts) > 1 {
				pageInfo := strings.Split(pageParts[1], " |")
				if len(pageInfo) > 0 {
					pageStr := strings.TrimSpace(pageInfo[0])
					if pageNum, err := strconv.Atoi(pageStr); err == nil {
						currentRecorte.Pagina = pageNum
					}
				}
			}

			// Extraer fecha y hora
			dateParts := strings.Split(line, "Añadido el ")
			if len(dateParts) > 1 {
				dateTimeStr := strings.TrimSpace(dateParts[1])
				dateTimeParts := strings.Split(dateTimeStr, " a las ")
				if len(dateTimeParts) == 2 {
					currentRecorte.FechaStr = strings.TrimSpace(dateTimeParts[0])
					currentRecorte.HoraStr = strings.TrimSpace(dateTimeParts[1])
					
					// Parsear fecha y hora completa para DateTime
					layout := "2006-01-02 15:04:05"
					fechaHora := fmt.Sprintf("%s %s", currentRecorte.FechaStr, currentRecorte.HoraStr)
					if t, err := time.Parse(layout, fechaHora); err == nil {
						currentRecorte.DateTime = t
					}
				}
			}
			continue
		}

		// Procesar el contenido del recorte
		if currentRecorte.Autor != "" && line != "" {
			if currentRecorte.Contenido == "" {
				currentRecorte.Contenido = line
			} else {
				currentRecorte.Contenido += "\n" + line
			}
		}
	}

	// Añadir el último recorte si existe
	if currentRecorte.Autor != "" && currentRecorte.Contenido != "" {
		currentRecorte.ID = currentID
		recortes = append(recortes, currentRecorte)
	}

	return recortes, nil
}