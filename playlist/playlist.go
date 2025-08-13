package playlist // <-- OJO: El paquete ahora es "playlist"

import (
	"fmt"
	"path/filepath"
)

// Playlist gestiona la lista de canciones.
type Playlist struct {
	MusicDir string
}

// GetSongs busca y devuelve la lista de ficheros .mp3 en el directorio.
func (p *Playlist) GetSongs() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(p.MusicDir, "*.mp3"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no se encontraron ficheros .mp3 en '%s'", p.MusicDir)
	}
	return files, nil
}