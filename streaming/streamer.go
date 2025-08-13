// Package streaming define la interfaz para todos los tipos de clientes de streaming.
package streaming


import "mi-radio/playlist" // <-- IMPORT CORREGIDO

type StatusUpdater func(message string)
type Streamer interface {
	Stream(p *playlist.Playlist, status StatusUpdater) error // <-- Añadimos el asterisco *
}