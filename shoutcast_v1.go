// Fichero: streaming/shoutcast_v1.go
// VERSIÓN ESTABLE Y FUNCIONAL - SIN METADATOS

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

// Stream conecta y transmite la playlist con una conexión persistente.
func (s *ShoutcastV1Streamer) Stream(p *playlist.Playlist, status StatusUpdater) error {
	status("[yellow]Iniciando AutoDJ (Versión Estable)...")
	songs, err := p.GetSongs()
	if err != nil {
		return err
	}
	status(fmt.Sprintf("[green]Playlist cargada con %d canciones.", len(songs)))

	// Bucle de reconexión infinito
	for {
		status("\n[yellow]Estableciendo conexión persistente...")
		conn, err := s.connectAndAuth(status)

		if err != nil {
			status(fmt.Sprintf("[red]Fallo de conexión inicial: %v", err))
			status("[yellow]Reintentando en 10 segundos...")
			time.Sleep(10 * time.Second)
			continue
		}

		status("[green]¡EMISORA EN EL AIRE! Transmitiendo playlist...")
		err = s.streamPlaylist(conn, songs, status)
		conn.Close()

		status(fmt.Sprintf("\n[red]¡CONEXIÓN PERDIDA! Error: %v", err))
		status("[yellow]Esperando 5 segundos antes de reconectar...")
		time.Sleep(5 * time.Second)
	}
}

// connectAndAuth se conecta y autentica. SIN cabeceras de metadatos.
func (s *ShoutcastV1Streamer) connectAndAuth(status StatusUpdater) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.Config.Hostname, s.Config.Port), 10*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(conn, "%s\r\n", s.Config.Password)
	if err != nil {
		conn.Close()
		return nil, err
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || !strings.HasPrefix(response, "OK2") {
		conn.Close()
		return nil, fmt.Errorf("contraseña rechazada (Respuesta: %s)", strings.TrimSpace(response))
	}
	status("[green]¡Contraseña aceptada! (Recibido OK2)")

	// Enviamos las cabeceras básicas, SIN "Icy-MetaData:1"
	headers := "icy-name:Mi Radio en Go (Estable)\r\n" + "icy-genre:Varios\r\n" + "icy-pub:1\r\n" + "\r\n"
	_, err = conn.Write([]byte(headers))
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// streamPlaylist gestiona el bucle de la playlist con una conexión simple y estable.
func (s *ShoutcastV1Streamer) streamPlaylist(conn net.Conn, songs []string, status StatusUpdater) error {
	for { // Bucle infinito para que la playlist se repita
		for _, songPath := range songs {
			cleanSongName := strings.TrimSuffix(filepath.Base(songPath), filepath.Ext(songPath))
			status(fmt.Sprintf("--> Próxima canción: [cyan]%s", cleanSongName))

			file, err := os.Open(songPath)
			if err != nil {
				status(fmt.Sprintf("[red]No se pudo abrir: %v", err))
				continue
			}

			// Usamos el simple y robusto io.Copy. Esto es lo que nos daba estabilidad.
			if _, err := io.Copy(conn, file); err != nil {
				file.Close()
				// Este error significa que la conexión se rompió.
				return fmt.Errorf("error durante la transmisión: %w", err)
			}

			file.Close()
			status(fmt.Sprintf("   '%s' [green]enviado.", cleanSongName))
		}
		status("[yellow]Fin de la lista. Reiniciando playlist...")
	}
}
