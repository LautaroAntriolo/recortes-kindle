package proceso

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Recorte representa un recorte procesado.
type Recorte struct {
	Autor       string    `json:"autor"`
	Nombre      string    `json:"nombre"`
	Pagina      int       `json:"pagina"`
	Contenido   string    `json:"contenido"`
	Visibilidad bool      `json:"visibilidad"`
	FechaStr    string    `json:"fecha"` // Fecha formateada como string (YYYY-MM-DD)
	HoraStr     string    `json:"hora"`  // Hora formateada como string (HH:MM:SS)
	DateTime    time.Time `json:"-"`     // Campo interno para cálculos (no se serializa a JSON)
}

// mesesEspañol mapea nombres de meses en español a sus equivalentes numéricos
var mesesEspañol = map[string]string{
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

// parseFechaEspañol convierte una fecha en formato español a time.Time
func parseFechaEspañol(fechaStr string) (time.Time, string, string, error) {
	// Eliminar "Añadido el " si está presente
	fechaStr = strings.TrimPrefix(fechaStr, "Añadido el ")

	// Dividir la fecha en sus componentes
	// Ejemplo: "martes, 14 de enero de 2025 20:55:34"
	partes := strings.Split(fechaStr, ", ")
	if len(partes) < 2 {
		return time.Time{}, "", "", fmt.Errorf("formato de fecha inválido: %s", fechaStr)
	}

	// Ignoramos el día de la semana y procesamos el resto
	restoFecha := partes[1]

	// Dividir en fecha y hora
	partesDateTime := strings.Split(restoFecha, " ")
	if len(partesDateTime) < 5 {
		return time.Time{}, "", "", fmt.Errorf("formato de fecha y hora inválido: %s", restoFecha)
	}

	dia := partesDateTime[0]
	// Ignoramos "de"
	mes := partesDateTime[2]
	// Ignoramos "de"
	año := partesDateTime[4]
	hora := ""
	if len(partesDateTime) > 5 {
		hora = partesDateTime[5]
	}

	// Convertir mes de texto a número
	mesNum, ok := mesesEspañol[strings.ToLower(mes)]
	if !ok {
		return time.Time{}, "", "", fmt.Errorf("mes inválido: %s", mes)
	}

	// Formatear la fecha en formato ISO 8601 pero sin la T y Z
	fechaISO := fmt.Sprintf("%s-%s-%s %s", año, mesNum, dia, hora)

	// Crear strings formateados para la fecha y la hora
	fechaStr = fmt.Sprintf("%s-%s-%s", año, mesNum, dia)
	horaStr := hora

	// Parsear la fecha formateada
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, fechaISO)
	if err != nil {
		return time.Time{}, "", "", err
	}

	return t, fechaStr, horaStr, nil
}

// ProcesoDeLineas convierte las líneas en una lista de recortes ordenada.
func ProcesoDeLineas(lines []string) ([]Recorte, error) {
	var recortes []Recorte
	var currentRecorte Recorte

	for _, line := range lines {
		// Si encontramos un separador, guardamos el recorte actual
		if strings.Contains(line, "==========") {
			if currentRecorte.Autor != "" {
				recortes = append(recortes, currentRecorte)
				currentRecorte = Recorte{
					Visibilidad: true, // Valor predeterminado para Visibilidad
				}
			}
			continue
		}

		// Procesar la línea con el título y autor
		if strings.HasPrefix(line, "- La subrayado") {
			parts := strings.Split(line, "en la página ")
			if len(parts) > 1 {
				pageInfo := strings.Split(parts[1], " ")
				if len(pageInfo) > 0 {
					pageStr := strings.TrimSpace(pageInfo[0])
					pageStr = strings.Split(pageStr, "|")[0] // Por si hay un separador
					currentRecorte.Pagina, _ = strconv.Atoi(pageStr)
				}
			}

			// Extraer la fecha de la línea
			dateParts := strings.Split(line, "Añadido el ")
			if len(dateParts) > 1 {
				dateStr := strings.TrimSpace(dateParts[1])
				parsedTime, fechaStr, horaStr, err := parseFechaEspañol(dateStr)
				if err == nil {
					currentRecorte.DateTime = parsedTime
					currentRecorte.FechaStr = fechaStr
					currentRecorte.HoraStr = horaStr
				}
			}
			continue
		}

		// Procesar la información del autor y nombre
		if currentRecorte.Autor == "" {
			parts := strings.Split(line, "(")
			if len(parts) > 1 {
				currentRecorte.Nombre = strings.TrimSpace(parts[0])
				currentRecorte.Autor = strings.TrimRight(parts[1], ")")
			}
			continue
		}

		// Procesar el contenido del recorte
		if currentRecorte.Autor != "" && currentRecorte.Contenido == "" {
			currentRecorte.Contenido = strings.TrimSpace(line)
		}
	}

	// Añadir el último recorte si existe
	if currentRecorte.Autor != "" {
		recortes = append(recortes, currentRecorte)
	}

	return recortes, nil
}
