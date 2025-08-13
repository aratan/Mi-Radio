// Fichero: streaming/shoutcast_v1.go
// VERSIÓN CON MEMORIA: PASA A LA SIGUIENTE CANCIÓN TRAS RECONEXIÓN

package streaming

import (
	"bufio"
	"fmt"
	"io"
	"mi-radio/config"
	"mi-radio/playlist"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ShoutcastV1Streamer implementa la interfaz Streamer para el protocolo Shoutcast v1.
type ShoutcastV1Streamer struct {
	Config config.Config
}

// Stream ahora gestiona el estado de la playlist (la canción actual).
func (s *ShoutcastV1Streamer) Stream(p *playlist.Playlist, status StatusUpdater) error {
	status("[yellow]Iniciando AutoDJ con Memoria de Pista...")
	songs, err := p.GetSongs()
	if err != nil {
		return err
	}
	status(fmt.Sprintf("[green]Playlist cargada con %d canciones.", len(songs)))

	// El índice de la canción actual. Empieza en 0.
	currentSongIndex := 0

	for { // Bucle de reconexión infinito
		status("\n[yellow]Estableciendo conexión...")
		conn, err := s.connectAndAuth(status)
		if err != nil {
			status(fmt.Sprintf("[red]Fallo de conexión inicial: %v", err))
			status("[yellow]Reintentando en 10 segundos...")
			time.Sleep(10 * time.Second)
			continue // Vuelve al inicio del bucle para reintentar la conexión
		}

		status("[green]¡EMISORA EN EL AIRE! Transmitiendo playlist...")
		
		// --- ¡LA LÓGICA CLAVE! ---
		// Le pasamos a streamPlaylist la lista de canciones A PARTIR de la actual.
		// El resultado 'nextIndex' nos dirá dónde continuar si todo va bien.
		nextIndex, err := s.streamPlaylist(conn, songs, currentSongIndex, status)
		conn.Close() // Siempre cerramos la conexión al terminar o fallar

		// Actualizamos el índice para la próxima iteración.
		currentSongIndex = nextIndex

		// Si se completó la playlist (nextIndex llegó al final), la reiniciamos.
		if currentSongIndex >= len(songs) {
			status("[yellow]Playlist completada. Reiniciando desde el principio.")
			currentSongIndex = 0
		}

		if err != nil {
			status(fmt.Sprintf("\n[red]¡CONEXIÓN PERDIDA! Error: %v", err))
			// YA ESTAMOS EN LA SIGUIENTE CANCIÓN, así que solo esperamos.
			status("[yellow]Esperando 5 segundos antes de reconectar...")
			time.Sleep(5 * time.Second)
		}
	}
}

// connectAndAuth no cambia.
func (s *ShoutcastV1Streamer) connectAndAuth(status StatusUpdater) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.Config.Hostname, s.Config.Port), 10*time.Second)
	if err != nil { return nil, err }
	_, err = fmt.Fprintf(conn, "%s\r\n", s.Config.Password)
	if err != nil { conn.Close(); return nil, err }
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || !strings.HasPrefix(response, "OK2") {
		conn.Close()
		return nil, fmt.Errorf("contraseña rechazada (Respuesta: %s)", strings.TrimSpace(response))
	}
	status("[green]¡Contraseña aceptada! (Recibido OK2)")
	headers := "icy-name:Mi Radio en Go (Estable)\r\n" + "icy-genre:Varios\r\n" + "icy-pub:1\r\n" + "\r\n"
	_, err = conn.Write([]byte(headers))
	if err != nil { conn.Close(); return nil, err }
	return conn, nil
}

// streamPlaylist ahora reproduce la lista UNA SOLA VEZ, empezando desde un índice dado.
// Devuelve el índice de la PRÓXIMA canción a reproducir y un error si la conexión se cae.
func (s *ShoutcastV1Streamer) streamPlaylist(conn net.Conn, songs []string, startIndex int, status StatusUpdater) (int, error) {
	// Bucle que empieza desde la canción que nos toca
	for i := startIndex; i < len(songs); i++ {
		songPath := songs[i]
		cleanSongName := strings.TrimSuffix(filepath.Base(songPath), filepath.Ext(songPath))
		status(fmt.Sprintf("--> Próxima canción (%d/%d): [cyan]%s", i+1, len(songs), cleanSongName))

		file, err := os.Open(songPath)
		if err != nil {
			status(fmt.Sprintf("[red]No se pudo abrir: %v", err))
			continue // Salta a la siguiente canción de la lista
		}

		if _, err := io.Copy(conn, file); err != nil {
			file.Close()
			// La conexión se cayó. Devolvemos el índice de la SIGUIENTE canción (i + 1)
			// y el error para que el bucle principal sepa que debe reconectar.
			return i + 1, fmt.Errorf("error durante la transmisión: %w", err)
		}

		file.Close()
		status(fmt.Sprintf("   '%s' [green]enviado.", cleanSongName))
	}

	// Si el bucle termina sin errores, significa que hemos completado la playlist.
	// Devolvemos un índice que está fuera de los límites de la lista
	// para que el bucle principal sepa que debe reiniciar.
	return len(songs), nil
}