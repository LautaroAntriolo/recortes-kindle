package lectura

import (
    "bufio"
    "os"
    "strings"
)

// ReadFile lee el archivo y devuelve una lista de líneas.
func LeerArchivo(filePath string) ([]string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line != "" { // Ignorar líneas vacías
            lines = append(lines, line)
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return lines, nil
}