// Fichero: playlist/playlist.go

package playlist

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath" // <-- Importamos el paquete para manejar rutas de ficheros
	"strings"
)

// Playlist ahora se refiere a un fichero de playlist, no a un directorio.
type Playlist struct {
	FilePath string
}

// GetSongs ahora lee el fichero M3U y construye las rutas completas.
func (p *Playlist) GetSongs() ([]string, error) {
	file, err := os.Open(p.FilePath)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el fichero de playlist '%s': %w", p.FilePath, err)
	}
	defer file.Close()

	var songs []string
	
	// --- ¡LA LÓGICA INTELIGENTE! ---
	// 1. Obtenemos la ruta del directorio que contiene el fichero de playlist.
	playlistDir := filepath.Dir(p.FilePath)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Ignoramos líneas vacías o comentarios (que suelen empezar con #)
		if line != "" && !strings.HasPrefix(line, "#") {
			// 2. Unimos la ruta del directorio con el nombre del fichero de la línea.
			// filepath.Join se encarga de poner las barras / o \ correctamente.
			fullPath := filepath.Join(playlistDir, line)
			
			// 3. Añadimos la ruta completa y correcta a nuestra lista de canciones.
			songs = append(songs, fullPath)
		}
	}
	// --- FIN DE LA LÓGICA INTELIGENTE ---

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error leyendo el fichero de playlist: %w", err)
	}

	if len(songs) == 0 {
		return nil, fmt.Errorf("no se encontraron canciones válidas en la playlist '%s'", p.FilePath)
	}

	return songs, nil
}