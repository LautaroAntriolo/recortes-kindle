package modelos

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