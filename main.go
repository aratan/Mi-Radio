// Fichero: main.go

package main

import (
	"fmt"
	"log"
	"mi-radio/config"
	"mi-radio/playlist"
	"mi-radio/streaming"
)

func main() {
	// 1. Cargar Configuración (Responsabilidad del paquete 'config')
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Error fatal cargando config.json: %v", err)
	}

	// 2. Crear el Streamer (Depende de la interfaz 'Streamer')
	// Si mañana creamos un streamer para Icecast, solo cambiaríamos esta línea.
	streamer := &streaming.ShoutcastV1Streamer{Config: cfg}

	// 3. Crear la UI y definir qué hacer cuando se "conecta"
	// La función que pasamos a NewTUI es un "callback". Se ejecutará cuando
	// la UI esté lista para iniciar la transmisión.
	var appUI *TUI
	appUI = NewTUI(func() {
		// --- Lógica que se ejecuta al iniciar la transmisión ---

		// 3a. Crear la Playlist, ahora leyendo un fichero .m3u
		// (Responsabilidad del paquete 'playlist')
		// Para hacerlo más configurable, esta ruta podría venir del config.json
		p := &playlist.Playlist{FilePath: "./music/playlist.m3u"}

		// 3b. Iniciar el streaming en una goroutine para no bloquear la UI.
		// Pasamos la playlist y el método para actualizar el estado de la UI.
		// ¡Esto es Inversión de Dependencias en acción!
		go func() {
			err := streamer.Stream(p, appUI.UpdateStatus)
			if err != nil {
				// Si el streamer devuelve un error fatal, lo mostramos en el dashboard.
				appUI.UpdateStatus(fmt.Sprintf("[red]ERROR FATAL DEL STREAMER: %v", err))
			}
		}()
	})

	// 4. Ejecutar la aplicación
	log.Println("Iniciando aplicación de radio...")
	if err := appUI.Run(); err != nil {
		log.Fatalf("Error al ejecutar la TUI: %v", err)
	}
}