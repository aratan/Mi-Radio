// Fichero: streaming/shoutcast_v1.go
// VERSIÓN FINAL Y ROBUSTA - CON FLUJO CONTINUO (io.MultiReader)

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

// Stream conecta y transmite la playlist.
func (s *ShoutcastV1Streamer) Stream(p *playlist.Playlist, status StatusUpdater) error {
	status("[yellow]Iniciando AutoDJ con Flujo Continuo (Estable)...")
	songs, err := p.GetSongs()
	if err != nil {
		return err
	}
	status(fmt.Sprintf("[green]Playlist cargada con %d canciones.", len(songs)))

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
		err = s.streamPlaylist(conn, songs, status) // Pasamos a la nueva función
		conn.Close()

		status(fmt.Sprintf("\n[red]¡CONEXIÓN PERDIDA! Error: %v", err))
		status("[yellow]Esperando 5 segundos antes de reconectar...")
		time.Sleep(5 * time.Second)
	}
}

// connectAndAuth se conecta y autentica (versión estable sin metadatos).
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

// streamPlaylist ahora usa io.MultiReader para un flujo ininterrumpido.
func (s *ShoutcastV1Streamer) streamPlaylist(conn net.Conn, songs []string, status StatusUpdater) error {
	// Bucle infinito para que la playlist se repita
	for {
		// 1. Abrimos TODOS los ficheros de la playlist y preparamos los "lectores".
		readers := make([]io.Reader, 0, len(songs))
		closers := make([]io.Closer, 0, len(songs))

		for _, songPath := range songs {
			file, err := os.Open(songPath)
			if err != nil {
				status(fmt.Sprintf("[red]Saltando fichero (no se pudo abrir): %s", filepath.Base(songPath)))
				continue
			}
			readers = append(readers, file)
			closers = append(closers, file)
		}

		// 2. Creamos el Super-Stream con io.MultiReader.
		//    Este es el lector virtual que contiene todas las canciones, una detrás de otra.
		playlistStream := io.MultiReader(readers...)

		// 3. Usamos io.Copy para transmitir el Super-Stream completo de una sola vez.
		//    Esto es bloqueante y solo devolverá un error si la conexión se cae.
		status(fmt.Sprintf("--> Transmitiendo playlist completa (%d canciones)...", len(readers)))
		if _, err := io.Copy(conn, playlistStream); err != nil {
			// Cerramos todos los ficheros antes de devolver el error
			for _, c := range closers {
				c.Close()
			}
			return fmt.Errorf("error durante la transmisión del flujo continuo: %w", err)
		}

		// Cerramos todos los ficheros al final de la playlist
		for _, c := range closers {
			c.Close()
		}

		status("[yellow]Fin de la lista. Reiniciando playlist...")
	}
}