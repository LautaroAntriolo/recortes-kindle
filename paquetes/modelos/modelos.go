package modelos

import "time"

// ResultadoAnalisis representa la estructura completa del JSON resultante
type ResultadoAnalisis struct {
	Metadata   Metadata            `json:"metadata"`
	Indices    Indices             `json:"indices"`
	Busquedas  map[string]Busqueda `json:"busquedas,omitempty"`
	Documentos []Documento         `json:"documentos,omitempty"` // Solo se incluye cuando no hay búsquedas
}

// Metadata contiene información sobre el análisis realizado
type Metadata struct {
	Version                string    `json:"version"`
	FechaAnalisis          time.Time `json:"fecha_analisis"`
	TotalDocumentos        int       `json:"total_documentos"`
	DocumentosConEtiquetas int       `json:"documentos_con_etiquetas"`
}

// Indices contiene los diferentes índices generados
type Indices struct {
	PorAutor    map[string][]int `json:"por_autor"`
	PorLibro    map[string][]int `json:"por_libro"`
	PorEtiqueta map[string][]int `json:"por_etiqueta"`
}

// Busqueda representa los resultados de una búsqueda específica
type Busqueda struct {
	Resultados         []Documento `json:"resultados"`
	TotalCoincidencias int         `json:"total_coincidencias"`
}

// Documento representa la estructura base de cada documento
type Documento struct {
	ID          int      `json:"id"`
	Autor       string   `json:"autor"`
	Nombre      string   `json:"nombre"`
	Contenido   string   `json:"contenido"`
	Pagina      int      `json:"pagina,omitempty"`
	Visibilidad bool     `json:"visibilidad,omitempty"`
	Fecha       string   `json:"fecha,omitempty"`
	Hora        string   `json:"hora,omitempty"`
	Etiquetas   []string `json:"etiquetas,omitempty"`
}

type Documentos []Documento

type Registro struct {
	ID          int    `json:"id"`
	Autor       string `json:"autor"`
	Nombre      string `json:"nombre"`
	Pagina      int    `json:"pagina"`
	Contenido   string `json:"contenido"`
	Visibilidad bool   `json:"visibilidad"`
	Fecha       string `json:"fecha"`
	Hora        string `json:"hora"`
}

type Registros []Registro
