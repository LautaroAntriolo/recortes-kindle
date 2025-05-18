package rutas

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "recortes.db"

func initDB(nombreArchivo string) (*sql.DB, error) {
	// Eliminar la extensión .txt y asegurar nombre válido
	nombreSinExt := strings.TrimSuffix(nombreArchivo, filepath.Ext(nombreArchivo))
	nombreDB := fmt.Sprintf("%s.db", nombreSinExt) // Ej: "My Clippings.db"

	dbPath := filepath.Join("datos", "salida", nombreDB)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir la base de datos: %v", err)
	}

	// Crear tablas
	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS recortes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			autor TEXT,
			nombre TEXT,
			contenido TEXT,
			pagina INTEGER,
			visibilidad BOOLEAN DEFAULT 1,
			fecha TEXT,
			nombreRecorte TEXT,
			hora TEXT,
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

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("error al crear tabla: %v", err)
		}
	}
	return nil
}
