package proceso

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Etiqueta struct {
	Nombre string `json:"nombre"` // Mejor usar nombre en español en lugar de eti_nombre
}

type Etiquetas []Etiqueta // Esto está bien, es un slice de Etiqueta

// Recorte representa un recorte procesado.
type Recorte struct {
	ID          int       `json:"id"`
	Autor       string    `json:"autor"`
	Nombre      string    `json:"nombre"`
	Pagina      int       `json:"pagina"`
	Contenido   string    `json:"contenido"`
	Visibilidad bool      `json:"visibilidad"`
	FechaStr    string    `json:"fecha"`     // Fecha formateada como string (YYYY-MM-DD)
	HoraStr     string    `json:"hora"`      // Hora formateada como string (HH:MM:SS)
	DateTime    time.Time `json:"-"`         // Campo interno para cálculos (no se serializa a JSON)
	Etiquetas   Etiquetas `json:"etiquetas"` // Ahora es un slice de Etiqueta
}

// ProcesoDeLineas convierte las líneas en una lista de recortes ordenada.
// ProcesoDeLineas convierte las líneas en una lista de recortes ordenada.
func ProcesoDeLineas(lines []string) ([]Recorte, error) {
	var recortes []Recorte
	var currentRecorte Recorte
	currentID := 1 // Contador para los IDs
	inContent := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Si encontramos el separador, finalizamos el recorte actual y comenzamos uno nuevo
		if strings.Contains(line, "==========") {
			if currentRecorte.Autor != "" {
				currentRecorte.ID = currentID
				currentRecorte.Visibilidad = true
				recortes = append(recortes, currentRecorte)
				currentID++
				currentRecorte = Recorte{
					Visibilidad: true,
				}
			}
			inContent = false
			continue
		}

		// Procesar la línea con el nombre del libro y autor (siempre es la primera línea de un recorte)
		if !inContent && strings.Contains(line, "(") && strings.HasSuffix(line, ")") && currentRecorte.Autor == "" {
			parts := strings.Split(line, "(")
			if len(parts) > 1 {
				bookName := strings.TrimSpace(parts[0])
				authorPart := parts[len(parts)-1]
				author := strings.TrimSuffix(authorPart, ")")

				currentRecorte.Nombre = bookName
				currentRecorte.Autor = strings.TrimSpace(author)
			}
			continue
		}

		// Procesar la línea con página y fecha
		if !inContent && strings.Contains(line, "página") && (strings.Contains(line, "subrayado") || strings.Contains(line, "recorte")) {
			// Extraer número de página
			pageRegex := regexp.MustCompile(`página (\d+)`)
			pageMatches := pageRegex.FindStringSubmatch(line)
			if len(pageMatches) > 1 {
				if pageNum, err := strconv.Atoi(pageMatches[1]); err == nil {
					currentRecorte.Pagina = pageNum
				}
			}

			// Extraer fecha y hora
			if strings.Contains(line, "Añadido el") {
				dateRegex := regexp.MustCompile(`Añadido el ([^,]+, \d+ de [^ ]+ de \d+) (\d+:\d+:\d+)`)
				dateMatches := dateRegex.FindStringSubmatch(line)
				if len(dateMatches) > 2 {
					fechaStr := dateMatches[1]
					horaStr := dateMatches[2]

					currentRecorte.FechaStr = convertirFechaEspanolAISO(fechaStr)
					currentRecorte.HoraStr = horaStr

					// Parsear fecha y hora completa para DateTime
					layout := "2006-01-02 15:04:05"
					fechaHora := fmt.Sprintf("%s %s", currentRecorte.FechaStr, currentRecorte.HoraStr)
					if t, err := time.Parse(layout, fechaHora); err == nil {
						currentRecorte.DateTime = t
					}
				}
			}
			// Después de procesar la línea con página y fecha, las líneas siguientes son contenido
			inContent = true
			continue
		}

		// Procesar el contenido del recorte (después de la línea de página/fecha)
		if inContent && currentRecorte.Autor != "" {
			if currentRecorte.Contenido == "" {
				currentRecorte.Contenido = line
			} else {
				currentRecorte.Contenido += "\n" + line
			}
		}
	}

	// Añadir el último recorte si existe
	if currentRecorte.Autor != "" {
		currentRecorte.ID = currentID
		currentRecorte.Visibilidad = true
		recortes = append(recortes, currentRecorte)
	}

	return recortes, nil
}

// Función auxiliar para convertir fechas del formato español al formato ISO YYYY-MM-DD
func convertirFechaEspanolAISO(fechaEsp string) string {
	// Mapeo de nombres de meses en español a números
	mesesMap := map[string]string{
		"enero":      "01",
		"febrero":    "02",
		"marzo":      "03",
		"abril":      "04",
		"mayo":       "05",
		"junio":      "06",
		"julio":      "07",
		"agosto":     "08",
		"septiembre": "09",
		"octubre":    "10",
		"noviembre":  "11",
		"diciembre":  "12",
	}

	// Extraer componentes de la fecha
	re := regexp.MustCompile(`(\w+), (\d+) de (\w+) de (\d+)`)
	matches := re.FindStringSubmatch(fechaEsp)

	if len(matches) < 5 {
		return fechaEsp // Devolver la fecha original si no se puede procesar
	}

	dia := matches[2]
	mes := mesesMap[matches[3]]
	ano := matches[4]

	// Asegurarse de que el día tenga dos dígitos
	if len(dia) == 1 {
		dia = "0" + dia
	}

	// Formato ISO: YYYY-MM-DD
	return fmt.Sprintf("%s-%s-%s", ano, mes, dia)
}
