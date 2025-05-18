package rutas

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"recortesKindle/paquetes/lectura"
	"recortesKindle/paquetes/modelos"
	"recortesKindle/paquetes/proceso"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

const (
	uploadDir = "./datos/entrada"
	outputDir = "./datos/salida"
)

// Eliminamos la variable global db, cada handler manejará su propia conexión

func CargarArchivoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Procesar archivo subido
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("archivo")
	if err != nil {
		http.Error(w, "Error al obtener el archivo: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validar extensión
	if filepath.Ext(handler.Filename) != ".txt" {
		http.Error(w, "Solo se permiten archivos .txt", http.StatusBadRequest)
		return
	}

	// 2. Guardar archivo temporal
	filePath := filepath.Join(uploadDir, handler.Filename)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Error al crear directorio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error al guardar el archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error al copiar archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Procesar líneas
	lines, err := lectura.LeerArchivo(filePath)
	if err != nil {
		http.Error(w, "Error al leer archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	recortes, err := proceso.ProcesoDeLineas(lines)
	if err != nil {
		http.Error(w, "Error al procesar recortes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Guardar en SQLite específica para este archivo
	nombreDB := strings.TrimSuffix(handler.Filename, ".txt") + ".db"
	dbPath := filepath.Join(outputDir, nombreDB)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		http.Error(w, "Error al abrir DB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Crear tablas si no existen
	if err := crearTablas(db); err != nil {
		http.Error(w, "Error al crear tablas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Transacción para insertar datos
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Error al iniciar transacción: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := insertarRecortes(tx, recortes, handler.Filename); err != nil {
		tx.Rollback()
		http.Error(w, "Error al insertar datos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Error al confirmar transacción: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	json.NewEncoder(w).Encode(map[string]string{
		"mensaje":       "Archivo procesado correctamente",
		"base_de_datos": dbPath,
	})
}

// Función para crear tablas
func crearTablas(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS recortes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			autor TEXT,
			nombre TEXT,
			contenido TEXT,
			pagina INTEGER,
			visibilidad BOOLEAN DEFAULT 1,
			fecha TEXT,
			hora TEXT,
			nombreRecorte TEXT,
			favorito BOOLEAN DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS etiquetas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nombre TEXT UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS recortes_etiquetas (
			recorte_id INTEGER,
			etiqueta_id INTEGER,
			FOREIGN KEY (recorte_id) REFERENCES recortes(id) ON DELETE CASCADE,
			FOREIGN KEY (etiqueta_id) REFERENCES etiquetas(id) ON DELETE CASCADE,
			PRIMARY KEY (recorte_id, etiqueta_id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error al ejecutar query %q: %v", query, err)
		}
	}
	return nil
}

// Función para insertar recortes
func insertarRecortes(tx *sql.Tx, recortes []modelos.Recorte, nombreArchivo string) error {
	nombreBase := strings.TrimSuffix(nombreArchivo, ".txt")

	for _, recorte := range recortes {
		// Insertar recorte principal
		res, err := tx.Exec(`
			INSERT INTO recortes (
				autor, nombre, contenido, pagina, fecha, hora, 
				favorito, nombreRecorte, visibilidad
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			recorte.Autor,
			recorte.Nombre,
			recorte.Contenido,
			recorte.Pagina,
			recorte.FechaStr,
			recorte.HoraStr,
			recorte.Favorito,
			nombreBase,
			true, // visibilidad por defecto
		)
		if err != nil {
			return err
		}

		recorteID, err := res.LastInsertId()
		if err != nil {
			return err
		}

		// Insertar etiquetas
		for _, etiqueta := range recorte.Etiquetas {
			// Insertar etiqueta si no existe
			_, err := tx.Exec(
				"INSERT OR IGNORE INTO etiquetas (nombre) VALUES (?)",
				etiqueta.Nombre,
			)
			if err != nil {
				return err
			}

			// Relacionar recorte con etiqueta
			_, err = tx.Exec(
				`INSERT INTO recortes_etiquetas (recorte_id, etiqueta_id)
				VALUES (?, (SELECT id FROM etiquetas WHERE nombre = ?))`,
				recorteID,
				etiqueta.Nombre,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ObtenerDatosHandler modificado para usar DB específica
func ObtenerDatosHandler(w http.ResponseWriter, r *http.Request) {
	nombreArchivo := r.URL.Query().Get("archivo")
	if nombreArchivo == "" {
		http.Error(w, "Parámetro 'archivo' es requerido", http.StatusBadRequest)
		return
	}

	dbPath := filepath.Join("datos", "salida", nombreArchivo+".db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		http.Error(w, "Base de datos no encontrada", http.StatusNotFound)
		return
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		http.Error(w, "Error al abrir la base de datos", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Consulta optimizada para la vista
	rows, err := db.Query(`
        SELECT 
            r.id, r.autor, r.nombre, 
            substr(r.contenido, 1, 100) as contenido_preview,
            r.pagina,
            GROUP_CONCAT(e.nombre, ', ') as etiquetas
        FROM recortes r
        LEFT JOIN recortes_etiquetas re ON r.id = re.recorte_id
        LEFT JOIN etiquetas e ON re.etiqueta_id = e.id
        GROUP BY r.id
        ORDER BY r.id
    `)
	if err != nil {
		http.Error(w, "Error en consulta: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var resultados []struct {
		ID        int    `json:"id"`
		Autor     string `json:"autor"`
		Nombre    string `json:"nombre"`
		Contenido string `json:"contenido"`
		Pagina    int    `json:"pagina"`
		Etiquetas string `json:"etiquetas"`
	}

	for rows.Next() {
		var r struct {
			ID        int
			Autor     sql.NullString
			Nombre    sql.NullString
			Contenido sql.NullString
			Pagina    sql.NullInt64
			Etiquetas sql.NullString
		}

		if err := rows.Scan(&r.ID, &r.Autor, &r.Nombre, &r.Contenido, &r.Pagina, &r.Etiquetas); err != nil {
			http.Error(w, "Error al leer datos: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resultados = append(resultados, struct {
			ID        int    `json:"id"`
			Autor     string `json:"autor"`
			Nombre    string `json:"nombre"`
			Contenido string `json:"contenido"`
			Pagina    int    `json:"pagina"`
			Etiquetas string `json:"etiquetas"`
		}{
			ID:        r.ID,
			Autor:     r.Autor.String,
			Nombre:    r.Nombre.String,
			Contenido: r.Contenido.String,
			Pagina:    int(r.Pagina.Int64),
			Etiquetas: r.Etiquetas.String,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultados)
}

// ListarBasesDeDatosHandler devuelve la lista de archivos .db
func ListarBasesDeDatosHandler(w http.ResponseWriter, r *http.Request) {
	archivos, err := filepath.Glob("datos/salida/*.db") //Todos los que terminan en .db
	if err != nil {
		http.Error(w, "Error al leer archivos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extraer solo los nombres sin ruta ni extensión
	var nombres []string
	for _, archivo := range archivos {
		nombre := filepath.Base(archivo)
		nombre = strings.TrimSuffix(nombre, ".db")
		nombres = append(nombres, nombre)
	}

	// Devuelvo todo como json
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nombres)
}

// Estructura para pasar datos al template
type TablaData struct {
	Recortes []RecorteTabla
	Pagina   int
	Total    int
}

type RecorteTabla struct {
	ID        int
	Autor     string
	Nombre    string
	Contenido string
	Pagina    int
	Etiquetas []modelos.Etiqueta
}

func ConfigurarTablasRouter(r *mux.Router) {
	r.HandleFunc("/tabla", MostrarTablaHandler).Methods("GET")
	r.HandleFunc("/api/tabla", ObtenerDatosTablaHandler).Methods("GET")
}

func MostrarTablaHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"templates/index.html",
		"templates/tabla.html",
	)
	if err != nil {
		log.Printf("Error al parsear templates: %v", err)
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Error al renderizar template: %v", err)
		http.Error(w, "Error interno", http.StatusInternalServerError)
	}
}

func ObtenerDatosTablaHandler(w http.ResponseWriter, r *http.Request) {
	// Obtener parámetros
	nombreArchivo := r.URL.Query().Get("archivo")
	pagina, _ := strconv.Atoi(r.URL.Query().Get("pagina"))
	if pagina < 1 {
		pagina = 1
	}
	porPagina := 10 // Items por página

	// Abrir la base de datos
	dbPath := filepath.Join("datos", "salida", nombreArchivo+".db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		http.Error(w, "Error al abrir la base de datos", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Consulta con paginación
	offset := (pagina - 1) * porPagina
	query := `
		SELECT 
			r.id, r.autor, r.nombre, r.contenido, r.pagina,
			(SELECT GROUP_CONCAT(e.nombre, ', ') 
			FROM etiquetas e
			JOIN recortes_etiquetas re ON e.id = re.etiqueta_id
			WHERE re.recorte_id = r.id) as etiquetas
		FROM recortes r
		ORDER BY r.id
		LIMIT ? OFFSET ?
	`

	rows, err := db.Query(query, porPagina, offset)
	if err != nil {
		http.Error(w, "Error en consulta: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var recortes []RecorteTabla
	for rows.Next() {
		var r RecorteTabla
		var etiquetasStr sql.NullString

		err := rows.Scan(
			&r.ID, &r.Autor, &r.Nombre, &r.Contenido, &r.Pagina, &etiquetasStr,
		)
		if err != nil {
			http.Error(w, "Error al leer datos: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Procesar etiquetas
		if etiquetasStr.Valid {
			etiquetas := strings.Split(etiquetasStr.String, ", ")
			for _, nombre := range etiquetas {
				if nombre != "" {
					r.Etiquetas = append(r.Etiquetas, modelos.Etiqueta{Nombre: nombre})
				}
			}
		}

		recortes = append(recortes, r)
	}

	// Obtener total de registros para paginación
	var total int
	db.QueryRow("SELECT COUNT(*) FROM recortes").Scan(&total)

	// Renderizar solo el fragmento de tabla
	tmpl, err := template.ParseFiles("templates/tabla.html")
	if err != nil {
		http.Error(w, "Error al parsear template", http.StatusInternalServerError)
		return
	}

	data := TablaData{
		Recortes: recortes,
		Pagina:   pagina,
		Total:    total,
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error al renderizar tabla", http.StatusInternalServerError)
	}
}
